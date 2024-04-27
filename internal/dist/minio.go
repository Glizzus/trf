package dist

import (
	"context"
	"fmt"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Minio struct {
	Bucket string
	client *minio.Client
}

func NewMinio(bucket, endpoint, accessKey, secretKey string, useSSL bool) (*Minio, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create Minio client: %w", err)
	}
	if err = client.MakeBucket(context.Background(), bucket, minio.MakeBucketOptions{}); err != nil {
		response := minio.ToErrorResponse(err)
		switch response.Code {
		case "BucketAlreadyOwnedByYou", "BucketAlreadyExists":
			break
		default:
			return nil, fmt.Errorf("unable to create bucket: %w", err)
		}
	}
	return &Minio{
		Bucket: bucket,
		client: client,
	}, nil
}

func (m *Minio) Has(ctx context.Context, slug string) (bool, error) {
	_, err := m.client.StatObject(ctx, m.Bucket, slug, minio.StatObjectOptions{})
	if err != nil {
		response := minio.ToErrorResponse(err)
		if response.Code == "NoSuchKey" {
			return false, nil
		}
		return false, fmt.Errorf("unable to check if object exists: %w", err)
	}
	return true, nil
}

func (m *Minio) Put(ctx context.Context, slug string, reader io.Reader) error {
	const sizeUnknown = -1
	path := slug + ".html"
	_, err := m.client.PutObject(ctx, m.Bucket, path, reader, sizeUnknown, minio.PutObjectOptions{})
	if err != nil {
		return fmt.Errorf("unable to put object: %w", err)
	}
	return nil
}
