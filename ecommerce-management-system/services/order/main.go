package main

import (
	"ecommerce-microservices/services/order/client"
	"ecommerce-microservices/services/order/config"
	"ecommerce-microservices/services/order/controller"
	"ecommerce-microservices/services/order/database"
	"ecommerce-microservices/services/order/health"
	"ecommerce-microservices/services/order/logging"
	"ecommerce-microservices/services/order/middleware"
	"ecommerce-microservices/services/order/model"
	"ecommerce-microservices/services/order/repository"
	"ecommerce-microservices/services/order/service"
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
		log.Fatalf("order database connection error: %v", err)
	}
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	if err := db.AutoMigrate(&model.Order{}); err != nil {
		log.Fatalf("order auto-migration failed: %v", err)
	}

	repo := repository.NewPostgresOrderRepository(db)
	productClient := client.NewProductClient(cfg.ProductServiceURL)
	svc := service.NewOrderService(repo, productClient)
	ctrl := controller.NewOrderController(svc)

	r := gin.New()

	// Observability & Recovery Middleware
	r.Use(middleware.RequestID())
	r.Use(middleware.Recovery(logger))
	r.Use(middleware.Logging(logger))

	// Health Endpoints
	r.GET("/health", health.Live)
	r.GET("/health/live", health.Live)
	r.GET("/health/ready", health.Ready(db))

	// Direct backwards compatibility for Gateway & GraphQL
	r.POST("/orders", ctrl.Create)
	r.GET("/orders/:id", ctrl.GetByID)

	// API v1 with JWT authentication
	api := r.Group("/api/v1")
	api.Use(middleware.JWTAuth(cfg.JWTSecret))
	{
		api.POST("/orders", ctrl.Create)
		api.GET("/orders/:id", ctrl.GetByID)
		api.GET("/orders", ctrl.List)
		api.PUT("/orders/:id/status", ctrl.UpdateStatus)
	}

	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	logger.Info("Order Service listening on http://localhost:" + cfg.Port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error(err.Error())
	}
}
