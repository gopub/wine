package storage

import (
	"bytes"
	"context"
	"fmt"
	"net/url"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/gopub/wine"
)

type AliBucket struct {
	bucket       *oss.Bucket
	ACL          oss.ACLType
	CacheControl string
	baseURL      string
}

func NewAliBucket(endpoint, keyID, keySecret, name string) *AliBucket {
	if endpoint == "" {
		panic("storage: missing endpoint")
	}
	if keyID == "" {
		panic("storage: missing keyID")
	}
	if keySecret == "" {
		panic("storage: missing keySecret")
	}
	if name == "" {
		panic("storage: missing name")
	}
	u, err := url.Parse(endpoint)
	if err != nil {
		panic("storage: invalid endpoint: " + endpoint)
	}
	u.Host = fmt.Sprintf("%s.%s", name, u.Host)
	client, err := oss.New(endpoint, keyID, keySecret)
	if err != nil {
		panic(fmt.Sprintf("Cannot create oss client: %v", err))
	}
	b, err := client.Bucket(name)
	if err != nil {
		panic("storage: cannot create bucket: " + err.Error())
	}
	return &AliBucket{
		bucket:       b,
		ACL:          oss.ACLPublicRead,
		CacheControl: cacheControl,
		baseURL:      u.String(),
	}
}

func (b *AliBucket) Name() string {
	return b.bucket.BucketName
}

func (b *AliBucket) Write(ctx context.Context, obj *Object) (string, error) {
	options := []oss.Option{oss.ACL(b.ACL), oss.CacheControl(b.CacheControl)}
	if obj.Type != "" {
		options = append(options, oss.ContentType(obj.Type))
	}

	errC := make(chan error, 1)
	go func() {
		defer close(errC)
		if err := b.bucket.PutObject(obj.Name, bytes.NewBuffer(obj.Content), options...); err != nil {
			errC <- fmt.Errorf("put object: %w", err)
			return
		}

		// TODO: ACL options above doesn't work, set ACL again
		err := b.bucket.SetObjectACL(obj.Name, oss.ACLPublicRead)
		if err != nil {
			errC <- fmt.Errorf("set object ACL: %w", err)
			return
		}
		errC <- nil
	}()

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case err := <-errC:
		if err != nil {
			return "", err
		}
		return wine.JoinURL(b.baseURL, obj.Name), nil
	}
}
