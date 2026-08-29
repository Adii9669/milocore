package storage

import (
	"context"
	"io"
)

type FileStorage interface {
	Upload(ctx context.Context,
		objectName string,
		reader io.Reader,
		size int64,
		contentType string,
	) error
	Delete(ctx context.Context, bucket, objectName string) error
}
