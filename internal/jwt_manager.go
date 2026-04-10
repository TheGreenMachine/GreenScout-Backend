package internal

// Utilities for managing JWT bearer tokens

import (
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"io"
	"os"
	"strings"
	"time"
)

type accessTokenClaims struct {
	UUID     string `json:"uuid"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

func jwtSecret() ([]byte, error) {
	privFile, openErr := os.Open(RSAPrivateKeyPath)
	if openErr != nil {
		LogErrorf(openErr, "Problem opening %v", RSAPrivateKeyPath)
		return nil, openErr
	}

	defer privFile.Close()

	keyBytes, readErr := io.ReadAll(privFile)
	if readErr != nil {
		LogErrorf(readErr, "Problem reading %v", RSAPrivateKeyPath)
		return nil, readErr
	}

	if len(keyBytes) < 32 {
		return nil, errors.New("JWT SECRET too short (use 32+ chars)")
	}
	return keyBytes, nil
}

func mintAccessToken(uuid, username, role string, ttl time.Duration) (string, error) {
	secret, err := jwtSecret()
	if err != nil {
		return "", err
	}

	now := time.Now()
	claims := accessTokenClaims{
		UUID:     uuid,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   uuid,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}

	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return tok.SignedString(secret)
}

func parseBearerToken(authz string) (string, bool) {
	authz = strings.TrimSpace(authz)
	if authz == "" {
		return "", false
	}
	const pfx = "Bearer "
	if !strings.HasPrefix(authz, pfx) {
		return "", false
	}
	return strings.TrimSpace(authz[len(pfx):]), true
}

func verifyAccessToken(tokenString string) (*accessTokenClaims, error) {
	secret, err := jwtSecret()
	if err != nil {
		return nil, err
	}

	var claims accessTokenClaims
	tok, err := jwt.ParseWithClaims(tokenString, &claims, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return secret, nil
	})
	if err != nil {
		return nil, err
	}
	if !tok.Valid {
		return nil, errors.New("invalid token")
	}
	return &claims, nil
}
