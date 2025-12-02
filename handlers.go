package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func HandleLogin(cfg *Config, db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var user User
		if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		dbU, found := FindUser(db, &user)
		if !found || !LoginUser(dbU.Password, user.Password) {
			http.Error(w, "Invalid login or password", http.StatusUnauthorized)
			return
		}

		token, _ := GenerateJWT(cfg, dbU.ID, dbU.Login)
		http.SetCookie(w, &http.Cookie{
			Name:     "jwt",
			Value:    token,
			Path:     "/",
			HttpOnly: true,
			Secure:   false,
			SameSite: http.SameSiteLaxMode,
		})

		var isAdmin int
		_ = db.QueryRow("SELECT isAdmin FROM users WHERE ID = ?", dbU.ID).Scan(&isAdmin)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"status":  "ok",
			"isAdmin": isAdmin != 0,
		})
	}
}

func HandleRegister(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var user User
		if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		if err := ValidatePassword(user.Password); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := RegisterUser(db, &user); err != nil {
			http.Error(w, "Registration failed", http.StatusBadRequest)
			return
		}

		w.Write([]byte(`{"status":"ok"}`))
	}
}

func HandleAddPhoto(cfg *Config, db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		userID := r.Context().Value(ctxKeyID).(int64)
		userLogin := r.Context().Value(ctxKeyLogin).(string)

		publicStr := r.FormValue("public")
		imageIsPublic := 0
		if publicStr == "1" {
			imageIsPublic = 1
		}

		file, header, err := r.FormFile("photo")
		if err != nil {
			http.Error(w, "Failed to read photo", http.StatusBadRequest)
			return
		}
		defer file.Close()

		userDir := fmt.Sprintf("%s/%s", cfg.Photos.Directory, userLogin)
		os.MkdirAll(userDir, os.ModePerm)

		filename := fmt.Sprintf("%s/%s", userDir, header.Filename)
		out, _ := os.Create(filename)
		defer out.Close()
		io.Copy(out, file)

		db.Exec("INSERT INTO photos (imagePath, imageIsPublic, userID) VALUES (?, ?, ?)", filename, imageIsPublic, userID)

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"message": "Photo uploaded"})
	}
}

func HandleGetPhotos(cfg *Config, db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) < 4 {
			http.Error(w, "Invalid URL", http.StatusBadRequest)
			return
		}

		userLogin := parts[3]
		var filename string
		if len(parts) >= 5 {
			filename = parts[4]
		}

		var authorized bool
		if cookie, err := r.Cookie("jwt"); err == nil {
			if claims, err := parseJWT(cfg, cookie.Value); err == nil {
				authorized = claims["user_login"] == userLogin
			}
		}

		if filename != "" {
			var imagePath string
			var imageIsPublic int
			err := db.QueryRow(`SELECT imagePath, imageIsPublic FROM photos p JOIN users u ON p.userID=u.ID WHERE u.login=? AND p.imagePath LIKE ?`, userLogin, "%/"+filename).Scan(&imagePath, &imageIsPublic)
			if err != nil || (imageIsPublic == 0 && !authorized) {
				http.Error(w, "Forbidden or not found", http.StatusForbidden)
				return
			}
			http.ServeFile(w, r, imagePath)
			return
		}

		rows, err := db.Query(`SELECT p.imagePath, p.imageIsPublic FROM photos p JOIN users u ON p.userID=u.ID WHERE u.login=?`, userLogin)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode([]Photo{})
			return
		}
		defer rows.Close()

		var photos []Photo
		for rows.Next() {
			var imagePath string
			var imageIsPublic int
			rows.Scan(&imagePath, &imageIsPublic)
			if imageIsPublic != 0 || authorized {
				photos = append(photos, Photo{Filename: filepath.Base(imagePath), Public: imageIsPublic != 0})
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(photos)
	}
}

func HandleDeletePhoto(cfg *Config, db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		parts := strings.Split(r.URL.Path, "/")
		if len(parts) < 5 {
			http.Error(w, "Invalid URL", http.StatusBadRequest)
			return
		}

		userLogin := parts[3]
		filename := parts[4]
		userID := r.Context().Value(ctxKeyID).(int64)

		var dbUserID int64
		err := db.QueryRow(`SELECT u.ID FROM users u JOIN photos p ON p.userID=u.ID WHERE u.login=? AND p.imagePath=?`, userLogin, fmt.Sprintf("%s/%s/%s", cfg.Photos.Directory, userLogin, filename)).Scan(&dbUserID)
		if err != nil || dbUserID != userID {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		os.Remove(fmt.Sprintf("%s/%s/%s", cfg.Photos.Directory, userLogin, filename))
		db.Exec(`DELETE FROM photos WHERE imagePath=?`, fmt.Sprintf("%s/%s/%s", cfg.Photos.Directory, userLogin, filename))

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "Photo deleted"})
	}
}

func HandleTogglePhotoPublic(cfg *Config, db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		userID := r.Context().Value(ctxKeyID).(int64)
		userLogin := r.Context().Value(ctxKeyLogin).(string)

		var req UpdatePublicRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		imagePath := fmt.Sprintf("%s/%s/%s", cfg.Photos.Directory, userLogin, req.Filename)

		var dbUserID int64
		err := db.QueryRow(`SELECT userID FROM photos WHERE userID=? AND imagePath=?`, userID, imagePath).Scan(&dbUserID)
		if err != nil || dbUserID != userID {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		_, err = db.Exec(`UPDATE photos SET imageIsPublic=? WHERE userID=? AND imagePath=?`, req.Public, userID, imagePath)
		if err != nil {
			http.Error(w, "Failed to update photo", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Photo public state updated",
			"public":  fmt.Sprintf("%d", req.Public),
		})
	}
}

func HandlePublicGallery(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query(`
			SELECT u.login, p.imagePath 
			FROM photos p 
			JOIN users u ON p.userID = u.ID 
			WHERE p.imageIsPublic = 1 AND u.isBanned = 0
			ORDER BY p.ID DESC
		`)
		if err != nil {
			http.Error(w, "DB error", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var list []PublicPhoto

		for rows.Next() {
			var login string
			var path string
			rows.Scan(&login, &path)

			list = append(list, PublicPhoto{
				User:     login,
				Filename: filepath.Base(path),
			})
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(list)
	}
}

func HandleGetUsers(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query(`SELECT login, isAdmin, isBanned FROM users ORDER BY login`)
		if err != nil {
			http.Error(w, "DB error", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var users []UserResponse
		for rows.Next() {
			var login string
			var isAdmin, isBanned int
			rows.Scan(&login, &isAdmin, &isBanned)
			if isAdmin == 0 {
				users = append(users, UserResponse{
					Login:    login,
					IsBanned: isBanned != 0,
				})
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(users)
	}
}

func HandleManageBanStatus(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		isAdmin := r.Context().Value("isAdmin").(bool)
		if !isAdmin {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		var req ManageBanRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		userLogin := r.Context().Value(ctxKeyLogin).(string)
		if req.Login == userLogin {
			http.Error(w, "Cannot ban yourself", http.StatusForbidden)
			return
		}

		_, err := db.Exec("UPDATE users SET isBanned = ? WHERE login = ?", req.Banned, req.Login)
		if err != nil {
			http.Error(w, "Failed to update ban status", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"login":   req.Login,
			"banned":  fmt.Sprintf("%d", req.Banned),
			"message": "Ban status updated",
		})
	}
}

