package main

import "strings"

func normalizeFileName(filePath string) string {
	return strings.Replace(strings.ToLower(filePath), "-", "_", -1)
}

// handlePypiDirlist works around the fact that dirlistings are case insensitive so direct search for the path might fail
func handlePypiDirList(fetcher FileFetcher, path string) ([]DirListEntry, error) {

	prefix := strings.TrimPrefix(path, "/")  // remove initial /
	prefix = strings.TrimSuffix(prefix, "/") // and last one

	// case-insensitive search, search for X* + x*
	firstLetter := prefix[0:1]
	lowerFiles, err := fetcher.ListBucket(strings.ToLower(firstLetter))
	if err != nil {
		return lowerFiles, err
	}
	upperFiles, err := fetcher.ListBucket(strings.ToUpper(firstLetter))
	if err != nil {
		return upperFiles, err
	}

	// now merge both and filter by normalized prefix comparison.
	allFiles := append(lowerFiles, upperFiles...)
	normalizedPrefix := normalizeFileName(prefix)
	results := make([]DirListEntry, 0)
	for _, entry := range allFiles {
		fileName := normalizeFileName(entry.Name)
		if strings.HasPrefix(fileName, normalizedPrefix) {
			results = append(results, entry)
		}
	}
	return results, nil
}
