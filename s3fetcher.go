package main

import (
	"io/ioutil"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// S3configs holds basic config data required by S3fetcher
type S3configs struct {
	BucketName      string
	CredentialsFile string
	Region          string
	s3svc           *s3.S3
}

// S3File contains the payload of an S3File (or key) and other important fields.
type S3File struct {
	Payload    []byte
	Etag       string
	Key        string
	BucketName string
}

// NewS3Fetcher is a S3 backed implementation of the FileFetcher interface.
// it does the setup of the S3 service session state required to implement FileFetcher interface
func NewS3Fetcher(cfg S3configs) FileFetcher {
	svc := s3.New(session.New(
		aws.NewConfig().WithRegion(cfg.Region).WithCredentials(
			credentials.NewSharedCredentials(cfg.CredentialsFile, "default"))))
	cfg.s3svc = svc

	return cfg
}

// GetFile from S3 bucket identified by key
func (s3cfg S3configs) GetFile(key string) (*S3File, error) {
	res, err := s3cfg.s3svc.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(s3cfg.BucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	ret := &S3File{Payload: data, Etag: *res.ETag, Key: key, BucketName: s3cfg.BucketName}

	return ret, nil
}

// ListBucket returns the contents of our preconfigured bucket that start with path
func (s3cfg S3configs) ListBucket(path string) ([]DirListEntry, error) {

	params := &s3.ListObjectsInput{
		Bucket: aws.String(s3cfg.BucketName),
		Prefix: aws.String(path),
	}
	items := &s3KeyList{keyList: make([]DirListEntry, 0)}

	if err := s3cfg.s3svc.ListObjectsPages(params, items.processPage); err != nil {
		return nil, err
	}

	return items.keyList, nil
}

type s3KeyList struct {
	keyList []DirListEntry
}

func (list *s3KeyList) processPage(page *s3.ListObjectsOutput, more bool) bool {

	for _, obj := range page.Contents {
		entry := DirListEntry{
			Name:         *obj.Key,
			LastModified: obj.LastModified.Format(time.RFC3339),
			SizeKb:       *obj.Size / 1024,
		}
		list.keyList = append(list.keyList, entry)
	}

	return true
}
