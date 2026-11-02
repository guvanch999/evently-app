package config

import (
	"fmt"
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

	// Schema implements fmt.Stringer, so secrets are redacted here.
	log.Printf("Configuration loaded: %v", cfg)

	return &cfg
}

// String returns a log-safe representation of the configuration: secrets and
// the database DSN (which usually embeds a password) are never included.
func (s Schema) String() string {
	return fmt.Sprintf(
		"{Environment:%s Port:%s AppUrl:%s PostgresUri:%s DisableAutoMigration:%t SecretKeyForHash:%s "+
			"GAuthClientId:%s GAuthClientSecret:%s GAuthRedirectUrl:%s GAuthUserInfoUrl:%s}",
		s.Environment, s.Port, s.AppUrl, redact(s.PostgresUri), s.DisableAutoMigration, redact(s.SecretKeyForHash),
		redact(s.GAuthClientId), redact(s.GAuthClientSecret), s.GAuthRedirectUrl, s.GAuthUserInfoUrl,
	)
}

// GoString keeps %#v from bypassing the redaction in String.
func (s Schema) GoString() string {
	return "config.Schema" + s.String()
}

func redact(value string) string {
	if value == "" {
		return "<unset>"
	}
	return "<redacted>"
}
