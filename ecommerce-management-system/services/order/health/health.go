package health

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func Live(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"service": "order-service",
		"status":  "UP",
	})
}

func Ready(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		sqlDB, err := db.DB()
		if err != nil || sqlDB.PingContext(c.Request.Context()) != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"service":  "order-service",
				"status":   "DOWN",
				"database": "DOWN",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"service":  "order-service",
			"status":   "READY",
			"database": "UP",
		})
	}
}
