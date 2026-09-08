package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"time"

	"ecommerce-api/internal/middleware"
	"ecommerce-api/internal/models"
	"ecommerce-api/internal/repository"
	"ecommerce-api/internal/utils"
)

type ReviewHandler struct {
	reviews  repository.ReviewRepository
	products repository.ProductRepository
}

func NewReviewHandler(reviews repository.ReviewRepository, products repository.ProductRepository) *ReviewHandler {
	return &ReviewHandler{reviews: reviews, products: products}
}

func (h *ReviewHandler) ListByProduct(w http.ResponseWriter, r *http.Request) {
	productID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid product id")
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	reviews, total, err := h.reviews.ListByProduct(r.Context(), productID, page, limit)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	totalPages := total / limit
	if total%limit != 0 {
		totalPages++
	}

	utils.WriteSuccess(w, http.StatusOK, models.PaginatedResponse{
		Items: reviews, Total: total, Page: page, Limit: limit, TotalPages: totalPages,
	})
}

func (h *ReviewHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)

	productID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid product id")
		return
	}

	product, err := h.products.GetByID(r.Context(), productID)
	if err != nil || product == nil {
		utils.WriteError(w, http.StatusNotFound, "product not found")
		return
	}

	existing, err := h.reviews.GetByUserAndProduct(r.Context(), claims.UserID, productID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if existing != nil {
		utils.WriteError(w, http.StatusConflict, "you have already reviewed this product")
		return
	}

	var req models.CreateReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Rating < 1 || req.Rating > 5 {
		utils.WriteError(w, http.StatusBadRequest, "rating must be between 1 and 5")
		return
	}

	now := time.Now()
	review := &models.Review{
		ID:        uuid.New(),
		UserID:    claims.UserID,
		ProductID: productID,
		Rating:    req.Rating,
		Comment:   req.Comment,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := h.reviews.Create(r.Context(), review); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	utils.WriteSuccess(w, http.StatusCreated, review)
}

func (h *ReviewHandler) Update(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)

	reviewID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid review id")
		return
	}

	review, err := h.reviews.GetByID(r.Context(), reviewID)
	if err != nil || review == nil {
		utils.WriteError(w, http.StatusNotFound, "review not found")
		return
	}

	if review.UserID != claims.UserID && claims.Role != models.RoleAdmin {
		utils.WriteError(w, http.StatusForbidden, "not allowed to edit this review")
		return
	}

	var req models.UpdateReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Rating != nil {
		if *req.Rating < 1 || *req.Rating > 5 {
			utils.WriteError(w, http.StatusBadRequest, "rating must be between 1 and 5")
			return
		}
		review.Rating = *req.Rating
	}
	if req.Comment != nil {
		review.Comment = *req.Comment
	}

	if err := h.reviews.Update(r.Context(), review); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	utils.WriteSuccess(w, http.StatusOK, review)
}

func (h *ReviewHandler) Delete(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)

	reviewID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid review id")
		return
	}

	isAdmin := claims.Role == models.RoleAdmin
	if err := h.reviews.Delete(r.Context(), reviewID, claims.UserID, isAdmin); err != nil {
		if err.Error() == "review not found or not owned by user" {
			utils.WriteError(w, http.StatusNotFound, "review not found")
			return
		}
		utils.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	utils.WriteMessage(w, http.StatusOK, "review deleted")
}
