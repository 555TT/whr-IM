package config

import (
	"os"
	"testing"
)

func TestLoadReadsMySQLAndServerConfigFromYAML(t *testing.T) {
	cfg, err := Load("../../config/config.yaml")
	if err != nil {
		t.Fatalf("expected config load success, got error: %v", err)
	}

	if cfg.Server.Port != ":8080" {
		t.Fatalf("expected server port :8080, got %q", cfg.Server.Port)
	}

	if cfg.MySQL.DSN == "" {
		t.Fatal("expected mysql dsn to be populated from yaml")
	}
	if cfg.ObjectStorage.Endpoint == "" {
		t.Fatal("expected object storage endpoint to be populated from yaml")
	}
	if cfg.ObjectStorage.Bucket == "" {
		t.Fatal("expected object storage bucket to be populated from yaml")
	}
	if cfg.ObjectStorage.PublicBaseURL == "" {
		t.Fatal("expected object storage public base url to be populated from yaml")
	}
}

func TestLoadAppliesEnvironmentOverrides(t *testing.T) {
	t.Setenv("MYSQL_DSN", "mysql-from-env")
	t.Setenv("MINIO_ENDPOINT", "minio.example.com:9000")
	t.Setenv("MINIO_ACCESS_KEY", "ak-env")
	t.Setenv("MINIO_SECRET_KEY", "sk-env")
	t.Setenv("MINIO_BUCKET", "bucket-env")
	t.Setenv("MINIO_USE_SSL", "true")
	t.Setenv("MINIO_PUBLIC_BASE_URL", "https://cdn.example.com/bucket")

	cfg, err := Load("../../config/config.yaml")
	if err != nil {
		t.Fatalf("expected config load success, got error: %v", err)
	}

	if cfg.MySQL.DSN != "mysql-from-env" {
		t.Fatalf("expected mysql dsn from env, got %q", cfg.MySQL.DSN)
	}
	if cfg.ObjectStorage.Endpoint != "minio.example.com:9000" {
		t.Fatalf("expected endpoint from env, got %q", cfg.ObjectStorage.Endpoint)
	}
	if cfg.ObjectStorage.AccessKey != "ak-env" {
		t.Fatalf("expected access key from env, got %q", cfg.ObjectStorage.AccessKey)
	}
	if cfg.ObjectStorage.SecretKey != "sk-env" {
		t.Fatalf("expected secret key from env, got %q", cfg.ObjectStorage.SecretKey)
	}
	if cfg.ObjectStorage.Bucket != "bucket-env" {
		t.Fatalf("expected bucket from env, got %q", cfg.ObjectStorage.Bucket)
	}
	if !cfg.ObjectStorage.UseSSL {
		t.Fatal("expected useSSL true from env")
	}
	if cfg.ObjectStorage.PublicBaseURL != "https://cdn.example.com/bucket" {
		t.Fatalf("expected public base url from env, got %q", cfg.ObjectStorage.PublicBaseURL)
	}
}

func TestLoadIgnoresInvalidBooleanOverride(t *testing.T) {
	t.Setenv("MINIO_USE_SSL", "not-a-bool")

	cfg, err := Load("../../config/config.yaml")
	if err != nil {
		t.Fatalf("expected config load success, got error: %v", err)
	}

	_, exists := os.LookupEnv("MINIO_USE_SSL")
	if !exists {
		t.Fatal("expected MINIO_USE_SSL env to exist in test")
	}
	if cfg.ObjectStorage.UseSSL {
		t.Fatal("expected invalid bool env override to be ignored")
	}
}
