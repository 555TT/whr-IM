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

	expectedDSN := "root:easy-chat@tcp(192.168.18.66:13306)/whr_im?charset=utf8mb4&parseTime=True&loc=Local"
	if cfg.MySQL.DSN != expectedDSN {
		t.Fatalf("expected mysql dsn %q, got %q", expectedDSN, cfg.MySQL.DSN)
	}
}
