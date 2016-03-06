package main

import (
	"os"
	"path/filepath"
)

var _ FileFetcher = (*localFileFetcher)(nil)

type localFileFetcher struct {
	localDirectory string
}

func NewLocalFileFetcher(localPath string) FileFetcher {
	newFileFetcher := localFileFetcher{
		localDirectory: localPath,
	}
	return newFileFetcher
}

func (this localFileFetcher) GetFile(path string) (*FetchedFile, error) {
	fullPath := filepath.Join(this.localDirectory, path)
	f, err := os.Open(fullPath)
	if err != nil {
		return nil, err
	}
	file := &FetchedFile{
		Payload:    f,
		BucketName: this.localDirectory,
		Key:        path,
		Etag:       fullPath,
	}
	return file, nil
}

func (this localFileFetcher) ListDir(path string) ([]ListDirEntry, error) {
	fullPath := filepath.Join(this.localDirectory, path)

	var fileList fileList
	if err := filepath.Walk(fullPath, fileList.processFile); err != nil {
		return nil, err
	}
	return fileList.fList, nil
}

type fileList struct {
	fList []ListDirEntry
}

func (list *fileList) processFile(filePath string, info os.FileInfo, err error) error {
	if info.IsDir() {
		return filepath.SkipDir
	}
	entry := ListDirEntry{
		Name:         info.Name(),
		LastModified: info.ModTime(),
		SizeKb:       info.Size() / 1024,
	}
	list.fList = append(list.fList, entry)
	return nil
}
