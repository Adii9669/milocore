package storage

import (
	"chat-server/internals/config"
	"context"
	"io"
	"mime"
	"net/url"
	"time"

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
			Secure: cfg.MinIO.Secure,
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

func (s *MinioStorage) EnsureBucket(ctx context.Context) error {
	exists, err := s.Client.BucketExists(ctx, s.Bucket)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	return s.Client.MakeBucket(ctx, s.Bucket, minio.MakeBucketOptions{})
}

func (s *MinioStorage) CheckConnection(ctx context.Context) error {
	return s.EnsureBucket(ctx)
}

func (s *MinioStorage) DeleteObject(ctx context.Context, objectName string) error {
	return s.Delete(ctx, s.Bucket, objectName)
}

func (s *MinioStorage) PresignedGetObject(ctx context.Context, objectName, filename, disposition string, expiry time.Duration) (string, error) {
	headerValue := mime.FormatMediaType(disposition, map[string]string{"filename": filename})
	if headerValue == "" {
		headerValue = disposition
	}
	params := url.Values{"response-content-disposition": []string{headerValue}}
	url, err := s.Client.PresignedGetObject(ctx, s.Bucket, objectName, expiry, params)
	if err != nil {
		return "", err
	}
	return url.String(), nil
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
