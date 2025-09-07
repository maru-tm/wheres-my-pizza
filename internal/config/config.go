package config

import (
	"fmt"
	"restaurant-system/pkg/yaml"
)

type Config struct {
	Database DatabaseConfig
	RabbitMQ RabbitMQConfig
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
}

type RabbitMQConfig struct {
	Host     string
	Port     int
	User     string
	Password string
}

func Load(path string) (*Config, error) {
	data, err := yaml.ParseFile(path)
	if err != nil {
		return nil, fmt.Errorf("не удалось прочитать yaml: %w", err)
	}

	cfg := &Config{
		Database: DatabaseConfig{
			Host:     data["database"]["host"],
			Port:     yaml.Atoi(data["database"]["port"]),
			User:     data["database"]["user"],
			Password: data["database"]["password"],
			Database: data["database"]["database"],
		},
		RabbitMQ: RabbitMQConfig{
			Host:     data["rmq"]["host"],
			Port:     yaml.Atoi(data["rmq"]["port"]),
			User:     data["rmq"]["user"],
			Password: data["rmq"]["password"],
		},
	}

	return cfg, nil
}
