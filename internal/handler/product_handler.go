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

// CreateProduct godoc
// @Summary      Создать новый товар
// @Description  Добавляет новый товар в базу данных склада
// @Tags         products
// @Accept       json
// @Produce      json
// @Param        product  body      domain.Product  true  "Данные товара"
// @Success      201      {object}  domain.Product
// @Failure      400      {string}  string "invalid request body"
// @Failure      500      {string}  string "internal server error"
// @Router       /products [post]
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

// GetBySKU godoc
// @Summary      Получить товар по SKU
// @Description  Возвращает полную информацию о товаре по его уникальному артикулу
// @Tags         products
// @Produce      json
// @Param        sku   path      string  true  "Артикул товара (SKU)"
// @Success      200   {object}  domain.Product
// @Failure      404   {string}  string "product not found"
// @Router       /products/{sku} [get]
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

// TestTransferStock godoc
// @Summary      Перевод остатков (Транзакция)
// @Description  Переводит количество товара с одного SKU на другой внутри атомарной транзакции
// @Tags         inventory
// @Produce      json
// @Param        from    query     string  true  "SKU отправителя"
// @Param        to      query     string  true  "SKU получателя"
// @Param        amount  query     int     false "Количество (по умолчанию 1)"
// @Success      200     {string}  string "Transaction success"
// @Failure      500     {object}  map[string]string "error description"
// @Router       /products/transfer [get]
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

// GetProducts godoc
// @Summary      Список товаров с пагинацией
// @Description  Возвращает список товаров с фильтрацией по цене и поддержкой страниц
// @Tags         products
// @Produce      json
// @Param        limit     query     int    false  "Лимит (по умолчанию 10)"
// @Param        offset    query     int    false  "Сдвиг (по умолчанию 0)"
// @Param        minPrice  query     number  false  "Минимальная цена"
// @Success      200       {array}   domain.Product
// @Failure      500       {string}  string "internal server error"
// @Router       /products [get]
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

// GetAnalytics godoc
// @Summary      Аналитика склада
// @Description  Возвращает агрегированную статистику: общее количество, стоимость и среднюю цену
// @Tags         analytics
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Failure      500  {string}  string "internal server error"
// @Router       /products/analytics [get]
func (h *ProductHandler) GetAnalytics(w http.ResponseWriter, r *http.Request) {
	stats, err := h.service.GetAnalytics(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(stats)
}

