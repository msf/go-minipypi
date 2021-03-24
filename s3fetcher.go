package main

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

var _ FileFetcher = (*S3configs)(nil)

// S3configs holds basic config data required by S3fetcher
type S3configs struct {
	BucketName      string
	CredentialsFile string
	Region          string
	s3svc           *s3.S3
}

// NewS3Fetcher is a S3 backed implementation of the FileFetcher interface.
// it does the setup of the S3 service session state required to implement FileFetcher interface
func NewS3Fetcher(cfg S3configs) FileFetcher {
	awsCfg := aws.NewConfig().WithRegion(cfg.Region)
	if cfg.CredentialsFile != "" {
		log.Printf("using AWS credentials from %s", cfg.CredentialsFile)
		awsCfg = awsCfg.WithCredentials(credentials.NewSharedCredentials(cfg.CredentialsFile, "default"))
	} else {
		log.Print("no AWS credentials file specified, using the default credentials chain")
	}

	svc := s3.New(session.New(awsCfg))
	cfg.s3svc = svc

	return cfg
}

// GetFile from S3 bucket identified by key
func (s3cfg S3configs) GetFile(key string) (*FetchedFile, error) {
	res, err := s3cfg.s3svc.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(s3cfg.BucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}

	ret := &FetchedFile{Payload: res.Body, Etag: *res.ETag, Key: key, BucketName: s3cfg.BucketName}

	return ret, nil
}

// ListDir returns the contents of our preconfigured bucket that start with path
func (s3cfg S3configs) ListDir(path string) ([]ListDirEntry, error) {

	params := &s3.ListObjectsInput{
		Bucket: aws.String(s3cfg.BucketName),
		Prefix: aws.String(path),
	}

	var items s3KeyList
	if err := s3cfg.s3svc.ListObjectsPages(params, items.processPage); err != nil {
		return nil, err
	}

	return items.keyList, nil
}

type s3KeyList struct {
	keyList []ListDirEntry
}

func (list *s3KeyList) processPage(page *s3.ListObjectsOutput, more bool) bool {
	for _, obj := range page.Contents {
		entry := ListDirEntry{
			Name:         *obj.Key,
			LastModified: *obj.LastModified,
			SizeKb:       *obj.Size / 1024,
		}
		list.keyList = append(list.keyList, entry)
	}
	return true
}
