package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type DefaultConfig struct {
	InputPath  string `yaml:"input"`
	OutputPath string `yaml:"output"`
}
type Config struct {
	Default  DefaultConfig
	Pipeline map[string][]ParamsConfig `yaml:"pipeline"`
}
type ParamsConfig map[string]string

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)

	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(data, &config)

	if err != nil {
		return nil, err
	}

	return &config, nil
}
