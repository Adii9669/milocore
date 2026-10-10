package storage

import (
	"chat-server/internals/config"
	"context"
	"fmt"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioStorage struct {
	Client *minio.Client
	Bucket string
}

func (s *MinioStorage) Upload(ctx context.Context,
	objectName string,
	reader io.Reader,
	size int64,
	contentType string,
) error {
	_, err := s.Client.PutObject(
		ctx,
		s.Bucket,
		objectName,
		reader,
		size,
		minio.PutObjectOptions{
			ContentType: contentType,
		},
	)
	return err
}

func NewMinioStorage(cfg config.AppConfig) (*MinioStorage, error) {
	client, err := minio.New(
		cfg.MinIO.Endpoint,
		&minio.Options{
			Creds: credentials.NewStaticV4(
				cfg.MinIO.Username,
				cfg.MinIO.Password,
				"",
			),
			Secure: false,
		},
	)
	if err != nil {
		return nil, err
	}

	return &MinioStorage{
		Client: client,
		Bucket: cfg.MinIO.Bucket,
	}, nil
}

func (s *MinioStorage) CheckConnection(ctx context.Context) error {

	exists, err := s.Client.BucketExists(ctx, s.Bucket)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("bucket %q dose not exists", s.Bucket)
	}
	return nil
}

// --For Deletion
func (s *MinioStorage) Delete(ctx context.Context, bucket, objectName string) error {
	return s.Client.RemoveObject(
		ctx,
		bucket,
		objectName,
		minio.RemoveObjectOptions{},
	)
}
