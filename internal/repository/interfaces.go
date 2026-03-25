package repository

import (
	"context"

	"github.com/avpavlov-cloud/inventory-service/internal/domain"
)

type ProductRepository interface {
	Create(ctx context.Context, product *domain.Product) error
	GetBySKU(ctx context.Context, sku string) (*domain.Product, error)
	UpdateQuantity(ctx context.Context, id string, amount int) error
}
