package service

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

var ErrInvalidUpload = fmt.Errorf("invalid upload")

type UploadedImage struct {
	ObjectKey string `json:"objectKey"`
	URL       string `json:"url"`
}

type ObjectStorage interface {
	UploadImage(ctx context.Context, objectKey string, contentType string, data []byte) (string, error)
	ObjectURL(ctx context.Context, objectKey string) (string, error)
}

type UploadService struct {
	storage ObjectStorage
}

func NewUploadService(storage ObjectStorage) *UploadService {
	return &UploadService{storage: storage}
}

func (s *UploadService) UploadImage(ctx context.Context, userID uint64, filename string, contentType string, data []byte) (*UploadedImage, error) {
	if s.storage == nil {
		return nil, fmt.Errorf("upload storage is not configured")
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("%w: file is required", ErrInvalidUpload)
	}
	if !strings.HasPrefix(contentType, "image/") {
		return nil, fmt.Errorf("%w: only image upload is supported", ErrInvalidUpload)
	}

	ext := strings.ToLower(filepath.Ext(filename))
	if ext == "" {
		ext = ".bin"
	}
	objectKey := fmt.Sprintf("uploads/users/%d/%d%s", userID, time.Now().UnixNano(), ext)
	url, err := s.storage.UploadImage(ctx, objectKey, contentType, data)
	if err != nil {
		return nil, err
	}

	return &UploadedImage{ObjectKey: objectKey, URL: url}, nil
}

type StaticObjectStorage struct {
	baseURL string
}

func NewStaticObjectStorage(baseURL string) *StaticObjectStorage {
	return &StaticObjectStorage{baseURL: strings.TrimRight(baseURL, "/")}
}

func (s *StaticObjectStorage) UploadImage(_ context.Context, objectKey string, _ string, _ []byte) (string, error) {
	return s.baseURL + "/" + objectKey, nil
}

func (s *StaticObjectStorage) ObjectURL(_ context.Context, objectKey string) (string, error) {
	return s.baseURL + "/" + objectKey, nil
}
