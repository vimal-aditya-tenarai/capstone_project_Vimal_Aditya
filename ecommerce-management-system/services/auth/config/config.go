package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port              string
	JWTSecret         string
	AdminUsername     string
	AdminPasswordHash string
	UserUsername      string
	UserPasswordHash  string
	JWTExpiryMinutes  int
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

	expiry, _ := strconv.Atoi(os.Getenv("JWT_EXPIRY_MINUTES"))
	if expiry <= 0 {
		expiry = 60
	}

	port := os.Getenv("AUTH_SERVICE_PORT")
	if port == "" {
		port = "8083"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "secret-key-vimal-aditya-12345"
	}

	adminUsername := os.Getenv("ADMIN_USERNAME")
	if adminUsername == "" {
		adminUsername = "admin"
	}
	adminHash := os.Getenv("ADMIN_PASSWORD_HASH")
	if adminHash == "" {
		adminHash = "240be518fabd2724ddb6f04eeb1da5967448d7e831c08c8fa822809f74c720a9"
	}

	userUsername := os.Getenv("USER_USERNAME")
	if userUsername == "" {
		userUsername = "user"
	}
	userHash := os.Getenv("USER_PASSWORD_HASH")
	if userHash == "" {
		userHash = "e606e38b0d8c19b24cf0ee3808183162ea7cd63ff7912dbb22b5e803286b4446"
	}

	cfg := Config{
		Port:              port,
		JWTSecret:         jwtSecret,
		AdminUsername:     adminUsername,
		AdminPasswordHash: adminHash,
		UserUsername:      userUsername,
		UserPasswordHash:  userHash,
		JWTExpiryMinutes:  expiry,
	}

	log.Printf("[CONFIG] Auth Service Config: Port=%s, Admin=%s, User=%s", cfg.Port, cfg.AdminUsername, cfg.UserUsername)
	return cfg
}
