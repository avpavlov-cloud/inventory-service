package main

import (
	"log"
	"net/http"
	"os"

	"github.com/avpavlov-cloud/inventory-service/internal/handler"
	"github.com/avpavlov-cloud/inventory-service/internal/repository"
	"github.com/avpavlov-cloud/inventory-service/internal/service"
	"github.com/avpavlov-cloud/inventory-service/pkg/mongodb"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017" // дефолт для локального запуска
	}

	client, err := mongodb.NewClient(mongoURI)
	if err != nil {
		log.Fatal(err)
	}
	db := client.Database("inventory_db")

	//  Создаем Transaction Manager
	tm := repository.NewTransactionManager(client)

	// 2. Инициализация слоев
	repo := repository.NewMongoProductRepository(db)
	svc := service.NewProductService(repo, tm)
	h := handler.NewProductHandler(svc)

	// 3. Настройка роутера
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Post("/products", h.CreateProduct)
	r.Get("/products/{sku}", h.GetBySKU)

	r.Post("/products/transfer", h.TestTransferStock) 
	
	log.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal(err)
	}
}
