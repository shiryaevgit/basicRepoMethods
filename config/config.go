package config

import (
	"fmt"
	"github.com/spf13/viper"
)

type Config struct {
	ServerConfig
	DatabaseConfig
}

type ServerConfig struct {
	HTTPPort int
}

type DatabaseConfig struct {
	DatabaseURL string
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
