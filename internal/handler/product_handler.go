package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/avpavlov-cloud/inventory-service/internal/domain"
	"github.com/avpavlov-cloud/inventory-service/internal/service"
)

type ProductHandler struct {
	service service.ProductService
}

func NewProductHandler(s service.ProductService) *ProductHandler {
	return &ProductHandler{service: s}
}

func (h *ProductHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var p domain.Product
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.service.RegisterNewProduct(r.Context(), &p); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(p)
}

func (h *ProductHandler) GetBySKU(w http.ResponseWriter, r *http.Request) {
	sku := chi.URLParam(r, "sku")
	product, err := h.service.GetStockInfo(r.Context(), sku)
	if err != nil {
		http.Error(w, "product not found", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(product)
}

func (h *ProductHandler) TestTransaction(w http.ResponseWriter, r *http.Request) {
	// Для простоты возьмем параметры из URL: /test-tx?from=A&to=B&amount=5
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")

	err := h.service.TransferStock(r.Context(), from, to, 5)
	if err != nil {
		http.Error(w, "Transaction failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write([]byte("Transaction successful"))
}

// TestTransferStock — хендлер для проверки транзакции
func (h *ProductHandler) TestTransferStock(w http.ResponseWriter, r *http.Request) {
	// Читаем параметры из URL: ?from=apple&to=samsung&amount=5
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")
	amountStr := r.URL.Query().Get("amount")

	// В реальном проекте здесь была бы валидация чисел
	amount := 1
	if amountStr != "" {
		fmt.Sscanf(amountStr, "%d", &amount)
	}

	// Вызываем сервис, который мы обернули в транзакцию
	err := h.service.TransferStock(r.Context(), from, to, amount)
	if err != nil {
		// Если транзакция откатилась, мы увидим ошибку здесь
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Write([]byte("Transaction success: stock transferred"))
}

func (h *ProductHandler) GetProducts(w http.ResponseWriter, r *http.Request) {
	// Читаем параметры и конвертируем их (в реальном проекте лучше через вспомогательную функцию)
	query := r.URL.Query()

	limit, _ := strconv.ParseInt(query.Get("limit"), 10, 64)
	if limit == 0 {
		limit = 10
	} // Значение по умолчанию

	offset, _ := strconv.ParseInt(query.Get("offset"), 10, 64)
	minPrice, _ := strconv.ParseFloat(query.Get("minPrice"), 64)

	products, err := h.service.GetList(r.Context(), minPrice, limit, offset)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	json.NewEncoder(w).Encode(products)
}

func (h *ProductHandler) GetAnalytics(w http.ResponseWriter, r *http.Request) {
	stats, err := h.service.GetAnalytics(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(stats)
}

