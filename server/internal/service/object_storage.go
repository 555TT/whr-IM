package service

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"whr-im/server/internal/config"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioObjectStorage struct {
	client        *minio.Client
	bucket        string
	publicBaseURL string
}

func NewObjectStorageFromConfig(cfg config.ObjectStorageConfig) (ObjectStorage, error) {
	if cfg.Endpoint == "" {
		return NewStaticObjectStorage(cfg.PublicBaseURL), nil
	}
	if cfg.Bucket == "" {
		return nil, fmt.Errorf("object storage bucket is required")
	}

	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, err
	}

	return &MinioObjectStorage{
		client:        client,
		bucket:        cfg.Bucket,
		publicBaseURL: strings.TrimRight(cfg.PublicBaseURL, "/"),
	}, nil
}

func (s *MinioObjectStorage) UploadImage(ctx context.Context, objectKey string, contentType string, data []byte) (string, error) {
	reader := bytes.NewReader(data)
	_, err := s.client.PutObject(ctx, s.bucket, objectKey, reader, int64(len(data)), minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return "", err
	}
	return s.ObjectURL(ctx, objectKey)
}

func (s *MinioObjectStorage) ObjectURL(_ context.Context, objectKey string) (string, error) {
	if s.publicBaseURL != "" {
		if strings.HasPrefix(s.publicBaseURL, "http://") || strings.HasPrefix(s.publicBaseURL, "https://") {
			presigned, err := s.client.PresignedGetObject(context.Background(), s.bucket, objectKey, 24*time.Hour, url.Values{})
			if err != nil {
				return "", err
			}
			return presigned.String(), nil
		}
	}
	presigned, err := s.client.PresignedGetObject(context.Background(), s.bucket, objectKey, 24*time.Hour, url.Values{})
	if err != nil {
		return "", err
	}
	return presigned.String(), nil
}
