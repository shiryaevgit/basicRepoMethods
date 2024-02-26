package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	DatabaseURL string `json:"database_url"`
	HTTPPort    int    `json:"http_port"`
}

func LoadConfig(filePath string) (Config, error) {

	fmt.Println(filePath)
	configFile, err := os.Open(filePath)
	if err != nil {
		return Config{}, fmt.Errorf("os.Open(): %w", err)
	}
	defer configFile.Close()

	var config Config
	decoder := json.NewDecoder(configFile)
	if err = decoder.Decode(&config); err != nil {
		return Config{}, fmt.Errorf("decoder.Decode(): %w", err)
	}
	return config, nil
}
