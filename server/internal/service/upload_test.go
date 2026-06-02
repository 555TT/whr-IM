package service

import (
	"context"
	"testing"
)

func TestStaticObjectStorageTrimsTrailingSlash(t *testing.T) {
	storage := NewStaticObjectStorage("http://localhost:9000/media/")

	url, err := storage.UploadImage(context.Background(), "uploads/users/1/a.png", "image/png", []byte("x"))
	if err != nil {
		t.Fatalf("expected no upload error, got %v", err)
	}

	want := "http://localhost:9000/media/uploads/users/1/a.png"
	if url != want {
		t.Fatalf("expected url %q, got %q", want, url)
	}
}
