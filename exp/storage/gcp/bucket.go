package gcp

import (
	"context"
	"fmt"
	"path/filepath"

	"cloud.google.com/go/storage"
	winestorage "github.com/gopub/wine/exp/storage"
)

var DefaultACL = []storage.ACLRule{{
	Entity: storage.AllUsers,
	Role:   storage.RoleReader,
}}

type Bucket struct {
	name         string
	handle       *storage.BucketHandle
	acl          []storage.ACLRule
	cacheControl string
	baseURL      string
}

var _ winestorage.Writer = (*Bucket)(nil)

func NewBucket(name, baseURL, cacheControl string, handle *storage.BucketHandle, acl []storage.ACLRule) *Bucket {
	return &Bucket{
		name:         name,
		handle:       handle,
		acl:          acl,
		cacheControl: cacheControl,
		baseURL:      baseURL,
	}
}

func (b *Bucket) Name() string {
	return b.name
}

func (b *Bucket) Write(ctx context.Context, obj *winestorage.Object) (string, error) {
	wc := b.handle.Object(obj.Name).NewWriter(ctx)
	wc.ACL = b.acl
	wc.ContentType = obj.Type
	wc.CacheControl = b.cacheControl
	if _, err := wc.Write(obj.Content); err != nil {
		return "", fmt.Errorf("write: %w", err)
	}

	if err := wc.Close(); err != nil {
		return "", fmt.Errorf("close: %w", err)
	}
	return filepath.Join(b.baseURL, obj.Name), nil
}
