package config

import (
	"os"

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
	return &cfg, nil
}
