package config

import (
	"errors"
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/caarlos0/env"
	"github.com/joho/godotenv"
)

const minJWTSecretLength = 32

type Schema struct {
	Environment          string        `env:"GO_ENV" envDefault:"development"`
	Port                 string        `env:"PORT" envDefault:"8080"`
	AppUrl               string        `env:"APP_URL" envDefault:"http://localhost:8080"`
	PostgresUri          string        `env:"POSTGRES_URI"`
	DisableAutoMigration bool          `env:"DISABLE_AUTO_MIGRATION" envDefault:"true"`
	CorsAllowOrigins     []string      `env:"CORS_ALLOW_ORIGINS" envSeparator:"," envDefault:"http://localhost:3000"`
	JWTSecret            string        `env:"JWT_SECRET"`
	JWTIssuer            string        `env:"JWT_ISSUER" envDefault:"evently-app"`
	JWTAccessTTL         time.Duration `env:"JWT_ACCESS_TTL" envDefault:"1h"`
	GAuthClientId        string        `env:"GAUTH_CLIENT_ID"`
	GAuthClientSecret    string        `env:"GAUTH_CLIENT_SECRET"`
	GAuthRedirectUrl     string        `env:"GAUTH_CLIENT_REDIRECT_URI"`
	GAuthUserInfoUrl     string        `env:"GAUTH_USER_INFO_URL" envDefault:"https://www.googleapis.com/oauth2/v2/userinfo"`
}

var cfg Schema

func GetConfig() *Schema {
	_ = godotenv.Load(filepath.Join(".env"))

	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("Error on parsing configuration, error: %v", err)
	}

	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	// Schema implements fmt.Stringer, so secrets are redacted here.
	log.Printf("Configuration loaded: %v", cfg)

	return &cfg
}

// Validate checks that every setting required to run the service is present.
func (s Schema) Validate() error {
	var errs []error
	required := []struct{ key, value string }{
		{"POSTGRES_URI", s.PostgresUri},
		{"JWT_SECRET", s.JWTSecret},
		{"GAUTH_CLIENT_ID", s.GAuthClientId},
		{"GAUTH_CLIENT_SECRET", s.GAuthClientSecret},
		{"GAUTH_CLIENT_REDIRECT_URI", s.GAuthRedirectUrl},
	}
	for _, r := range required {
		if strings.TrimSpace(r.value) == "" {
			errs = append(errs, fmt.Errorf("%s is required", r.key))
		}
	}
	if s.JWTSecret != "" && len(s.JWTSecret) < minJWTSecretLength {
		errs = append(errs, fmt.Errorf("JWT_SECRET must be at least %d characters", minJWTSecretLength))
	}
	if s.JWTAccessTTL <= 0 {
		errs = append(errs, errors.New("JWT_ACCESS_TTL must be a positive duration"))
	}
	return errors.Join(errs...)
}

// IsDevelopment reports whether the service runs in the development environment.
func (s Schema) IsDevelopment() bool {
	return s.Environment == "development"
}

// String returns a log-safe representation of the configuration: secrets and
// the database DSN (which usually embeds a password) are never included.
func (s Schema) String() string {
	return fmt.Sprintf(
		"{Environment:%s Port:%s AppUrl:%s PostgresUri:%s DisableAutoMigration:%t CorsAllowOrigins:%v "+
			"JWTSecret:%s JWTIssuer:%s JWTAccessTTL:%s GAuthClientId:%s GAuthClientSecret:%s "+
			"GAuthRedirectUrl:%s GAuthUserInfoUrl:%s}",
		s.Environment, s.Port, s.AppUrl, redact(s.PostgresUri), s.DisableAutoMigration, s.CorsAllowOrigins,
		redact(s.JWTSecret), s.JWTIssuer, s.JWTAccessTTL, redact(s.GAuthClientId), redact(s.GAuthClientSecret),
		s.GAuthRedirectUrl, s.GAuthUserInfoUrl,
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
