package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Project struct {
	Name        string `yaml:"name"`
	ProcessName string `yaml:"processName"`
	Port        int    `yaml:"port"`
}

type Proxy struct {
	ProcessName     string `yaml:"processName"`
	Port            int    `yaml:"port"`
	ClashApiPort    int    `yaml:"clashApiPort"`
	TestURL         string `yaml:"testUrl"`
	SubscriptionURL string `yaml:"subscriptionUrl"`
}

type Config struct {
	Listen   string    `yaml:"listen"`
	Projects []Project `yaml:"projects"`
	Proxy    Proxy     `yaml:"proxy"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	cfg := &Config{Listen: ":9527"}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
