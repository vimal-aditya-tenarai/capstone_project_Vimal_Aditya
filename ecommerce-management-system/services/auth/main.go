package main

import (
	"ecommerce-microservices/services/auth/auth"
	"ecommerce-microservices/services/auth/config"
	"ecommerce-microservices/services/auth/controller"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	authService := auth.New(
		cfg.JWTSecret,
		cfg.JWTExpiryMinutes,
		cfg.AdminUsername,
		cfg.AdminPasswordHash,
		cfg.UserUsername,
		cfg.UserPasswordHash,
	)
	authCtrl := controller.NewAuthController(authService)

	r := gin.Default()

	// Public health endpoints for Consul and load balancers
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"service": "auth-service", "status": "UP"})
	})
	r.GET("/health/live", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"service": "auth-service", "status": "UP"})
	})

	// Login routes (accessible via both /login and /auth/login)
	r.POST("/login", authCtrl.Login)
	r.POST("/auth/login", authCtrl.Login)

	log.Println("Auth Service listening on http://localhost:" + cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Auth Service failed to start: %v", err)
	}
}
