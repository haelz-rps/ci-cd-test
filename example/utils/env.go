package utils

import (
	"fmt"
	"os"
)

func SetTestEnv() {
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_USER", "security-bulletins")
	os.Setenv("DB_PASSWORD", "pass")
	os.Setenv("DB_NAME", "security-bulletins")
	os.Setenv("DB_PORT", "5432")
	os.Setenv("DB_SSLMODE", "disable")
}

func BuildDBConnectionString() string {
	dbHost := os.Getenv("DB_HOST")
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbPort := os.Getenv("DB_PORT")
	dbSSLMode := os.Getenv("DB_SSLMODE")

	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		dbHost,
		dbUser,
		dbPass,
		dbName,
		dbPort,
		dbSSLMode,
	)
}

func EnvString(env, fallback string) string {
	e := os.Getenv(env)
	if e == "" {
		return fallback
	}
	return e
}
