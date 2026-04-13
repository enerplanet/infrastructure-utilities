package config

import (
	"fmt"
	"os"

	"platform.local/platform/auth"

	goredis "github.com/redis/go-redis/v9"
)

// EmailSettings captures the SMTP parameters shared across services.
type EmailSettings struct {
	SMTPHost      string
	SMTPPort      int
	SMTPUsername  string
	SMTPPassword  string
	SMTPFromName  string
	SMTPFromEmail string
	SMTPUseTLS    bool
}

// AuthConfigFromEnv builds an auth.Config from the standard Keycloak env vars.
func AuthConfigFromEnv() auth.Config {
	return auth.Config{
		BaseURL:      os.Getenv("KEYCLOAK_URL"),
		ClientID:     os.Getenv("KEYCLOAK_CLIENT_ID"),
		RedirectURL:  os.Getenv("REDIRECT_URL"),
		ClientSecret: os.Getenv("KEYCLOAK_CLIENT_SECRET"),
		Realm:        os.Getenv("KEYCLOAK_REALM"),
	}
}

// RedisOptionsFromEnv returns Redis options derived from the shared env naming.
func RedisOptionsFromEnv(db int) goredis.Options {
	return goredis.Options{
		Addr:     fmt.Sprintf("%s:%s", os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT")),
		Username: os.Getenv("REDIS_USERNAME"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       db,
	}
}

// EmailSettingsFromEnv reads SMTP configuration once so every service can share
func EmailSettingsFromEnv() EmailSettings {
	smtpPort, err := GetEnvInt("SMTP_PORT", 587)
	if err != nil {
		smtpPort = 587
	}

	return EmailSettings{
		SMTPHost:      os.Getenv("SMTP_HOST"),
		SMTPPort:      smtpPort,
		SMTPUsername:  os.Getenv("SMTP_USERNAME"),
		SMTPPassword:  os.Getenv("SMTP_PASSWORD"),
		SMTPFromName:  os.Getenv("SMTP_FROM_NAME"),
		SMTPFromEmail: os.Getenv("SMTP_FROM_EMAIL"),
		SMTPUseTLS:    GetEnvBool("SMTP_USE_TLS", true),
	}
}
