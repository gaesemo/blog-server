package objectstorage

import (
	"context"
	"io"
	"time"

	"github.com/minio/minio-go"
)

var _ ObjectStorage = (*minioObjectStorage)(nil)

type minioObjectStorage struct {
	client *minio.Client
}

func newMinIOObjectStorage(endpoint, accessKeyID, secretAccessKey string) (*minioObjectStorage, error) {
	client, err := minio.New(endpoint, accessKeyID, secretAccessKey, false)
	if err != nil {
		return nil, err
	}
	return &minioObjectStorage{
		client: client,
	}, nil
}

func (s *minioObjectStorage) Upload(ctx context.Context, bucket string, key string, reader io.Reader) error {
	return nil
}
func (s *minioObjectStorage) Download(ctx context.Context, bucket string, key string) (io.ReadCloser, error) {
	return nil, nil
}
func (s *minioObjectStorage) Delete(ctx context.Context, bucket string, key string) error {
	return nil
}
func (s *minioObjectStorage) PresignedURL(ctx context.Context, bucket string, key string, expires time.Duration) (string, error) {
	return "", nil
}
