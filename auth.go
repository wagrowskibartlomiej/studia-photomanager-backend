package main

import (
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func GenerateJWT(cfg *Config, userID int64, userLogin string) (string, error) {
	claims := jwt.MapClaims{
		"user_id":    userID,
		"user_login": userLogin,
		"exp":        time.Now().Add(cfg.JWT.Timeout()).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.JWT.SecretKey))
}

func getJWTFromCookie(r *http.Request) (string, error) {
	c, err := r.Cookie("jwt")
	if err != nil {
		return "", err
	}
	return c.Value, nil
}

func parseJWT(cfg *Config, tokenStr string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (any, error) {
		return []byte(cfg.JWT.SecretKey), nil
	})
	if err != nil || !token.Valid {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, jwt.ErrInvalidKey
	}

	return claims, nil
}

