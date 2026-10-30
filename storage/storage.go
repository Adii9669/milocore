package storage

import (
	"context"
	"io"
	"time"
)

type FileStorage interface {
	Upload(ctx context.Context,
		objectName string,
		reader io.Reader,
		size int64,
		contentType string,
	) error
	Delete(ctx context.Context, bucket, objectName string) error
	DeleteObject(ctx context.Context, objectName string) error
	EnsureBucket(ctx context.Context) error
	PresignedGetObject(ctx context.Context, objectName, filename, disposition string, expiry time.Duration) (string, error)
}
