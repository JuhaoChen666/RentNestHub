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
	AIThinking    string
	AIReasoning   string
	RedisAddr     string
	RedisPassword string
	SMTPHost      string
	SMTPPort      string
	SMTPUsername  string
	SMTPPassword  string
	MailFrom      string
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
		AIThinking:    envOrDefault("AI_THINKING", "disabled"),
		AIReasoning:   envOrDefault("AI_REASONING_EFFORT", "low"),
		RedisAddr:     envOrDefault("REDIS_ADDR", "localhost:6379"),
		RedisPassword: os.Getenv("REDIS_PASSWORD"),
		SMTPHost:      envOrDefault("SMTP_HOST", "localhost"),
		SMTPPort:      envOrDefault("SMTP_PORT", "1025"),
		SMTPUsername:  os.Getenv("SMTP_USERNAME"),
		SMTPPassword:  os.Getenv("SMTP_PASSWORD"),
		MailFrom:      envOrDefault("MAIL_FROM", "RentNestHub <no-reply@rentnesthub.local>"),
	}
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
