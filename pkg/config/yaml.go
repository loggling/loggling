// Package config handles the initialization and hot-reloading of project settings.
// yaml.go implements config loading from YAML files using standard library and third-party decoders.
package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type DefaultConfig struct {
	Inputs   []string `yaml:"inputs"`
	Output   string   `yaml:"output"`
	DLQ string 
	Registry string   `yaml:"registry"`
}
type Config struct {
	Default  DefaultConfig
	Server   ServerConfig
	Pipeline map[string][]ParamsConfig `yaml:"pipeline"`
}

type ServerConfig struct {
	Enabled bool `yaml:"enabled"`
	Port    int  `yaml:"port"`
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
