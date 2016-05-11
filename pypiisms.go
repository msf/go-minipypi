package main

import "strings"

func normalizeFileName(filePath string) string {
	return strings.Replace(strings.ToLower(filePath), "-", "_", -1)
}

func handlePypiFileNames(key string) string {
	pathParts := strings.Split(key, "/")
	nparts := len(pathParts)
	if nparts > 2 {
		// requests can come in the form: /packagename/packagename-version-blblablabla.xx
		// the real key in these cases is: /packagename-version-blablalbalbla.xx
		key = "/" + pathParts[nparts-1]
	}
	return key
}

// handlePypiDirlist works around the fact that dirlistings are case insensitive so direct search for the path might fail
func handlePypiListDir(fetcher FileFetcher, path string) ([]ListDirEntry, error) {

	prefix := strings.TrimPrefix(path, "/")  // remove initial /
	prefix = strings.TrimSuffix(prefix, "/") // and last one

	// case-insensitive search, search for X* + x*
	firstLetter := prefix[0:1]
	lowerFiles, err := fetcher.ListDir(strings.ToLower(firstLetter))
	if err != nil {
		return lowerFiles, err
	}
	upperFiles, err := fetcher.ListDir(strings.ToUpper(firstLetter))
	if err != nil {
		return upperFiles, err
	}

	// now merge both and filter by normalized prefix comparison.
	allFiles := append(lowerFiles, upperFiles...)
	normalizedPrefix := normalizeFileName(prefix)
	var results []ListDirEntry
	for _, entry := range allFiles {
		fileName := normalizeFileName(entry.Name)
		if strings.HasPrefix(fileName, normalizedPrefix) {
			results = append(results, entry)
		}
	}
	return results, nil
}
