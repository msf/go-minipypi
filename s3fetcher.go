package minipi

import "log"

type S3configs struct {
	AccessKey  string
	SecretKey  string
	BucketName string
}

type S3File struct {
	Payload    []byte
	Etag       string
	Name       string
	BucketName string
}

type S3fetcher interface {
	GetFile(path string) *S3File
	ListBucket(bucketName string) []string
}

var configs S3configs

func S3Fetcher(cfg S3configs) S3fetcher {
	return cfg
}

func (s3cfg S3configs) GetFile(path string) *S3File {
	log.Print("GetFile:%v", path)
	f := &S3File{
		BucketName: s3cfg.BucketName,
		Etag:       "yayaya",
		Name:       path,
		Payload:    []byte(path),
	}
	return f
}

func (s3cfg S3configs) ListBucket(bucket string) []string {
	return make([]string, 1)
}
