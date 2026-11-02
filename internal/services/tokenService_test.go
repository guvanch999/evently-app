package services

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const testSecret = "test-secret-that-is-at-least-32-bytes!"

func TestIssueAndVerifyAccessToken(t *testing.T) {
	tokens := NewTokenService(testSecret, "evently-test", time.Hour)
	userId := uuid.New()

	signed, err := tokens.IssueAccessToken(userId)
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	if strings.Count(signed, ".") != 2 {
		t.Fatalf("expected a compact JWT, got %q", signed)
	}

	got, err := tokens.VerifyAccessToken(signed)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if got != userId {
		t.Fatalf("subject mismatch: got %s want %s", got, userId)
	}
}

func TestVerifyAccessTokenRejectsBadTokens(t *testing.T) {
	tokens := NewTokenService(testSecret, "evently-test", time.Hour)
	userId := uuid.New()
	valid, err := tokens.IssueAccessToken(userId)
	if err != nil {
		t.Fatal(err)
	}

	expired := NewTokenService(testSecret, "evently-test", time.Minute)
	expired.now = func() time.Time { return time.Now().Add(-2 * time.Hour) }
	expiredToken, _ := expired.IssueAccessToken(userId)

	otherSecret, _ := NewTokenService("another-secret-that-is-32-bytes-long!!", "evently-test", time.Hour).IssueAccessToken(userId)
	otherIssuer, _ := NewTokenService(testSecret, "someone-else", time.Hour).IssueAccessToken(userId)

	noneToken, _ := jwt.NewWithClaims(jwt.SigningMethodNone, jwt.RegisteredClaims{
		Subject:   userId.String(),
		Issuer:    "evently-test",
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
	}).SignedString(jwt.UnsafeAllowNoneSignatureType)

	badSubject, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Subject:   "not-a-uuid",
		Issuer:    "evently-test",
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
	}).SignedString([]byte(testSecret))

	noExpiry, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Subject: userId.String(),
		Issuer:  "evently-test",
	}).SignedString([]byte(testSecret))

	cases := map[string]string{
		"empty":        "",
		"garbage":      "not.a.jwt",
		"tampered":     valid[:len(valid)-2] + "xx",
		"expired":      expiredToken,
		"wrong secret": otherSecret,
		"wrong issuer": otherIssuer,
		"alg none":     noneToken,
		"bad subject":  badSubject,
		"missing exp":  noExpiry,
	}
	for name, token := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := tokens.VerifyAccessToken(token); !errors.Is(err, ErrInvalidToken) {
				t.Fatalf("expected ErrInvalidToken, got %v", err)
			}
		})
	}
}
