package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/avpavlov-cloud/inventory-service/docs"
	"github.com/avpavlov-cloud/inventory-service/internal/handler"
	"github.com/avpavlov-cloud/inventory-service/internal/repository"
	"github.com/avpavlov-cloud/inventory-service/internal/service"
	"github.com/avpavlov-cloud/inventory-service/pkg/mongodb"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

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
	repo, err := repository.NewMongoProductRepository(ctx, db)
	if err != nil {
		log.Fatalf("CRITICAL: Repo failed to start: %v", err)
	}
	svc := service.NewProductService(repo, tm)
	h := handler.NewProductHandler(svc)

	// 3. Настройка роутера
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route("/products", func(r chi.Router) {
		// Конкретные пути внутри /products/
		r.Get("/transfer", h.TestTransferStock) // путь: GET /products/transfer

		r.Get("/analytics", h.GetAnalytics)
		// Операции со списком и создание
		r.Get("/", h.GetProducts)    // путь: GET /products (ВАЖНО: тут просто "/")
		r.Post("/", h.CreateProduct) // путь: POST /products

		// Параметризованный путь (всегда в конце группы)
		r.Get("/{sku}", h.GetBySKU) // путь: GET /products/apple-15
	})

	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:8080/swagger/doc.json"),
	))

	log.Println("!!! Server VERSION 3 !!!")
	log.Println("!!! SERVER STARTED !!!", "port", "8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal(err)
	}
}
