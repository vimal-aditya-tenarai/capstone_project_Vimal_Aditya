package middleware

import (
	"ecommerce-microservices/services/product/logging"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Recovery(logger *logging.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				errStr := fmt.Sprintf("panic recovered: %v", r)
				logger.Error(errStr)
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error": "internal server error",
				})
			}
		}()
		c.Next()
	}
}
