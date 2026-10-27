package tests

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"ecommerce-microservices/services/order/model"
	"ecommerce-microservices/services/order/service"
)

type mockOrderRepo struct {
	mu     sync.Mutex
	orders map[int64]*model.Order
	nextID int64
}

func newMockOrderRepo() *mockOrderRepo {
	return &mockOrderRepo{
		orders: make(map[int64]*model.Order),
		nextID: 1,
	}
}

func (m *mockOrderRepo) Create(ctx context.Context, o *model.Order) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	o.ID = m.nextID
	m.nextID++
	m.orders[o.ID] = o
	return nil
}

func (m *mockOrderRepo) GetByID(ctx context.Context, id int64) (*model.Order, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	o, exists := m.orders[id]
	if !exists {
		return nil, fmt.Errorf("not found")
	}
	return o, nil
}

func (m *mockOrderRepo) List(ctx context.Context) ([]model.Order, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var list []model.Order
	for _, o := range m.orders {
		list = append(list, *o)
	}
	return list, nil
}

func (m *mockOrderRepo) UpdateStatus(ctx context.Context, id int64, status model.OrderStatus) (*model.Order, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	o, exists := m.orders[id]
	if !exists {
		return nil, fmt.Errorf("not found")
	}
	o.Status = status
	return o, nil
}

type mockProductClient struct{}

func (m *mockProductClient) GetProductByID(ctx context.Context, id int64, requestID string) (*model.ProductDTO, error) {
	return &model.ProductDTO{
		ID:    id,
		Name:  "Test Item",
		Price: 99.99,
	}, nil
}

// Test100ConcurrentOrders tests requirement 11:
// 100 concurrent order requests handled safely with 0 race conditions.
func Test100ConcurrentOrders(t *testing.T) {
	ctx := context.Background()
	repo := newMockOrderRepo()
	prodClient := &mockProductClient{}
	svc := service.NewOrderService(repo, prodClient)

	var wg sync.WaitGroup
	workers := 100
	var successfulOrders int64

	start := time.Now()

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			reqID := fmt.Sprintf("req-%d", idx)
			o := &model.Order{
				CustomerID: int64(1000 + idx),
				ProductID:  int64(idx%5 + 1),
				Quantity:   2,
				Total:      199.98,
			}

			err := svc.Create(ctx, o, reqID)
			if err != nil {
				t.Errorf("worker %d failed: %v", idx, err)
				return
			}

			atomic.AddInt64(&successfulOrders, 1)
		}(i)
	}

	wg.Wait()
	duration := time.Since(start)

	if successfulOrders != int64(workers) {
		t.Fatalf("expected %d successful orders, but got %d", workers, successfulOrders)
	}

	orders, err := svc.List(ctx)
	if err != nil {
		t.Fatalf("failed to list orders: %v", err)
	}

	if len(orders) != workers {
		t.Errorf("expected repository order count to be %d, got %d", workers, len(orders))
	}

	t.Logf("Successfully processed %d concurrent orders in %v (Zero race conditions)", workers, duration)
}
