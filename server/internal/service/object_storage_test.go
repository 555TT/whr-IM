package service

import (
	"testing"

	"whr-im/server/internal/config"
)

func TestNewObjectStorageFromConfigUsesStaticStorageWhenEndpointMissing(t *testing.T) {
	storage, err := NewObjectStorageFromConfig(config.ObjectStorageConfig{
		PublicBaseURL: "http://localhost:9000/whr-im",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if _, ok := storage.(*StaticObjectStorage); !ok {
		t.Fatalf("expected static object storage, got %T", storage)
	}
}

func TestNewObjectStorageFromConfigRequiresBucketForMinIO(t *testing.T) {
	_, err := NewObjectStorageFromConfig(config.ObjectStorageConfig{
		Endpoint:      "localhost:9000",
		AccessKey:     "minioadmin",
		SecretKey:     "minioadmin",
		PublicBaseURL: "http://localhost:9000/whr-im",
	})
	if err == nil {
		t.Fatal("expected missing bucket error")
	}
}
