package config

import (
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server        ServerConfig        `yaml:"server"`
	MySQL         MySQLConfig         `yaml:"mysql"`
	ObjectStorage ObjectStorageConfig `yaml:"objectStorage"`
}

type ServerConfig struct {
	Port string `yaml:"port"`
}

type MySQLConfig struct {
	DSN string `yaml:"dsn"`
}

type ObjectStorageConfig struct {
	Endpoint      string `yaml:"endpoint"`
	AccessKey     string `yaml:"accessKey"`
	SecretKey     string `yaml:"secretKey"`
	Bucket        string `yaml:"bucket"`
	UseSSL        bool   `yaml:"useSSL"`
	PublicBaseURL string `yaml:"publicBaseUrl"`
}

func Load(path string) (*Config, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(content, &cfg); err != nil {
		return nil, err
	}
	applyEnvOverrides(&cfg)
	return &cfg, nil
}

func applyEnvOverrides(cfg *Config) {
	if value := os.Getenv("MYSQL_DSN"); value != "" {
		cfg.MySQL.DSN = value
	}
	if value := os.Getenv("MINIO_ENDPOINT"); value != "" {
		cfg.ObjectStorage.Endpoint = value
	}
	if value := os.Getenv("MINIO_ACCESS_KEY"); value != "" {
		cfg.ObjectStorage.AccessKey = value
	}
	if value := os.Getenv("MINIO_SECRET_KEY"); value != "" {
		cfg.ObjectStorage.SecretKey = value
	}
	if value := os.Getenv("MINIO_BUCKET"); value != "" {
		cfg.ObjectStorage.Bucket = value
	}
	if value := os.Getenv("MINIO_PUBLIC_BASE_URL"); value != "" {
		cfg.ObjectStorage.PublicBaseURL = value
	}
	if value := os.Getenv("MINIO_USE_SSL"); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			cfg.ObjectStorage.UseSSL = parsed
		}
	}
}
