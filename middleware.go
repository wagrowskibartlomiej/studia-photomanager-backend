package main

import (
	"context"
	"database/sql"
	"net/http"
)

type contextKey string

const (
	ctxKeyLogin contextKey = "user_login"
	ctxKeyID    contextKey = "user_id"
)

func AuthMiddleware(cfg *Config, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenStr, err := getJWTFromCookie(r)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		claims, err := parseJWT(cfg, tokenStr)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		userIDFloat, ok := claims["user_id"].(float64)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		userLogin, ok := claims["user_login"].(string)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), ctxKeyID, int64(userIDFloat))
		ctx = context.WithValue(ctx, ctxKeyLogin, userLogin)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func AuthMiddlewareAdministration(cfg *Config, db *sql.DB, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("jwt")
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		claims, err := parseJWT(cfg, cookie.Value)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		userLogin, ok := claims["user_login"].(string)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var userID int64
		var isAdmin int
		var isBanned int
		err = db.QueryRow("SELECT ID, isAdmin, isBanned FROM users WHERE login = ?", userLogin).Scan(&userID, &isAdmin, &isBanned)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), ctxKeyID, userID)
		ctx = context.WithValue(ctx, ctxKeyLogin, userLogin)
		ctx = context.WithValue(ctx, "isAdmin", isAdmin != 0)
		ctx = context.WithValue(ctx, "isBanned", isBanned != 0)

		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

