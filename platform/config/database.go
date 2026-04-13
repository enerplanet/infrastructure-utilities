package config

import "fmt"

// DatabaseConfig represents basic PostgreSQL credentials used by all services.
type DatabaseConfig struct {
	Host     string
	Port     string
	Name     string
	User     string
	Password string
}

// DSN renders a PostgreSQL connection string with UTC timezone.
func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable TimeZone=UTC",
		d.Host, d.Port, d.User, d.Password, d.Name)
}

// DatabaseFromEnv reads the standard DB_* variables and produces a config
// instance. The helper keeps defaults aligned across services.
func DatabaseFromEnv() DatabaseConfig {
	return DatabaseConfig{
		Host:     GetEnv("DB_HOST", "localhost"),
		Port:     GetEnv("DB_PORT", "5432"),
		Name:     GetEnv("DB_DATABASE", "spatialhub"),
		User:     GetEnv("DB_USERNAME", "postgres"),
		Password: GetEnv("DB_PASSWORD", "postgres"),
	}
}

// AppDatabaseFromEnv reads the APP_DB_* variables for the application database.
// Falls back to DB_* variables if APP_DB_* are not set.
func AppDatabaseFromEnv() DatabaseConfig {
	return DatabaseConfig{
		Host:     GetEnv("APP_DB_HOST", GetEnv("DB_HOST", "localhost")),
		Port:     GetEnv("APP_DB_PORT", GetEnv("DB_PORT", "5432")),
		Name:     GetEnv("APP_DB_DATABASE", "spatialai"),
		User:     GetEnv("APP_DB_USERNAME", GetEnv("DB_USERNAME", "postgres")),
		Password: GetEnv("APP_DB_PASSWORD", GetEnv("DB_PASSWORD", "postgres")),
	}
}
