package main

import (
	"io"
	"log"
	"net/http"
	"strings"

	"ecommerce-microservices/gateway/client"
	"ecommerce-microservices/gateway/config"
	gqlserver "ecommerce-microservices/gateway/graphql"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// createProxy creates a reverse proxy dynamically resolving via Consul or fallback URL
func createProxy(consulClient *client.ConsulClient, serviceName, fallbackBaseURL, prefixToRemove string) gin.HandlerFunc {
	return func(c *gin.Context) {
		targetBaseURL := fallbackBaseURL
		if consulClient != nil && serviceName != "" {
			if u, err := consulClient.GetServiceURL(serviceName); err == nil && u != "" {
				targetBaseURL = u
			}
		}

		targetPath := strings.TrimPrefix(c.Request.URL.Path, prefixToRemove)
		if targetPath == "" {
			targetPath = "/"
		}
		targetURL := targetBaseURL + targetPath
		if c.Request.URL.RawQuery != "" {
			targetURL += "?" + c.Request.URL.RawQuery
		}

		req, err := http.NewRequestWithContext(c.Request.Context(), c.Request.Method, targetURL, c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
			return
		}

		req.Header = c.Request.Header.Clone()
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "downstream service unavailable: " + err.Error()})
			return
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), body)
	}
}

func jwtMiddleware(secret []byte) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing Bearer token"})
			return
		}

		tokenString := strings.TrimPrefix(header, "Bearer ")
		token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
			if t.Method != jwt.SigningMethodHS256 {
				return nil, jwt.ErrSignatureInvalid
			}
			return secret, nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}
		c.Next()
	}
}

func main() {
	cfg := config.Load()
	consulClient := client.NewConsulClient(cfg.ConsulURL)

	gs, err := gqlserver.NewServer(consulClient)
	if err != nil {
		log.Fatalf("GraphQL schema failed: %v", err)
	}

	r := gin.Default()

	// 1. Gateway Health Endpoints
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"service": "api-gateway", "status": "UP"})
	})
	r.GET("/health/live", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"service": "api-gateway", "status": "UP"})
	})

	authServiceURL := cfg.AuthServiceURL
	productServiceURL := "http://localhost:8081"
	orderServiceURL := "http://localhost:8082"

	// 2. Auth Routes forwarded to Auth Service (Consul: "auth-service" / fallback :8083)
	// Specific exact routes - no wildcard collision with /auth/*path
	r.POST("/login", createProxy(consulClient, "auth-service", authServiceURL, ""))
	r.POST("/auth/login", createProxy(consulClient, "auth-service", authServiceURL, "/auth"))

	// 3. GraphQL Endpoints
	h := gqlserver.NewHTTPHandler(gs)
	r.Any("/graphql", gin.WrapH(h))
	r.Any("/playground", gin.WrapH(h))

	// 4. Direct REST Proxy endpoints
	r.Any("/products", createProxy(consulClient, "product-service", productServiceURL, "/products"))
	r.Any("/products/*path", createProxy(consulClient, "product-service", productServiceURL, "/products"))
	r.Any("/orders", createProxy(consulClient, "order-service", orderServiceURL, "/orders"))
	r.Any("/orders/*path", createProxy(consulClient, "order-service", orderServiceURL, "/orders"))

	// 5. API Gateway routes with JWT protection from JWT-Gateway-with3services
	protected := r.Group("/api", jwtMiddleware([]byte(cfg.JWTSecret)))
	{
		protected.Any("/v1/orders", createProxy(consulClient, "order-service", orderServiceURL, ""))
		protected.Any("/v1/orders/*path", createProxy(consulClient, "order-service", orderServiceURL, ""))
		protected.Any("/v1/products", createProxy(consulClient, "product-service", productServiceURL, ""))
		protected.Any("/v1/products/*path", createProxy(consulClient, "product-service", productServiceURL, ""))
	}

	log.Println("API Gateway & Navigation Proxy running at http://localhost:" + cfg.Port)
	log.Println("GraphQL UI running at http://localhost:" + cfg.Port + "/playground")
	log.Fatal(r.Run(":" + cfg.Port))
}
