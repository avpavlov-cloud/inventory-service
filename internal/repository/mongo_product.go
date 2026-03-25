package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/avpavlov-cloud/inventory-service/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoProductRepository struct {
	collection *mongo.Collection
}

// NewMongoProductRepository — конструктор репозитория
func NewMongoProductRepository(db *mongo.Database) *MongoProductRepository {
	return &MongoProductRepository{
		collection: db.Collection("products"),
	}
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
	// Ищем по строке SKU, а не по ObjectID
	filter := bson.M{"sku": sku}
	
	update := bson.M{
		"$inc": bson.M{"quantity": amount},
		"$set": bson.M{"updated_at": time.Now()},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("product with sku %s not found", sku)
	}
	return nil
}

