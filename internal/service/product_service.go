package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/avpavlov-cloud/inventory-service/internal/domain"
	"github.com/avpavlov-cloud/inventory-service/internal/repository"
)

var (
	ErrProductAlreadyExists = errors.New("product with this SKU already exists")
	ErrInvalidQuantity      = errors.New("quantity cannot be negative")
)

type ProductUseCase struct {
	repo repository.ProductRepository
}

func NewProductService(repo repository.ProductRepository) *ProductUseCase {
	return &ProductUseCase{
		repo: repo,
	}
}

func (s *ProductUseCase) RegisterNewProduct(ctx context.Context, p *domain.Product) error {
	// 1. Бизнес-валидация
	if p.Quantity < 0 {
		return ErrInvalidQuantity
	}

	// 2. Проверка уникальности SKU
	existing, _ := s.repo.GetBySKU(ctx, p.SKU)
	if existing != nil {
		return ErrProductAlreadyExists
	}

	// 3. Сохранение
	if err := s.repo.Create(ctx, p); err != nil {
		return fmt.Errorf("service.RegisterNewProduct: %w", err)
	}

	return nil
}

func (s *ProductUseCase) GetStockInfo(ctx context.Context, sku string) (*domain.Product, error) {
	return s.repo.GetBySKU(ctx, sku)
}
