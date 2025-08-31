package config

import (
	"log"

	"github.com/MostajeranMohammad/dekamond-auth-challenge/pkg/utils"
	"github.com/ilyakaznacheev/cleanenv"
)

type (
	Config struct {
		HTTP
		Log
		PG
		Redis
		AUTH
	}

	HTTP struct {
		Port string `validate:"required" env:"HTTP_PORT"`
	}

	Log struct {
		Level string `validate:"required" env:"LOG_LEVEL"`
	}

	PG struct {
		DSN           string `validate:"required" env:"PG_DSN"`
		RunMigrations bool   `validate:"required" env:"RUN_MIGRATIONS"`
	}

	Redis struct {
		Addr     string `validate:"required" env:"REDIS_ADDR"`
		Password string `validate:"required" env:"REDIS_PASSWORD"`
		DB       int    `validate:"required" env:"REDIS_DB"`
	}

	AUTH struct {
		JwtSecret string `validate:"required"  env:"JWT_SECRET"`
	}
)

func NewConfig() (*Config, error) {
	cfg := &Config{}

	parseConfigFiles([]string{"./.env"}, cfg)

	err := cleanenv.ReadEnv(cfg)
	if err != nil {
		return nil, err
	}

	err = utils.ValidateStruct(cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func parseConfigFiles(files []string, cfg *Config) {
	for _, path := range files {
		err := cleanenv.ReadConfig(path, cfg)
		if err != nil {
			log.Printf("WARN: config error: %v\n", err)
		}
	}
}
