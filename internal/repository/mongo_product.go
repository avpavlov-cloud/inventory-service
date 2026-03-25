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
	slog.Info("Main: Calling NewRepo now...")
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

func (r *MongoProductRepository) GetWarehouseAnalytics(ctx context.Context) (map[string]interface{}, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: nil},
			{Key: "total_items", Value: bson.D{{Key: "$sum", Value: "$quantity"}}},
			{Key: "total_value", Value: bson.D{{Key: "$sum", Value: bson.D{
				{Key: "$multiply", Value: bson.A{"$price", "$quantity"}},
			}}}},
			{Key: "avg_price", Value: bson.D{{Key: "$avg", Value: "$price"}}},
			{Key: "max_price", Value: bson.D{{Key: "$max", Value: "$price"}}},
		}}},
		{{Key: "$project", Value: bson.D{
			{Key: "_id", Value: 0},
			{Key: "total_count", Value: "$total_items"},
			{Key: "inventory_value", Value: "$total_value"},
			{Key: "average_price", Value: "$avg_price"},
			{Key: "most_expensive", Value: "$max_price"},
		}}},
	}

	slog.Info("Main: Repo initialized", "pointer", r) // ДОЛЖНО БЫТЬ НЕ NIL

	// 1. Выполняем запрос
	cursor, err := r.collection.Aggregate(ctx, pipeline)

	// 2. КРИТИЧЕСКИ ВАЖНО: Если ошибка, выходим СРАЗУ.
	// Если err != nil, то cursor БУДЕТ nil, и обращение к нему вызовет панику.
	if err != nil {
		return nil, fmt.Errorf("aggregate failed: %w", err)
	}

	// 3. Теперь, когда мы уверены, что cursor не nil, вешаем закрытие
	defer cursor.Close(ctx)

	var results []map[string]interface{}
	// 4. Декодируем
	if err := cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("decode failed: %w", err)
	}

	// 5. Обработка пустого результата (если товаров 0)
	if len(results) == 0 {
		return map[string]interface{}{
			"total_count":     0,
			"inventory_value": 0,
			"status":          "empty",
		}, nil
	}

	return results[0], nil
}
