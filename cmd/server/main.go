package main

import (
	"log"
	"net/http"

	"github.com/avpavlov-cloud/inventory-service/internal/handler"
	"github.com/avpavlov-cloud/inventory-service/internal/repository"
	"github.com/avpavlov-cloud/inventory-service/internal/service"
	"github.com/avpavlov-cloud/inventory-service/pkg/mongodb"
	"github.com/go-chi/chi/v5"
)

func main() {
	// 1. Подключение к Mongo
	client, err := mongodb.NewClient("mongodb://localhost:27017")
	if err != nil {
		log.Fatal(err)
	}
	db := client.Database("inventory_db")

	// 2. Инициализация слоев
	repo := repository.NewMongoProductRepository(db)
	svc := service.NewProductService(repo)
	h := handler.NewProductHandler(svc)

	// 3. Настройка роутера
	r := chi.NewRouter()
	r.Post("/products", h.CreateProduct)
	r.Get("/products/{sku}", h.GetBySKU)

	log.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal(err)
	}
}
