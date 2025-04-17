package config

import (
	"fmt"
	"log"

	"github.com/Tech-Arch1tect/config"
)

// use global variable to store config as this is generally just a POC
var C *Config

type Config struct {
	StoragePath string `env:"STORAGE_PATH" validate:"required"`
}

func (c *Config) SetDefaults() {
	c.StoragePath = "/rkllm"
}

func LoadConfig() {
	fmt.Println("Loading config")
	var cfg Config
	if err := config.Load(&cfg); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	fmt.Printf("Loaded configuration: %+v\n", cfg)
	C = &cfg
}
