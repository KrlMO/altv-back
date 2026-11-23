package config

import (
	"errors"
	"os"
)

type Config struct {
	DataBaseURL   string
	ServerPort    string
	JWTSecret     string
	SMTPHost      string
	SMTPPort      string
	SMTPUser      string
	SMTPPass      string
	SMTPFrom      string
	BaseURL       string
	FrontendURL   string
	CORSOrigin    string
	StorageDriver string
	MediaURL      string
	MediaRoot     string
}

func Load() (*Config, error) {
	cfg := &Config{
		DataBaseURL:   os.Getenv("DATABASE_URL"),
		ServerPort:    os.Getenv("SERVER_PORT"),
		JWTSecret:     os.Getenv("JWT_SECRET"),
		SMTPHost:      os.Getenv("SMTP_HOST"),
		SMTPPort:      os.Getenv("SMTP_PORT"),
		SMTPUser:      os.Getenv("SMTP_USER"),
		SMTPPass:      os.Getenv("SMTP_PASS"),
		SMTPFrom:      os.Getenv("SMTP_FROM"),
		BaseURL:       os.Getenv("BASE_URL"),
		FrontendURL:   os.Getenv("FRONTEND_URL"),
		CORSOrigin:    os.Getenv("CORS_ORIGIN"),
		StorageDriver: os.Getenv("STORAGE_DRIVER"),
		MediaURL:      os.Getenv("MEDIA_URL"),
		MediaRoot:     os.Getenv("MEDIA_ROOT"),
	}

	if cfg.DataBaseURL == "" {
		return nil, errors.New("DATABASE_URL is required")
	}
	if cfg.JWTSecret == "" {
		return nil, errors.New("JWT_SECRET is required")
	}
	if cfg.ServerPort == "" {
		cfg.ServerPort = "8080"
	}
	if cfg.StorageDriver == "" {
		cfg.StorageDriver = "local"
	}

	return cfg, nil
}
