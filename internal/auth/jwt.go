package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Claims struct {
	EmailVerified bool `json:"email_verified"`
	jwt.RegisteredClaims
}

var ErrInvalidToken = errors.New("invalid token")

func IssueToken(secret []byte, userID uuid.UUID, emailVerified bool, ttl time.Duration) (string, error) {
	now := time.Now()
	Claims := Claims{
		EmailVerified: emailVerified,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			Issuer:    "frontdock",
			ID:        uuid.NewString(),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, Claims).SignedString(secret)
}

func ParseToken(secret []byte, tokenStr string) (*Claims, error) {
	claims := &Claims{}
	_, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return secret, nil
	}, jwt.WithIssuer("frontdock"), jwt.WithExpirationRequired())

	if err != nil {
		return nil, ErrInvalidToken
	}
	return claims, nil
}
