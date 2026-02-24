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
	}
}
