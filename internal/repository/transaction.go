package repository

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
)

type TransactionManager struct {
	client *mongo.Client
}

func NewTransactionManager(client *mongo.Client) *TransactionManager {
	return &TransactionManager{client: client}
}

// Execute обеспечивает атомарность: либо всё успешно, либо всё откатывается
func (tm *TransactionManager) Execute(ctx context.Context, fn func(mongo.SessionContext) error) error {
	session, err := tm.client.StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(ctx)

	// Включаем режим транзакции
	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		return nil, fn(sessCtx)
	})
	return err
}
