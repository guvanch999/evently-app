package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var ErrInvalidToken = errors.New("invalid or expired token")

// TokenService issues and verifies the application's own HS256 access tokens.
// Google tokens are only used during sign-in and are never handed to clients.
type TokenService struct {
	secret []byte
	issuer string
	ttl    time.Duration
	now    func() time.Time
}

func NewTokenService(secret, issuer string, ttl time.Duration) *TokenService {
	return &TokenService{
		secret: []byte(secret),
		issuer: issuer,
		ttl:    ttl,
		now:    time.Now,
	}
}

// TTL returns the lifetime of issued access tokens.
func (s *TokenService) TTL() time.Duration {
	return s.ttl
}

// IssueAccessToken returns a signed JWT whose subject is the user id.
func (s *TokenService) IssueAccessToken(userId uuid.UUID) (string, error) {
	now := s.now()
	claims := jwt.RegisteredClaims{
		Subject:   userId.String(),
		Issuer:    s.issuer,
		IssuedAt:  jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(s.ttl)),
		ID:        uuid.NewString(),
	}
	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.secret)
	if err != nil {
		return "", fmt.Errorf("sign access token: %w", err)
	}
	return signed, nil
}

// VerifyAccessToken validates signature, algorithm, issuer and expiry and
// returns the user id stored in the subject claim.
func (s *TokenService) VerifyAccessToken(tokenString string) (uuid.UUID, error) {
	claims := &jwt.RegisteredClaims{}
	_, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(*jwt.Token) (interface{}, error) { return s.secret, nil },
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithIssuer(s.issuer),
		jwt.WithExpirationRequired(),
		jwt.WithTimeFunc(s.now),
	)
	if err != nil {
		return uuid.Nil, ErrInvalidToken
	}
	userId, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.Nil, ErrInvalidToken
	}
	return userId, nil
}
