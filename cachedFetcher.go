package main

import (
	"log"
	"time"
)

// CacheConfigs is the configuration for the cached fetcher
type CacheConfigs struct {
	ExpireSecs  int64
	realFetcher FileFetcher
	// this can grow unbounded. entries are purged only on ListBucket(key) calls
	cachedListBucket  map[string]CachedDirListResult
	cacheUnixTimeToGC int64
}

type CachedDirListResult struct {
	result             []DirListEntry
	expirationUnixTime int64
}

var config CacheConfigs

const intervalSecsRunGC = 600

func NewCachedFetcher(configs CacheConfigs, fileFetcher FileFetcher) FileFetcher {
	config = configs
	config.realFetcher = fileFetcher
	config.cachedListBucket = make(map[string]CachedDirListResult)
	config.cacheUnixTimeToGC = time.Now().Unix() + intervalSecsRunGC

	return config
}

// GetFile is pass-through
func (config CacheConfigs) GetFile(key string) (*S3File, error) {
	return config.realFetcher.GetFile(key)
}

// ListBucket caches results from realFetcher and also garbage collects the cache when required.
func (config CacheConfigs) ListBucket(path string) ([]DirListEntry, error) {
	now := time.Now().Unix()

	result := config.getFromCache(path, now)
	if result != nil {
		// cache hit
		return result, nil
	}

	realResult, err := config.realFetcher.ListBucket(path)
	if err != nil {
		return realResult, err
	}

	config.cachedListBucket[path] = CachedDirListResult{
		result:             realResult,
		expirationUnixTime: now + config.ExpireSecs,
	}

	return realResult, nil
}

func (config CacheConfigs) getFromCache(key string, unixTime int64) []DirListEntry {
	cachedResult, cacheHit := config.cachedListBucket[key]
	if cacheHit && unixTime < cachedResult.expirationUnixTime {
		return cachedResult.result
	} else if cacheHit {
		log.Printf("cache: deleting entry[%v] (staleSecs: %v) ",
			key, unixTime-cachedResult.expirationUnixTime)
		delete(config.cachedListBucket, key)
	}

	if config.shouldRunGC(unixTime) {
		config.garbageCollectCache()
	}
	return nil
}

func (config CacheConfigs) shouldRunGC(unixTime int64) bool {
	return config.cacheUnixTimeToGC < unixTime
}

// TODO: this is O(N)
func (config CacheConfigs) garbageCollectCache() {
	now := time.Now().Unix()
	for path, cachedResult := range config.cachedListBucket {
		staleSecs := now - cachedResult.expirationUnixTime
		if staleSecs > 0 {
			log.Printf("GCcache: deleting entry[%v] (staleSecs: %v)", path, staleSecs)
			delete(config.cachedListBucket, path)
		}
	}
	config.cacheUnixTimeToGC = now + intervalSecsRunGC
}
