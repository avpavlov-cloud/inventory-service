package service

import (
	"context"

	"github.com/avpavlov-cloud/inventory-service/internal/domain"
)

type ProductService interface {
	RegisterNewProduct(ctx context.Context, p *domain.Product) error
	GetStockInfo(ctx context.Context, sku string) (*domain.Product, error)
}
