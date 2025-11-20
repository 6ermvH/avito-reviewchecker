package config

import (
	"fmt"
	"os"

	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v3"
)

type Config struct {
	HTTP HTTPConfig `yaml:"server"`
	Log  LogConfig  `yaml:"log"`
	DB   DBConfig   `yaml:"db"`
}

type HTTPConfig struct {
	Addr         string `validate:"required" yaml:"addr"`
	ReadTimeout  int    `validate:"required" yaml:"readTimeout"`
	WriteTimeout int    `validate:"required" yaml:"writeTimeout"`
	IdleTimeout  int    `validate:"required" yaml:"idleTimeout"`
}

type LogConfig struct {
	Level string `validate:"required" yaml:"level"`
}

type DBConfig struct {
	DSN string `validate:"required" yaml:"dsn"`
}

func Load(path string) (*Config, error) {
	cfg := &Config{}

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("load config from %q: %w", path, err)
	}
	//nolint:errcheck
	defer file.Close()

	if err := yaml.NewDecoder(file).Decode(cfg); err != nil {
		return nil, fmt.Errorf("load config from %q: %w", path, err)
	}

	cfg.DB = DBConfig{DSN: os.Getenv("DSN")}

	validate := validator.New()
	if err := validate.Struct(cfg); err != nil {
		return cfg, fmt.Errorf("validatae config: %w", err)
	}

	return cfg, nil
}
