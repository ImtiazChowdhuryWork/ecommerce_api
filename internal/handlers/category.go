package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"ecommerce-api/internal/models"
	"ecommerce-api/internal/repository"
	"ecommerce-api/internal/utils"
)

type CategoryHandler struct {
	categories repository.CategoryRepository
}

func NewCategoryHandler(categories repository.CategoryRepository) *CategoryHandler {
	return &CategoryHandler{categories: categories}
}

func (h *CategoryHandler) List(w http.ResponseWriter, r *http.Request) {
	cats, err := h.categories.List(r.Context())
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	utils.WriteSuccess(w, http.StatusOK, cats)
}

func (h *CategoryHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid category id")
		return
	}

	cat, err := h.categories.GetByID(r.Context(), id)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if cat == nil {
		utils.WriteError(w, http.StatusNotFound, "category not found")
		return
	}
	utils.WriteSuccess(w, http.StatusOK, cat)
}

func (h *CategoryHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CreateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		utils.WriteError(w, http.StatusBadRequest, "name is required")
		return
	}

	now := time.Now()
	cat := &models.Category{
		ID:          uuid.New(),
		Name:        req.Name,
		Description: strings.TrimSpace(req.Description),
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := h.categories.Create(r.Context(), cat); err != nil {
		if strings.Contains(err.Error(), "unique") {
			utils.WriteError(w, http.StatusConflict, "category name already exists")
			return
		}
		utils.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	utils.WriteSuccess(w, http.StatusCreated, cat)
}

func (h *CategoryHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid category id")
		return
	}

	cat, err := h.categories.GetByID(r.Context(), id)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if cat == nil {
		utils.WriteError(w, http.StatusNotFound, "category not found")
		return
	}

	var req models.UpdateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name != nil {
		if n := strings.TrimSpace(*req.Name); n != "" {
			cat.Name = n
		}
	}
	if req.Description != nil {
		cat.Description = strings.TrimSpace(*req.Description)
	}

	if err := h.categories.Update(r.Context(), cat); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	utils.WriteSuccess(w, http.StatusOK, cat)
}

func (h *CategoryHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid category id")
		return
	}

	if err := h.categories.Delete(r.Context(), id); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	utils.WriteMessage(w, http.StatusOK, "category deleted")
}
