package repository

import (
	"context"
	"ecommerce-microservices/services/order/model"
	"errors"

	"gorm.io/gorm"
)

type OrderRepository interface {
	Create(context.Context, *model.Order) error
	GetByID(context.Context, int64) (*model.Order, error)
	List(context.Context) ([]model.Order, error)
	UpdateStatus(context.Context, int64, model.OrderStatus) (*model.Order, error)
}

type PostgresOrderRepository struct {
	db *gorm.DB
}

func NewPostgresOrderRepository(db *gorm.DB) OrderRepository {
	return &PostgresOrderRepository{db: db}
}

func (r *PostgresOrderRepository) Create(ctx context.Context, o *model.Order) error {
	if o.Status == "" {
		o.Status = model.StatusCreated
	}
	if o.CustomerID == 0 {
		o.CustomerID = 101
	}
	return r.db.WithContext(ctx).Create(o).Error
}

func (r *PostgresOrderRepository) GetByID(ctx context.Context, id int64) (*model.Order, error) {
	var o model.Order
	if err := r.db.WithContext(ctx).First(&o, id).Error; err != nil {
		return nil, err
	}
	return &o, nil
}

func (r *PostgresOrderRepository) List(ctx context.Context) ([]model.Order, error) {
	var orders []model.Order
	err := r.db.WithContext(ctx).Order("id DESC").Find(&orders).Error
	return orders, err
}

func (r *PostgresOrderRepository) UpdateStatus(ctx context.Context, id int64, status model.OrderStatus) (*model.Order, error) {
	var o model.Order
	if err := r.db.WithContext(ctx).First(&o, id).Error; err != nil {
		return nil, err
	}
	o.Status = status
	if err := r.db.WithContext(ctx).Save(&o).Error; err != nil {
		return nil, err
	}
	return &o, nil
}

func IsNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
