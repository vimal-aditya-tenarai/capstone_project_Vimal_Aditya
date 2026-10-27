package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port            string
	ConsulURL       string
	JWTSecret       string
	AuthServiceURL  string
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

	port := os.Getenv("GATEWAY_PORT")
	if port == "" {
		port = "8080"
	}
	consulURL := os.Getenv("CONSUL_URL")
	if consulURL == "" {
		consulURL = "http://localhost:8500"
	}
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "secret-key-vimal-aditya-12345"
	}
	authServiceURL := os.Getenv("AUTH_SERVICE_URL")
	if authServiceURL == "" {
		authServiceURL = "http://localhost:8083"
	}

	cfg := Config{
		Port:           port,
		ConsulURL:      consulURL,
		JWTSecret:      jwtSecret,
		AuthServiceURL: authServiceURL,
	}

	log.Printf("[CONFIG] Gateway Config: Port=%s, Consul=%s, AuthURL=%s", cfg.Port, cfg.ConsulURL, cfg.AuthServiceURL)
	return cfg
}
