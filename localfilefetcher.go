package main

import (
	"os"
	"path"
	"path/filepath"
	"strings"
)

var _ FileFetcher = (*localFileFetcher)(nil)

type localFileFetcher struct {
	localDirectory string
}

// NewLocalFileFetcher returns a FileFetcher that serves a local directory.
func NewLocalFileFetcher(localPath string) FileFetcher {
	newFileFetcher := localFileFetcher{
		localDirectory: localPath,
	}
	return newFileFetcher
}

func (fetcher localFileFetcher) GetFile(path string) (*FetchedFile, error) {
	fullPath := filepath.Join(fetcher.localDirectory, path)
	f, err := os.Open(fullPath)
	if err != nil {
		return nil, err
	}
	file := &FetchedFile{
		Payload:    f,
		BucketName: fetcher.localDirectory,
		Key:        path,
		Etag:       fullPath,
	}
	return file, nil
}

func (fetcher localFileFetcher) ListDir(prefix string) ([]ListDirEntry, error) {
	searchDir := searchDir{
		rootPath: fetcher.localDirectory,
		prefix:   path.Join(fetcher.localDirectory, prefix),
	}
	if err := filepath.Walk(fetcher.localDirectory, searchDir.processFile); err != nil {
		return nil, err
	}
	return searchDir.fList, nil
}

type searchDir struct {
	rootPath string
	prefix   string
	fList    []ListDirEntry
}

func (searchDir *searchDir) processFile(filePath string, info os.FileInfo, err error) error {

	if filePath == searchDir.rootPath {
		return nil
	}
	if info == nil {
		return nil
	}
	if info.IsDir() {
		return nil
	}
	if !strings.HasPrefix(filePath, searchDir.prefix) {
		return nil
	}
	entry := ListDirEntry{
		Name:         info.Name(),
		LastModified: info.ModTime(),
		SizeKb:       info.Size() / 1024,
	}
	searchDir.fList = append(searchDir.fList, entry)
	return nil
}
