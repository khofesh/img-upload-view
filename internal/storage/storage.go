package storage

import (
	"context"
	"errors"
	"time"
)

// ErrNotFound is returned when an object does not exist in the store.
var ErrNotFound = errors.New("object not found")

type PresignedRequest struct {
	URL       string
	Method    string
	Headers   map[string]string
	ExpiresAt time.Time
}

type ObjectInfo struct {
	Size        int64
	ContentType string
}

type ObjectStore interface {
	PresignPut(ctx context.Context, key, contentType string, size int64) (PresignedRequest, error)
	PresignGet(ctx context.Context, key string) (string, error)
	Head(ctx context.Context, key string) (ObjectInfo, error)
	Delete(ctx context.Context, key string) error
	EnsureBucket(ctx context.Context) error
}
