package repository

import (
	"context"

	"github.com/avpavlov-cloud/inventory-service/internal/domain"
)

type ProductRepository interface {
	Create(ctx context.Context, product *domain.Product) error
	GetBySKU(ctx context.Context, sku string) (*domain.Product, error)
	UpdateQuantity(ctx context.Context, id string, amount int) error
	GetList(ctx context.Context, minPrice float64, limit, offset int64) ([]domain.Product, error)
	GetWarehouseAnalytics(ctx context.Context) (map[string]interface{}, error) 
}
