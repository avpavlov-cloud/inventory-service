package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/avpavlov-cloud/inventory-service/internal/domain"
	"github.com/avpavlov-cloud/inventory-service/internal/repository"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	ErrProductAlreadyExists = errors.New("product with this SKU already exists")
	ErrInvalidQuantity      = errors.New("quantity cannot be negative")
)

type ProductUseCase struct {
	repo repository.ProductRepository
	tm   *repository.TransactionManager // <-- Добавь это поле
}

// Обнови конструктор (New...), чтобы он принимал tm
func NewProductService(repo repository.ProductRepository, tm *repository.TransactionManager) *ProductUseCase {
	return &ProductUseCase{
		repo: repo,
		tm:   tm, // <-- Инициализируй поле
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

func (s *ProductUseCase) SellProduct(ctx context.Context, sku string, quantity int) error {
	// Используем TransactionManager
	return s.tm.Execute(ctx, func(sessCtx mongo.SessionContext) error {

		// 1. Уменьшаем остаток (используем sessCtx!)
		err := s.repo.UpdateQuantity(sessCtx, sku, -quantity)
		if err != nil {
			return err
		}

		// 2. Создаем запись о продаже (другой репозиторий, тот же sessCtx)
		// err = s.salesRepo.Create(sessCtx, saleData)

		return nil // Если вернули nil, Mongo сделает COMMIT (сохранит всё)
	})
}

func (s *ProductUseCase) TransferStock(ctx context.Context, fromSKU, toSKU string, amount int) error {
	return s.tm.Execute(ctx, func(sessCtx mongo.SessionContext) error {
		// 1. Уменьшаем у первого (успешно)
		if err := s.repo.UpdateQuantity(sessCtx, fromSKU, -amount); err != nil {
			return err
		}

		// --- ИМИТАЦИЯ ОШИБКИ ДЛЯ ПРОВЕРКИ ---
		// Представь, что здесь произошел сбой сети или проверка не прошла
		if toSKU == "bad-sku" {
			return errors.New("SUDDEN ERROR: rollback should happen now")
		}

		// 2. Увеличиваем у второго
		if err := s.repo.UpdateQuantity(sessCtx, toSKU, amount); err != nil {
			return err
		}

		return nil
	})
}

func (s *ProductUseCase) GetList(ctx context.Context, minPrice float64, limit, offset int64) ([]domain.Product, error) {
	// Промышленный стандарт: защищаем сервер от слишком больших запросов
	if limit <= 0 {
		limit = 10 // Значение по умолчанию
	}
	if limit > 100 {
		limit = 100 // Ограничиваем максимум (защита от DDOS/перегрузки памяти)
	}

	if offset < 0 {
		offset = 0
	}

	// Вызываем репозиторий
	products, err := s.repo.GetList(ctx, minPrice, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("service.GetList: %w", err)
	}

	return products, nil
}

func (s *ProductUseCase) GetAnalytics(ctx context.Context) (map[string]interface{}, error) {
	// Репозиторий ТЕПЕРЬ возвращает (map[string]interface{}, error)
	stats, err := s.repo.GetWarehouseAnalytics(ctx)
	if err != nil {
		return nil, fmt.Errorf("service.GetAnalytics: %w", err)
	}

	// Больше не нужно брать results[0], так как stats — это уже и есть та самая карта
	return stats, nil
}
