package repository

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/avpavlov-cloud/inventory-service/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoProductRepository struct {
	collection *mongo.Collection
}

func NewMongoProductRepository(ctx context.Context, db *mongo.Database) (*MongoProductRepository, error) {
	col := db.Collection("products")

	// 1. Уникальный индекс на SKU (защита от дублей)
	_, err := col.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "sku", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create sku index: %w", err)
	}

	// 2. Индекс на Price (ускорение фильтрации и сортировки)
	_, err = col.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "price", Value: 1}},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create price index: %w", err)
	}

	return &MongoProductRepository{collection: col}, nil
}

func (r *MongoProductRepository) Create(ctx context.Context, p *domain.Product) error {
	p.ID = primitive.NewObjectID()
	p.UpdatedAt = time.Now()

	_, err := r.collection.InsertOne(ctx, p)
	return err
}

func (r *MongoProductRepository) GetBySKU(ctx context.Context, sku string) (*domain.Product, error) {
	var product domain.Product
	filter := bson.M{"sku": sku}

	err := r.collection.FindOne(ctx, filter).Decode(&product)
	if err != nil {
		return nil, err
	}
	return &product, nil
}

func (r *MongoProductRepository) UpdateQuantity(ctx context.Context, sku string, amount int) error {
	// Логируем начало операции
	slog.Info("DB UpdateQuantity start", "sku", sku, "amount", amount)

	filter := bson.M{"sku": sku}
	update := bson.M{
		"$inc": bson.M{"quantity": amount},
		"$set": bson.M{"updated_at": time.Now()},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		slog.Error("DB UpdateQuantity error", "error", err, "sku", sku)
		return err
	}

	// Логируем результат поиска
	slog.Info("DB UpdateQuantity result",
		"matched", result.MatchedCount,
		"modified", result.ModifiedCount,
		"sku", sku,
	)

	if result.MatchedCount == 0 {
		return fmt.Errorf("product %s not found", sku)
	}

	return nil
}

func (r *MongoProductRepository) GetList(ctx context.Context, minPrice float64, limit, offset int64) ([]domain.Product, error) {
	// 1. Создаем фильтр (например, товары дороже определенной цены)
	filter := bson.M{"price": bson.M{"$gte": minPrice}}

	// 2. Настраиваем опции пагинации
	findOptions := options.Find()
	findOptions.SetLimit(limit)
	findOptions.SetSkip(offset)
	findOptions.SetSort(bson.D{{Key: "price", Value: 1}}) // Сортировка по цене (возрастание)

	cursor, err := r.collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var products []domain.Product
	if err := cursor.All(ctx, &products); err != nil {
		return nil, err
	}

	return products, nil
}
