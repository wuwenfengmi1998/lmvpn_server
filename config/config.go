package config

import (
	"os"
	"path/filepath"

	"github.com/goccy/go-yaml"
)

type WebConfig struct {
	Port int    `yaml:"port"`
	Sock string `yaml:"sock"`
}

type Config struct {
	Web WebConfig `yaml:"web"`
}

func defaultConfig() *Config {
	return &Config{
		Web: WebConfig{
			Port: 8080,
			Sock: "/run/lmvpnweb.sock",
		},
	}
}

func Load(path string) (*Config, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	cfg := defaultConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		if !os.IsNotExist(err) {
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

	if err := saveConfig(path, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func saveConfig(path string, cfg *Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
