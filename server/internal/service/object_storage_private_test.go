package service

import (
	"context"
	"strings"
	"testing"

	"whr-im/server/internal/config"
)

func TestMinioObjectStorageCanGenerateObjectURLForPrivateBucket(t *testing.T) {
	storage, err := NewObjectStorageFromConfig(config.ObjectStorageConfig{
		Endpoint:      "121.41.66.206:9000",
		AccessKey:     "minioadmin",
		SecretKey:     "minioadmin123",
		Bucket:        "whr-im",
		UseSSL:        false,
		PublicBaseURL: "http://121.41.66.206:9000/whr-im",
	})
	if err != nil {
		t.Fatalf("expected no config error, got %v", err)
	}

	resolver, ok := storage.(interface {
		ObjectURL(context.Context, string) (string, error)
	})
	if !ok {
		t.Fatalf("expected storage to expose object url resolver, got %T", storage)
	}

	url, err := resolver.ObjectURL(context.Background(), "uploads/users/1/demo.png")
	if err != nil {
		t.Fatalf("expected object url generation success, got %v", err)
	}
	if url == "http://121.41.66.206:9000/whr-im/uploads/users/1/demo.png" {
		t.Fatalf("expected private bucket url to not be raw public path, got %q", url)
	}
	if !strings.Contains(url, "X-Amz-Algorithm") && !strings.Contains(url, "X-Amz-Signature") {
		t.Fatalf("expected presigned-style url, got %q", url)
	}
}
