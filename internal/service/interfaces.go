package service

import (
	"context"

	"github.com/avpavlov-cloud/inventory-service/internal/domain"
)

type ProductService interface {
	RegisterNewProduct(ctx context.Context, p *domain.Product) error
	GetStockInfo(ctx context.Context, sku string) (*domain.Product, error)
	TransferStock(ctx context.Context, fromSKU, toSKU string, amount int) error
	GetList(ctx context.Context, minPrice float64, limit, offset int64) ([]domain.Product, error)
}
