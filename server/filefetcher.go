package main

import (
	"io"
	"time"
)

// FetchedFile contains the Payload of an S3File (or key) and other important fields.
type FetchedFile struct {
	Payload    io.ReadCloser
	Etag       string
	Key        string
	BucketName string
}

type ListDirEntry struct {
	Name         string
	LastModified time.Time
	SizeKb       int64
}

// FileFetcher interface that the frontend uses to serve requests.
type FileFetcher interface {
	GetFile(path string) (*FetchedFile, error)
	ListDir(path string) ([]ListDirEntry, error)
}
