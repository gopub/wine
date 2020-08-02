package storage

import (
	"bytes"
	"context"
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

const (
	cacheControl = "public, max-age=14400"
)

type S3ACL string

// S3Bucket ACL
const (
	S3Private                S3ACL = "private"
	S3PublicRead             S3ACL = "public-read"
	S3PublicReadWrite        S3ACL = "public-read-write"
	S3AWSExecRead            S3ACL = "aws-exec-read"
	S3AuthenticatedRead      S3ACL = "authenticated-read"
	S3BucketOwnerRead        S3ACL = "bucket-owner-read"
	S3BucketOwnerFullControl S3ACL = "bucket-owner-full-control"
)

type S3Bucket struct {
	name         string
	uploader     *s3manager.Uploader
	ACL          *string
	CacheControl string
}

func NewS3(name string) *S3Bucket {
	return &S3Bucket{
		name:         name,
		uploader:     s3manager.NewUploader(session.Must(session.NewSession())),
		ACL:          aws.String(string(S3PublicRead)),
		CacheControl: cacheControl,
	}
}

func (s *S3Bucket) Write(ctx context.Context, o *Object) (string, error) {
	input := &s3manager.UploadInput{
		Bucket:       aws.String(s.name),
		Key:          aws.String(o.Name),
		Body:         bytes.NewBuffer(o.Content),
		ACL:          s.ACL,
		CacheControl: aws.String(s.CacheControl),
	}

	if o.Type != "" {
		input.ContentType = aws.String(o.Type)
	}

	resultOrErr := make(chan interface{}, 1)
	go func() {
		defer close(resultOrErr)
		result, err := s.uploader.Upload(input)
		if err != nil {
			resultOrErr <- err
		} else {
			resultOrErr <- result
		}
	}()
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case v := <-resultOrErr:
		if res, ok := v.(*s3manager.UploadOutput); ok {
			return res.Location, nil
		} else if err, ok := v.(error); ok {
			return "", err
		} else {
			return "", errors.New("unknown error")
		}
	}
}
