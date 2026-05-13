package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"ecommerce-api/internal/middleware"
	"ecommerce-api/internal/models"
	"ecommerce-api/internal/repository"
	"ecommerce-api/internal/utils"
)

type CartHandler struct {
	carts    repository.CartRepository
	products repository.ProductRepository
}

func NewCartHandler(carts repository.CartRepository, products repository.ProductRepository) *CartHandler {
	return &CartHandler{carts: carts, products: products}
}

func (h *CartHandler) GetCart(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	cart, err := h.carts.GetOrCreateByUserID(r.Context(), claims.UserID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	utils.WriteSuccess(w, http.StatusOK, cart)
}

func (h *CartHandler) AddItem(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)

	var req models.AddCartItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.ProductID == uuid.Nil {
		utils.WriteError(w, http.StatusBadRequest, "product_id is required")
		return
	}
	if req.Quantity < 1 {
		utils.WriteError(w, http.StatusBadRequest, "quantity must be at least 1")
		return
	}

	// Verify product exists and has stock
	product, err := h.products.GetByID(r.Context(), req.ProductID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if product == nil {
		utils.WriteError(w, http.StatusNotFound, "product not found")
		return
	}
	if product.Stock < req.Quantity {
		utils.WriteError(w, http.StatusBadRequest, "insufficient stock")
		return
	}

	cart, err := h.carts.GetOrCreateByUserID(r.Context(), claims.UserID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	item, err := h.carts.AddItem(r.Context(), cart.ID, req.ProductID, req.Quantity)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	// Return the full cart so the client has the updated state
	updatedCart, err := h.carts.GetOrCreateByUserID(r.Context(), claims.UserID)
	if err != nil {
		utils.WriteSuccess(w, http.StatusOK, item)
		return
	}

	utils.WriteSuccess(w, http.StatusOK, updatedCart)
}

func (h *CartHandler) UpdateItem(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)

	itemID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid item id")
		return
	}

	var req models.UpdateCartItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Quantity < 1 {
		utils.WriteError(w, http.StatusBadRequest, "quantity must be at least 1")
		return
	}

	// Ensure item belongs to this user's cart
	existingItem, err := h.carts.GetItemByID(r.Context(), itemID)
	if err != nil || existingItem == nil {
		utils.WriteError(w, http.StatusNotFound, "cart item not found")
		return
	}

	cart, err := h.carts.GetOrCreateByUserID(r.Context(), claims.UserID)
	if err != nil || cart.ID != existingItem.CartID {
		utils.WriteError(w, http.StatusNotFound, "cart item not found")
		return
	}

	if _, err := h.carts.UpdateItem(r.Context(), itemID, req.Quantity); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	updatedCart, err := h.carts.GetOrCreateByUserID(r.Context(), claims.UserID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	utils.WriteSuccess(w, http.StatusOK, updatedCart)
}

func (h *CartHandler) RemoveItem(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)

	itemID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid item id")
		return
	}

	cart, err := h.carts.GetOrCreateByUserID(r.Context(), claims.UserID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	if err := h.carts.RemoveItem(r.Context(), itemID, cart.ID); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	updatedCart, err := h.carts.GetOrCreateByUserID(r.Context(), claims.UserID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	utils.WriteSuccess(w, http.StatusOK, updatedCart)
}

func (h *CartHandler) ClearCart(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)

	cart, err := h.carts.GetOrCreateByUserID(r.Context(), claims.UserID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	if err := h.carts.ClearItems(r.Context(), cart.ID); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	utils.WriteMessage(w, http.StatusOK, "cart cleared")
}
