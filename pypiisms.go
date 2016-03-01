package main

import (
	"log"
	"strings"
)

func normalizeFileName(filePath string) string {
	return strings.Replace(strings.ToLower(filePath), "-", "_", -1)
}

// handlePypiDirlist works around the fact that dirlistings are case insensitive so direct search for the path might fail
func handlePypiDirList(fetcher FileFetcher, path string) ([]DirListEntry, error) {

	prefix := strings.TrimPrefix(path, "/")  // remove initial /
	prefix = strings.TrimSuffix(prefix, "/") // and last one
	fileList, err := fetcher.ListBucket(prefix)
	if err != nil {
		return fileList, err
	}
	if len(fileList) > 0 {
		log.Println("PypiDirList vanilla HIT")
		return fileList, nil
	}

	// lets try searching for normalized name
	fileList, err = fetcher.ListBucket(normalizeFileName(prefix))
	if err != nil {
		return fileList, err
	}
	if len(fileList) > 0 {
		log.Println("PypiDirList normalizeFileName HIT")
		return fileList, nil
	}

	// no results, lets try normalizing bucket entry names
	allFiles, err := fetcher.ListBucket("/")
	if err != nil {
		return allFiles, err
	}

	results := make([]DirListEntry, 0)
	normalizedPath := normalizeFileName(path)
	for _, entry := range allFiles {
		fileName := normalizeFileName(entry.Name)
		if strings.HasPrefix(fileName, normalizedPath) {
			results = append(results, entry)
		}
	}
	if len(results) > 0 {
		log.Println("PypiDirList normalizedBucketEntries HIT")
	}
	return results, nil
}
