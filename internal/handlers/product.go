package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"ecommerce-api/internal/models"
	"ecommerce-api/internal/repository"
	"ecommerce-api/internal/utils"
)

type ProductHandler struct {
	products repository.ProductRepository
}

func NewProductHandler(products repository.ProductRepository) *ProductHandler {
	return &ProductHandler{products: products}
}

func (h *ProductHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	page, _ := strconv.Atoi(q.Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(q.Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	params := models.ProductListParams{
		Page:      page,
		Limit:     limit,
		Search:    strings.TrimSpace(q.Get("search")),
		SortBy:    q.Get("sort_by"),
		SortOrder: q.Get("sort_order"),
	}

	if catStr := q.Get("category_id"); catStr != "" {
		catID, err := uuid.Parse(catStr)
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, "invalid category_id")
			return
		}
		params.CategoryID = &catID
	}

	products, total, err := h.products.List(r.Context(), params)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	totalPages := total / limit
	if total%limit != 0 {
		totalPages++
	}

	utils.WriteSuccess(w, http.StatusOK, models.PaginatedResponse{
		Items:      products,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	})
}

func (h *ProductHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid product id")
		return
	}

	product, err := h.products.GetByID(r.Context(), id)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if product == nil {
		utils.WriteError(w, http.StatusNotFound, "product not found")
		return
	}

	utils.WriteSuccess(w, http.StatusOK, product)
}

func (h *ProductHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		utils.WriteError(w, http.StatusBadRequest, "name is required")
		return
	}
	if req.Price < 0 {
		utils.WriteError(w, http.StatusBadRequest, "price must be non-negative")
		return
	}
	if req.Stock < 0 {
		utils.WriteError(w, http.StatusBadRequest, "stock must be non-negative")
		return
	}

	now := time.Now()
	product := &models.Product{
		ID:          uuid.New(),
		Name:        req.Name,
		Description: strings.TrimSpace(req.Description),
		Price:       req.Price,
		Stock:       req.Stock,
		CategoryID:  req.CategoryID,
		ImageURL:    strings.TrimSpace(req.ImageURL),
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := h.products.Create(r.Context(), product); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	utils.WriteSuccess(w, http.StatusCreated, product)
}

func (h *ProductHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid product id")
		return
	}

	product, err := h.products.GetByID(r.Context(), id)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if product == nil {
		utils.WriteError(w, http.StatusNotFound, "product not found")
		return
	}

	var req models.UpdateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name != nil {
		if n := strings.TrimSpace(*req.Name); n != "" {
			product.Name = n
		}
	}
	if req.Description != nil {
		product.Description = strings.TrimSpace(*req.Description)
	}
	if req.Price != nil {
		if *req.Price < 0 {
			utils.WriteError(w, http.StatusBadRequest, "price must be non-negative")
			return
		}
		product.Price = *req.Price
	}
	if req.Stock != nil {
		if *req.Stock < 0 {
			utils.WriteError(w, http.StatusBadRequest, "stock must be non-negative")
			return
		}
		product.Stock = *req.Stock
	}
	if req.CategoryID != nil {
		product.CategoryID = req.CategoryID
	}
	if req.ImageURL != nil {
		product.ImageURL = strings.TrimSpace(*req.ImageURL)
	}

	if err := h.products.Update(r.Context(), product); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	utils.WriteSuccess(w, http.StatusOK, product)
}

func (h *ProductHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid product id")
		return
	}

	if err := h.products.Delete(r.Context(), id); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	utils.WriteMessage(w, http.StatusOK, "product deleted")
}
