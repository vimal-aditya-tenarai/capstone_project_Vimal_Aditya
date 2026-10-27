package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string
	Port       string
	LogLevel   string
	JWTSecret  string
}

func Load() Config {
	loaded := false
	for _, f := range []string{".env", "../.env", "../../.env"} {
		if err := godotenv.Overload(f); err == nil {
			log.Printf("[CONFIG] Successfully loaded environment from: %s", f)
			loaded = true
			break
		}
	}
	if !loaded {
		log.Println("[CONFIG] Notice: No .env file found in search paths, using system env / defaults")
	}

	cfg := Config{
		DBHost:     get("PRODUCT_DB_HOST", "localhost"),
		DBPort:     get("PRODUCT_DB_PORT", "5432"),
		DBUser:     get("PRODUCT_DB_USER", "postgres"),
		DBPassword: get("PRODUCT_DB_PASSWORD", "postgres"),
		DBName:     get("PRODUCT_DB_NAME", "ecommerce_db"),
		DBSSLMode:  get("PRODUCT_DB_SSLMODE", "require"),
		Port:       get("PRODUCT_SERVICE_PORT", "8081"),
		LogLevel:   get("LOG_LEVEL", "info"),
		JWTSecret:  get("JWT_SECRET", "secret-key-vimal-aditya-12345"),
	}

	log.Printf("[CONFIG] Product Service Config: Port=%s, DB=%s@%s:%s/%s (sslmode=%s)", cfg.Port, cfg.DBUser, cfg.DBHost, cfg.DBPort, cfg.DBName, cfg.DBSSLMode)
	return cfg
}

func get(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func (c Config) DSN() string {
	if c.DBSSLMode != "" {
		return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName, c.DBSSLMode)
	}
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s", c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName)
}
