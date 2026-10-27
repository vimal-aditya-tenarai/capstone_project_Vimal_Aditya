package model

type OrderStatus string

const (
	StatusCreated   OrderStatus = "CREATED"
	StatusConfirmed OrderStatus = "CONFIRMED"
	StatusCancelled OrderStatus = "CANCELLED"
)

type Order struct {
	ID         int64       `gorm:"primaryKey" json:"id"`
	CustomerID int64       `gorm:"not null" json:"customerId"`
	ProductID  int64       `gorm:"not null" json:"productId"`
	Quantity   int         `gorm:"not null" json:"quantity"`
	Total      float64     `gorm:"not null" json:"total"`
	Status     OrderStatus `gorm:"size:20;not null;default:'CREATED'" json:"status"`
}

type CreateOrderRequest struct {
	CustomerID int64   `json:"customerId"`
	ProductID  int64   `json:"productId" binding:"required"`
	Quantity   int     `json:"quantity" binding:"required,min=1"`
	Total      float64 `json:"total"`
}

type UpdateStatusRequest struct {
	Status OrderStatus `json:"status" binding:"required"`
}

type ProductDTO struct {
	ID    int64   `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}
