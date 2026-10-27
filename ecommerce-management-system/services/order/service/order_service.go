package service

import (
	"context"
	"ecommerce-microservices/services/order/client"
	"ecommerce-microservices/services/order/model"
	"ecommerce-microservices/services/order/repository"
	"errors"
	"sync"
)

type OrderService interface {
	Create(context.Context, *model.Order, string) error
	GetByID(context.Context, int64) (*model.Order, error)
	List(context.Context) ([]model.Order, error)
	UpdateStatus(context.Context, int64, model.OrderStatus) (*model.Order, error)
}

type orderService struct {
	repo          repository.OrderRepository
	productClient client.ProductClient
	mu            sync.Mutex
}

func NewOrderService(repo repository.OrderRepository, productClient client.ProductClient) OrderService {
	return &orderService{
		repo:          repo,
		productClient: productClient,
	}
}

func (s *orderService) Create(ctx context.Context, o *model.Order, requestID string) error {
	if o.ProductID <= 0 {
		return errors.New("productId must be greater than zero")
	}
	if o.Quantity <= 0 {
		return errors.New("quantity must be greater than zero")
	}
	if o.CustomerID <= 0 {
		o.CustomerID = 101
	}

	if s.productClient != nil {
		prod, err := s.productClient.GetProductByID(ctx, o.ProductID, requestID)
		if err != nil {
			if o.Total <= 0 {
				return errors.New("product service unavailable or circuit open: " + err.Error())
			}
		} else if prod != nil && prod.Price > 0 {
			o.Total = float64(o.Quantity) * prod.Price
		}
	}

	if o.Total <= 0 {
		return errors.New("total must be greater than zero")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	return s.repo.Create(ctx, o)
}

func (s *orderService) GetByID(ctx context.Context, id int64) (*model.Order, error) {
	if id <= 0 {
		return nil, errors.New("order id must be greater than zero")
	}
	return s.repo.GetByID(ctx, id)
}

func (s *orderService) List(ctx context.Context) ([]model.Order, error) {
	return s.repo.List(ctx)
}

func (s *orderService) UpdateStatus(ctx context.Context, id int64, status model.OrderStatus) (*model.Order, error) {
	if id <= 0 {
		return nil, errors.New("order id must be greater than zero")
	}
	switch status {
	case model.StatusCreated, model.StatusConfirmed, model.StatusCancelled:
	default:
		return nil, errors.New("invalid status: must be CREATED, CONFIRMED, or CANCELLED")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	return s.repo.UpdateStatus(ctx, id, status)
}
