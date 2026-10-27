package main

import (
	"ecommerce-microservices/services/product/config"
	"ecommerce-microservices/services/product/controller"
	"ecommerce-microservices/services/product/database"
	"ecommerce-microservices/services/product/health"
	"ecommerce-microservices/services/product/logging"
	"ecommerce-microservices/services/product/middleware"
	"ecommerce-microservices/services/product/model"
	"ecommerce-microservices/services/product/repository"
	"ecommerce-microservices/services/product/service"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()
	logger := logging.New()

	db, err := database.NewConnection(cfg.DSN())
	if err != nil {
		log.Fatalf("product database connection error: %v", err)
	}
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	if err := db.AutoMigrate(&model.Product{}); err != nil {
		log.Fatalf("product auto-migration failed: %v", err)
	}

	repo := repository.NewPostgresProductRepository(db)
	svc := service.NewProductService(repo)
	ctrl := controller.NewProductController(svc)

	r := gin.New()

	// Observability & Recovery Middleware
	r.Use(middleware.RequestID())
	r.Use(middleware.Recovery(logger))
	r.Use(middleware.Logging(logger))

	// Health Endpoints
	r.GET("/health", health.Live)
	r.GET("/health/live", health.Live)
	r.GET("/health/ready", health.Ready(db))

	// Direct backwards compatibility routes for Gateway & GraphQL
	r.POST("/products", ctrl.Create)
	r.GET("/products/:id", ctrl.GetByID)
	r.GET("/products", ctrl.GetAll)

	// API v1 with JWT authentication
	api := r.Group("/api/v1")
	api.Use(middleware.JWTAuth(cfg.JWTSecret))
	{
		api.POST("/products", ctrl.Create)
		api.GET("/products/:id", ctrl.GetByID)
		api.GET("/products", ctrl.GetAll)
	}

	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	logger.Info("Product Service listening on http://localhost:" + cfg.Port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error(err.Error())
	}
}
