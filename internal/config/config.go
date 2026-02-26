package config

import "os"

type Config struct {
	DataBaseURL string
	ServerPort  string
	JWTSecret   string
	SMTPHost    string
	SMTPPort    string
	SMTPUser    string
	SMTPPass    string
	SMTPFrom    string
	BaseURL     string
}

func Load() *Config {
	return &Config{
		DataBaseURL: os.Getenv("DATABASE_URL"),
		ServerPort:  os.Getenv("SERVER_PORT"),
		JWTSecret:   os.Getenv("JWT_SECRET"),
		SMTPHost:    os.Getenv("SMTP_HOST"),
		SMTPPort:    os.Getenv("SMTP_PORT"),
		SMTPUser:    os.Getenv("SMTP_USER"),
		SMTPPass:    os.Getenv("SMTP_PASS"),
		SMTPFrom:    os.Getenv("SMTP_FROM"),
		BaseURL:     os.Getenv("BASE_URL"),
	}
}
