package config

import "os"

type Config struct {
	ServerPort    string
	DatabaseDSN   string
	UploadDir     string
	PublicBaseURL string
	AIAPIURL      string
	AIAPIKey      string
	AIModel       string
}

func Load() Config {
	return Config{
		ServerPort:    envOrDefault("SERVER_PORT", "8081"),
		DatabaseDSN:   envOrDefault("DATABASE_DSN", "rentnest:change-me@tcp(localhost:3306)/rentnesthub?parseTime=true&charset=utf8mb4"),
		UploadDir:     envOrDefault("UPLOAD_DIR", "./uploads"),
		PublicBaseURL: envOrDefault("PUBLIC_BASE_URL", "http://localhost:8080"),
		AIAPIURL:      os.Getenv("AI_API_URL"),
		AIAPIKey:      os.Getenv("AI_API_KEY"),
		AIModel:       os.Getenv("AI_MODEL"),
	}
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
