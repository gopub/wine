package oss

import (
	"bytes"
	"context"
	"fmt"
	"net/url"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/gopub/errors"
	"github.com/gopub/wine/exp/storage"
	"github.com/gopub/wine/urlutil"
)

const (
	cacheControl = "public, max-age=14400"
)

type Bucket struct {
	bucket       *oss.Bucket
	ACL          oss.ACLType
	CacheControl string
	baseURL      string
}

var _ storage.Writer = (*Bucket)(nil)

func NewBucket(endpoint, keyID, keySecret, name string) (*Bucket, error) {
	if endpoint == "" {
		return nil, errors.BadRequest("missing endpoint")
	}

	if keyID == "" {
		return nil, errors.BadRequest("missing keyID")
	}

	if keySecret == "" {
		return nil, errors.BadRequest("missing keySecret")
	}

	if name == "" {
		return nil, errors.BadRequest("missing name")
	}

	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, errors.BadRequest("invalid endpoint: %w", err)
	}

	u.Host = fmt.Sprintf("%s.%s", name, u.Host)
	client, err := oss.New(endpoint, keyID, keySecret)
	if err != nil {
		fmt.Errorf("create oss client: %w", err)
	}

	b, err := client.Bucket(name)
	if err != nil {
		fmt.Errorf("create oss bucket: %w", err)
	}

	return &Bucket{
		bucket:       b,
		ACL:          oss.ACLPublicRead,
		CacheControl: cacheControl,
		baseURL:      u.String(),
	}, nil
}

func (b *Bucket) Name() string {
	return b.bucket.BucketName
}

func (b *Bucket) Write(ctx context.Context, obj *storage.Object) (string, error) {
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
		return urlutil.Join(b.baseURL, obj.Name), nil
	}
}
