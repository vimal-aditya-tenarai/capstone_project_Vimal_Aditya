package controller

import (
	"ecommerce-microservices/services/auth/auth"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthController struct {
	auth *auth.AuthService
}

func NewAuthController(a *auth.AuthService) *AuthController {
	return &AuthController{auth: a}
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (c *AuthController) Login(ctx *gin.Context) {
	var req LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON request body"})
		return
	}

	token, role, err := c.auth.Login(req.Username, req.Password)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"token":     token,
		"tokenType": "Bearer",
		"username":  req.Username,
		"role":      role,
	})
}
