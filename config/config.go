package config

import (
	"fmt"
	"github.com/spf13/viper"
)

type Config struct {
	ServerConfig
	PostgresConfig
	MongoConfig
}

type ServerConfig struct {
	HTTPPort int
}

type PostgresConfig struct {
	PostgresURL string
}

type MongoConfig struct {
	MongoURI string
}

func LoadConfig(path string) (Config, error) {

	var config Config
	viper.SetConfigFile(path)

	if err := viper.ReadInConfig(); err != nil {
		return config, fmt.Errorf("LoadConfig() ReadInConfig: %w", err)
	}

	if err := viper.Unmarshal(&config); err != nil {
		return config, fmt.Errorf("LoadConfig() Unmarshal: %w", err)
	}

	return config, nil
}
