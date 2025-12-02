package main

import (
	"encoding/json"
	"os"
	"time"
)

type Config struct {
	Server   ServerConfig   `json:"server"`
	Database DatabaseConfig `json:"database"`
	JWT      JWTConfig      `json:"jwt"`
	Photos   PhotosConfig   `json:"photos"`
	Admin    AdminConfig    `json:"admin"`
	Password *PasswordConfig `json:"password,omitempty"`
}

type ServerConfig struct {
	Port string `json:"port"`
	Host string `json:"host"`
}

type DatabaseConfig struct {
	File string `json:"file"`
}

type JWTConfig struct {
	SecretKey      string `json:"secret_key"`
	TimeoutMinutes int    `json:"timeout_minutes"`
}

func (j JWTConfig) Timeout() time.Duration {
	return time.Duration(j.TimeoutMinutes) * time.Minute
}

type PhotosConfig struct {
	Directory string `json:"directory"`
}

type AdminConfig struct {
	DefaultLogin    string `json:"default_login"`
	DefaultPassword string `json:"default_password"`
}

type PasswordConfig struct {
	Mode   string         `json:"mode"` // no-validation, easy, medium, restrict, custom
	Custom *CustomValidator `json:"custom,omitempty"`
}

var appConfig *Config

func LoadConfig(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var cfg Config
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&cfg); err != nil {
		return nil, err
	}

	appConfig = &cfg
	return &cfg, nil
}

func GetConfig() *Config {
	return appConfig
}

