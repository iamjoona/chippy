package auth

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hashed), nil
}

func CheckPasswordHash(hash, password string) error {
	hashSuccess := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))

	return hashSuccess
}

func GetBearerToken(headers http.Header) (string, error) {
	authHeader := headers.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("no authorization header provided")
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid auth header format: got %d parts, expected 2", len(parts))
	}
	if parts[0] != "Bearer" {
		return "", fmt.Errorf("invalid auth header format: got %s, expected Bearer", parts[0])
	}

	return parts[1], nil
}
