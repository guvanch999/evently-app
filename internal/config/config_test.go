package config

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func validSchema() Schema {
	return Schema{
		Environment:       "development",
		Port:              "8080",
		PostgresUri:       "postgres://evently:db-password-123@localhost:5432/evently",
		JWTSecret:         "jwt-secret-value-that-is-long-enough-123",
		JWTIssuer:         "evently-app",
		JWTAccessTTL:      time.Hour,
		GAuthClientId:     "client-id-value.apps.googleusercontent.com",
		GAuthClientSecret: "client-secret-value",
		GAuthRedirectUrl:  "http://localhost:8080/v1/auth/google/callback",
		GAuthUserInfoUrl:  "https://www.googleapis.com/oauth2/v2/userinfo",
	}
}

func TestSchemaFormattingRedactsSecrets(t *testing.T) {
	s := validSchema()
	secrets := []string{"db-password-123", s.JWTSecret, s.GAuthClientId, s.GAuthClientSecret}

	for _, verb := range []string{"%v", "%+v", "%#v", "%s"} {
		for _, value := range []interface{}{s, &s} {
			out := fmt.Sprintf(verb, value)
			for _, secret := range secrets {
				if strings.Contains(out, secret) {
					t.Fatalf("%s output leaks a secret: %s", verb, out)
				}
			}
			if !strings.Contains(out, "<redacted>") {
				t.Fatalf("%s output should mark redacted fields: %s", verb, out)
			}
		}
	}

	if out := s.String(); !strings.Contains(out, s.GAuthRedirectUrl) {
		t.Fatalf("non-secret fields should stay visible: %s", out)
	}
}

func TestValidate(t *testing.T) {
	if err := validSchema().Validate(); err != nil {
		t.Fatalf("expected valid config, got %v", err)
	}

	missing := validSchema()
	missing.JWTSecret = ""
	missing.GAuthClientSecret = " "
	err := missing.Validate()
	if err == nil {
		t.Fatal("expected an error for missing settings")
	}
	for _, key := range []string{"JWT_SECRET", "GAUTH_CLIENT_SECRET"} {
		if !strings.Contains(err.Error(), key) {
			t.Fatalf("error should mention %s: %v", key, err)
		}
	}

	short := validSchema()
	short.JWTSecret = "too-short"
	if err := short.Validate(); err == nil || !strings.Contains(err.Error(), "at least") {
		t.Fatalf("expected a minimum length error, got %v", err)
	}

	noTTL := validSchema()
	noTTL.JWTAccessTTL = 0
	if err := noTTL.Validate(); err == nil {
		t.Fatal("expected an error for a zero JWT_ACCESS_TTL")
	}
}
