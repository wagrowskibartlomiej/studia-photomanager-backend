package main

import (
	"database/sql"
	"os"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

func InitDB(cfg *Config) (*sql.DB, error) {
	photosDir := cfg.Photos.Directory
	if err := os.MkdirAll(photosDir, os.ModePerm); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite3", cfg.Database.File)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS users (
		ID INTEGER PRIMARY KEY AUTOINCREMENT,
		login TEXT UNIQUE,
		password TEXT,
		isAdmin INTEGER NOT NULL,
		isBanned INTEGER NOT NULL
	);
	`)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS photos (
		ID INTEGER PRIMARY KEY AUTOINCREMENT,
		imagePath TEXT,
		imageIsPublic INTEGER NOT NULL,
		userID INTEGER NOT NULL,
		FOREIGN KEY (userID) REFERENCES users(ID) ON DELETE CASCADE
	);
	`)
	if err != nil {
		return nil, err
	}

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count); err != nil {
		return nil, err
	}
	if count == 0 {
		hash, _ := bcrypt.GenerateFromPassword([]byte(cfg.Admin.DefaultPassword), bcrypt.DefaultCost)
		_, _ = db.Exec("INSERT INTO users (login, password, isAdmin, isBanned) VALUES (?, ?, ?, ?)",
			cfg.Admin.DefaultLogin, string(hash), 1, 0)
	}

	return db, nil
}

func FindUser(db *sql.DB, u *User) (DBUser, bool) {
	var dbU DBUser
	err := db.QueryRow("SELECT ID, password FROM users WHERE login = ?", u.Login).Scan(&dbU.ID, &dbU.Password)
	if err != nil {
		return DBUser{}, false
	}
	dbU.Login = u.Login
	return dbU, true
}

func LoginUser(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

var globalPasswordValidator PasswordValidator

func InitPasswordValidator(cfg *Config) {
	globalPasswordValidator = GetPasswordValidator(cfg)
}

func ValidatePassword(password string) error {
	if globalPasswordValidator == nil {
		// Fallback for easy
		globalPasswordValidator = &EasyValidator{}
	}
	return globalPasswordValidator.Validate(password)
}

func RegisterUser(db *sql.DB, u *User) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = db.Exec("INSERT INTO users (login, password, isAdmin, isBanned) VALUES (?, ?, ?, ?)", u.Login, string(hash), 0, 0)
	return err
}

