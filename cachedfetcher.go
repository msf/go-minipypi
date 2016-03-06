package main

import (
	"log"
	"sync"
	"time"
)

const intervalSecsRunGC = 600

// CacheConfigs is the configuration for the cached fetcher
type CacheConfigs struct {
	ExpireSecs  int64
	realFetcher FileFetcher
	// this can grow unbounded. entries are purged only on ListBucket(key) calls
	cachedListBucket  map[string]CachedListBucketResult
	cacheLock         sync.Mutex
	clock             Clock
	cacheUnixTimeToGC int64
}

// CachedListBucketResult holds a result plus its expiration time.
type CachedListBucketResult struct {
	result             []DirListEntry
	expirationUnixTime int64
}

// NewCachedFetcher returns a FileFetcher that caches results from calls to ListBucket.
// cache is in-memory and very simplistic, garbage collection is a O(N) for N cache entries.
func NewCachedFetcher(configs CacheConfigs, fileFetcher FileFetcher) FileFetcher {
	configs.realFetcher = fileFetcher
	configs.clock = &realClock{}
	configs.cachedListBucket = make(map[string]CachedListBucketResult)
	configs.cacheUnixTimeToGC = configs.clock.Now().Unix() + intervalSecsRunGC
	configs.cacheLock = sync.Mutex{}

	return configs
}

// SetClock changes the timekeeping instance to use. this will also reset the cache.
func (this CacheConfigs) SetClock(newClock Clock) {
	this.clock = newClock
	this.cacheLock.Lock()
	this.cachedListBucket = make(map[string]CachedListBucketResult)
	this.cacheUnixTimeToGC = this.clock.Now().Unix() + intervalSecsRunGC
	this.cacheLock.Unlock()
}

// GetFile is pass-through, no caching is done.
func (this CacheConfigs) GetFile(key string) (*S3File, error) {
	return this.realFetcher.GetFile(key)
}

// ListBucket caches results from realFetcher and also garbage collects the cache when required.
func (this CacheConfigs) ListBucket(path string) ([]DirListEntry, error) {
	now := time.Now().Unix()

	result := this.getFromCache(path, now)
	if result != nil {
		// cache hit
		return result, nil
	}

	realResult, err := this.realFetcher.ListBucket(path)
	if err != nil {
		return realResult, err
	}

	this.addCacheEntry(path, CachedListBucketResult{
		result:             realResult,
		expirationUnixTime: now + this.ExpireSecs,
	})

	return realResult, nil
}

func (this CacheConfigs) getFromCache(key string, unixTime int64) []DirListEntry {
	cachedResult, cacheHit := this.cachedListBucket[key]
	if cacheHit && unixTime < cachedResult.expirationUnixTime {
		return cachedResult.result
	} else if cacheHit {
		log.Printf("cache: deleting entry[%v] (staleSecs: %v) ",
			key, unixTime-cachedResult.expirationUnixTime)
		this.delCacheEntry(key)
	}

	if this.shouldRunGC(unixTime) {
		this.garbageCollectCache()
	}
	return nil
}

func (this CacheConfigs) shouldRunGC(unixTime int64) bool {
	return this.cacheUnixTimeToGC < unixTime
}

// TODO: this is O(N)
func (this CacheConfigs) garbageCollectCache() {
	now := time.Now().Unix()
	for path, cachedResult := range this.cachedListBucket {
		staleSecs := now - cachedResult.expirationUnixTime
		if staleSecs > 0 {
			log.Printf("GCcache: deleting entry[%v] (staleSecs: %v)", path, staleSecs)
			this.delCacheEntry(path)
		}
	}
	this.cacheUnixTimeToGC = now + intervalSecsRunGC
}

func (this CacheConfigs) addCacheEntry(key string, result CachedListBucketResult) {
	this.cacheLock.Lock()
	this.cachedListBucket[key] = result
	this.cacheLock.Unlock()
}

func (this CacheConfigs) delCacheEntry(key string) {
	this.cacheLock.Lock()
	delete(this.cachedListBucket, key)
	this.cacheLock.Unlock()
}
