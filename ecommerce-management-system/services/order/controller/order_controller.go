package controller

import (
	"net/http"
	"strconv"

	"ecommerce-microservices/services/order/model"
	"ecommerce-microservices/services/order/repository"
	"ecommerce-microservices/services/order/service"

	"github.com/gin-gonic/gin"
)

type OrderController struct {
	service service.OrderService
}

func NewOrderController(s service.OrderService) *OrderController {
	return &OrderController{service: s}
}

func (c *OrderController) Create(ctx *gin.Context) {
	var req model.CreateOrderRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body: " + err.Error()})
		return
	}

	order := model.Order{
		CustomerID: req.CustomerID,
		ProductID:  req.ProductID,
		Quantity:   req.Quantity,
		Total:      req.Total,
		Status:     model.StatusCreated,
	}

	reqID, _ := ctx.Get("request_id")
	reqIDStr, _ := reqID.(string)

	if err := c.service.Create(ctx.Request.Context(), &order, reqIDStr); err != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, order)
}

func (c *OrderController) GetByID(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "order id must be numeric"})
		return
	}

	o, err := c.service.GetByID(ctx.Request.Context(), id)
	if repository.IsNotFound(err) {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	}
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, o)
}

func (c *OrderController) List(ctx *gin.Context) {
	orders, err := c.service.List(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if orders == nil {
		orders = []model.Order{}
	}
	ctx.JSON(http.StatusOK, orders)
}

func (c *OrderController) UpdateStatus(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "order id must be numeric"})
		return
	}

	var req model.UpdateStatusRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid status body"})
		return
	}

	updated, err := c.service.UpdateStatus(ctx.Request.Context(), id, req.Status)
	if repository.IsNotFound(err) {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	}
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, updated)
}
