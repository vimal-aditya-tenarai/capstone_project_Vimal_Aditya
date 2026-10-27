package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"ecommerce-microservices/services/order/model"

	"github.com/sony/gobreaker"
)

type ProductClient interface {
	GetProductByID(ctx context.Context, id int64, requestID string) (*model.ProductDTO, error)
}

type productClient struct {
	baseURL string
	client  *http.Client
	cb      *gobreaker.CircuitBreaker
}

func NewProductClient(baseURL string) ProductClient {
	st := gobreaker.Settings{
		Name:        "product-service",
		MaxRequests: 1,
		Interval:    10 * time.Second,
		Timeout:     10 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= 3
		},
	}

	return &productClient{
		baseURL: baseURL,
		client:  &http.Client{Timeout: 1500 * time.Millisecond},
		cb:      gobreaker.NewCircuitBreaker(st),
	}
}

func (c *productClient) GetProductByID(ctx context.Context, id int64, requestID string) (*model.ProductDTO, error) {
	result, err := c.cb.Execute(func() (interface{}, error) {
		var lastErr error
		url := fmt.Sprintf("%s/products/%d", c.baseURL, id)

		for attempt := 1; attempt <= 3; attempt++ {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
			}

			req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
			if err != nil {
				return nil, err
			}
			if requestID != "" {
				req.Header.Set("X-Request-ID", requestID)
			}

			resp, err := c.client.Do(req)
			if err == nil {
				defer resp.Body.Close()
				if resp.StatusCode == http.StatusOK {
					var prod model.ProductDTO
					if err := json.NewDecoder(resp.Body).Decode(&prod); err != nil {
						return nil, err
					}
					return &prod, nil
				}
				lastErr = fmt.Errorf("product service returned HTTP %d", resp.StatusCode)
			} else {
				lastErr = err
			}

			backoff := time.Duration(100*(1<<(attempt-1))) * time.Millisecond
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}
		return nil, lastErr
	})

	if err != nil {
		return nil, fmt.Errorf("resilience failure: %w", err)
	}

	return result.(*model.ProductDTO), nil
}
