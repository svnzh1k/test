package jwtService

import (
	"fmt"
	"os"

	"github.com/dgrijalva/jwt-go"
)

var secretKey string

func SetSecret() {
	// Read the content of the file
	content, err := os.ReadFile("env.txt")
	if err != nil {
		fmt.Println("Error reading secret file:", err)
		os.Exit(1)
	}

	// Set the secret key
	secretKey = string(content)
}

func GenerateJWT(claims jwt.Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secretKey))
}

func ValidateJWT(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, fmt.Errorf("token is invalid")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("failed to parse claims")
	}
	return claims, nil
}
