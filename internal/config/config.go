package config

import (
	"log"
	"path/filepath"

	"github.com/caarlos0/env"
	"github.com/joho/godotenv"
)

type Schema struct {
	Environment          string `env:"GO_ENV"`
	Port                 string `env:"PORT"`
	AppUrl               string `env:"APP_URL"`
	PostgresUri          string `env:"POSTGRES_URI"`
	DisableAutoMigration bool   `env:"DISABLE_AUTO_MIGRATION" envDefault:"true"`
	SecretKeyForHash     string `env:"SECRET_KEY_FOR_HASH"`
	GAuthClientId        string `env:"GAUTH_CLIENT_ID"`
	GAuthClientSecret    string `env:"GAUTH_CLIENT_SECRET"`
	GAuthRedirectUrl     string `env:"GAUTH_CLIENT_REDIRECT_URI"`
	GAuthUserInfoUrl     string `env:"GAUTH_USER_INFO_URL"`
}

var cfg Schema

func GetConfig() *Schema {
	_ = godotenv.Load(filepath.Join(".env"))

	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("Error on parsing configuration file, error: %v", err)
	}

	log.Printf("Конфигурация: %+v", cfg)

	return &cfg
}
