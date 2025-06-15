package objectstorage

import (
	"context"
	"io"
	"time"
)

type ObjectStorage interface {
	Upload(ctx context.Context, bucket string, key string, reader io.Reader) error
	Download(ctx context.Context, bucket string, key string) (io.ReadCloser, error)
	Delete(ctx context.Context, bucket string, key string) error
	PresignedURL(ctx context.Context, bucket string, key string, expires time.Duration) (string, error)
}

func NewObjectStorage(endpoint, accessKeyID, secretAccessKey string) (ObjectStorage, error) {
	return newMinIOObjectStorage(endpoint, accessKeyID, secretAccessKey)
}
