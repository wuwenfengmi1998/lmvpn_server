package config

import (
	cryptorand "crypto/rand"
	"encoding/base64"
	"os"
	"path/filepath"

	"github.com/goccy/go-yaml"
)

type WebConfig struct {
	Port        int    `yaml:"port"`
	Sock        string `yaml:"sock"`
	SockMode    string `yaml:"sock_mode"`
	SockGroup   string `yaml:"sock_group"`
	SockDirMode string `yaml:"sock_dir_mode"`
	JWTSecret   string `yaml:"jwt_secret"`
}

type DatabaseConfig struct {
	Type string `yaml:"type"`
	Path string `yaml:"path"`
	DSN  string `yaml:"dsn"`
}

type Config struct {
	Web      WebConfig      `yaml:"web"`
	Database DatabaseConfig `yaml:"database"`
}

func defaultConfig() *Config {
	return &Config{
		Web: WebConfig{
			Port:        8080,
			Sock:        "/run/lmvpnweb.sock",
			SockMode:    "0666",
			SockDirMode: "0755",
		},
		Database: DatabaseConfig{
			Type: "sqlite",
			Path: "data/lmvpn.db",
			DSN:  "",
		},
	}
}

func Load(path string) (*Config, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, err
	}

	cfg := defaultConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
		if err := resolveJWTSecret(cfg); err != nil {
			return nil, err
		}
		if err := saveConfig(path, cfg); err != nil {
			return nil, err
		}
		return cfg, nil
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	if err := resolveJWTSecret(cfg); err != nil {
		return nil, err
	}

	if err := saveConfig(path, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func resolveJWTSecret(cfg *Config) error {
	if envSecret := os.Getenv("LMVPN_JWT_SECRET"); envSecret != "" {
		cfg.Web.JWTSecret = envSecret
		return nil
	}
	if cfg.Web.JWTSecret != "" {
		return nil
	}
	secret, err := generateRandomSecret(32)
	if err != nil {
		return err
	}
	cfg.Web.JWTSecret = secret
	return nil
}

func generateRandomSecret(n int) (string, error) {
	b := make([]byte, n)
	if _, err := cryptorand.Read(b); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}

func saveConfig(path string, cfg *Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}
