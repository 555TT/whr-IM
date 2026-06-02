package config

import "testing"

func TestLoadReadsMySQLAndServerConfigFromYAML(t *testing.T) {
	cfg, err := Load("../../config/config.yaml")
	if err != nil {
		t.Fatalf("expected config load success, got error: %v", err)
	}

	if cfg.Server.Port != ":8080" {
		t.Fatalf("expected server port :8080, got %q", cfg.Server.Port)
	}

	expectedDSN := "root:wanghaoran666@tcp(localhost:3306)/whr_im?charset=utf8mb4&parseTime=True&loc=Local"
	if cfg.MySQL.DSN != expectedDSN {
		t.Fatalf("expected mysql dsn %q, got %q", expectedDSN, cfg.MySQL.DSN)
	}

	if cfg.ObjectStorage.Endpoint != "localhost:9000" {
		t.Fatalf("expected object storage endpoint localhost:9000, got %q", cfg.ObjectStorage.Endpoint)
	}
	if cfg.ObjectStorage.Bucket != "whr-im" {
		t.Fatalf("expected object storage bucket whr-im, got %q", cfg.ObjectStorage.Bucket)
	}
	if cfg.ObjectStorage.PublicBaseURL != "http://localhost:9000/whr-im" {
		t.Fatalf("expected object storage public base url http://localhost:9000/whr-im, got %q", cfg.ObjectStorage.PublicBaseURL)
	}
}
