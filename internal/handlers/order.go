package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"ecommerce-api/internal/middleware"
	"ecommerce-api/internal/models"
	"ecommerce-api/internal/repository"
	"ecommerce-api/internal/utils"
)

type OrderHandler struct {
	orders repository.OrderRepository
	carts  repository.CartRepository
}

func NewOrderHandler(orders repository.OrderRepository, carts repository.CartRepository) *OrderHandler {
	return &OrderHandler{orders: orders, carts: carts}
}

func (h *OrderHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)

	var req models.CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.ShippingAddress = strings.TrimSpace(req.ShippingAddress)
	if req.ShippingAddress == "" {
		utils.WriteError(w, http.StatusBadRequest, "shipping_address is required")
		return
	}

	cart, err := h.carts.GetOrCreateByUserID(r.Context(), claims.UserID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if len(cart.Items) == 0 {
		utils.WriteError(w, http.StatusBadRequest, "your cart is empty")
		return
	}

	order, err := h.orders.Create(r.Context(), claims.UserID, &req, cart)
	if err != nil {
		if strings.Contains(err.Error(), "insufficient stock") || strings.Contains(err.Error(), "empty") {
			utils.WriteError(w, http.StatusBadRequest, err.Error())
			return
		}
		utils.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	utils.WriteSuccess(w, http.StatusCreated, order)
}

func (h *OrderHandler) List(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	var (
		orders []*models.Order
		total  int
		err    error
	)

	if claims.Role == models.RoleAdmin {
		orders, total, err = h.orders.ListAll(r.Context(), page, limit)
	} else {
		orders, total, err = h.orders.ListByUser(r.Context(), claims.UserID, page, limit)
	}
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	totalPages := total / limit
	if total%limit != 0 {
		totalPages++
	}

	utils.WriteSuccess(w, http.StatusOK, models.PaginatedResponse{
		Items: orders, Total: total, Page: page, Limit: limit, TotalPages: totalPages,
	})
}

func (h *OrderHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid order id")
		return
	}

	order, err := h.orders.GetByID(r.Context(), id)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if order == nil {
		utils.WriteError(w, http.StatusNotFound, "order not found")
		return
	}

	// Non-admins can only see their own orders
	if claims.Role != models.RoleAdmin && order.UserID != claims.UserID {
		utils.WriteError(w, http.StatusNotFound, "order not found")
		return
	}

	utils.WriteSuccess(w, http.StatusOK, order)
}

func (h *OrderHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid order id")
		return
	}

	var req models.UpdateOrderStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	validStatuses := map[models.OrderStatus]bool{
		models.OrderStatusPending:   true,
		models.OrderStatusConfirmed: true,
		models.OrderStatusShipped:   true,
		models.OrderStatusDelivered: true,
		models.OrderStatusCancelled: true,
	}
	if !validStatuses[req.Status] {
		utils.WriteError(w, http.StatusBadRequest, "invalid status value")
		return
	}

	if err := h.orders.UpdateStatus(r.Context(), id, req.Status); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	utils.WriteMessage(w, http.StatusOK, "order status updated")
}

func (h *OrderHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid order id")
		return
	}

	isAdmin := claims.Role == models.RoleAdmin
	if err := h.orders.Cancel(r.Context(), id, claims.UserID, isAdmin); err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.WriteError(w, http.StatusNotFound, err.Error())
			return
		}
		if strings.Contains(err.Error(), "cannot cancel") {
			utils.WriteError(w, http.StatusBadRequest, err.Error())
			return
		}
		utils.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	utils.WriteMessage(w, http.StatusOK, "order cancelled and stock restored")
}
