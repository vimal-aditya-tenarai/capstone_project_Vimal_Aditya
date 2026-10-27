package middleware

import (
	"ecommerce-microservices/services/order/logging"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

func Logging(logger *logging.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		reqID, _ := c.Get("request_id")
		if reqID == nil {
			reqID = "-"
		}

		logger.Info(fmt.Sprintf(
			"request_id=%v method=%s path=%s status=%d duration=%s",
			reqID,
			c.Request.Method,
			c.Request.URL.RequestURI(),
			c.Writer.Status(),
			time.Since(start),
		))
	}
}
