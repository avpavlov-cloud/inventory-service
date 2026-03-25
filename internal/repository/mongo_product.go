package repository

import (
	"context"
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

// Добавьте этот метод к остальным методам структуры MongoProductRepository
func (r *MongoProductRepository) UpdateQuantity(ctx context.Context, id string, amount int) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	filter := bson.M{"_id": objectID}
	update := bson.M{
		"$inc": bson.M{"quantity": amount},
		"$set": bson.M{"updated_at": time.Now()},
	}

	_, err = r.collection.UpdateOne(ctx, filter, update)
	return err
}
