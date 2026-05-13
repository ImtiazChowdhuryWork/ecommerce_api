package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"ecommerce-api/internal/middleware"
	"ecommerce-api/internal/models"
	"ecommerce-api/internal/repository"
	"ecommerce-api/internal/utils"
)

type UserHandler struct {
	users repository.UserRepository
}

func NewUserHandler(users repository.UserRepository) *UserHandler {
	return &UserHandler{users: users}
}

func (h *UserHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	user, err := h.users.GetByID(r.Context(), claims.UserID)
	if err != nil || user == nil {
		utils.WriteError(w, http.StatusNotFound, "user not found")
		return
	}
	utils.WriteSuccess(w, http.StatusOK, user)
}

func (h *UserHandler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)

	var req models.UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	if req.Name == "" && req.Email == "" {
		utils.WriteError(w, http.StatusBadRequest, "provide at least one field to update")
		return
	}

	user, err := h.users.GetByID(r.Context(), claims.UserID)
	if err != nil || user == nil {
		utils.WriteError(w, http.StatusNotFound, "user not found")
		return
	}

	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Email != "" && req.Email != user.Email {
		exists, err := h.users.EmailExists(r.Context(), req.Email)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, "internal server error")
			return
		}
		if exists {
			utils.WriteError(w, http.StatusConflict, "email already in use")
			return
		}
		user.Email = req.Email
	}

	if err := h.users.Update(r.Context(), user); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	utils.WriteSuccess(w, http.StatusOK, user)
}

func (h *UserHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)

	var req models.ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.CurrentPassword == "" || req.NewPassword == "" {
		utils.WriteError(w, http.StatusBadRequest, "current_password and new_password are required")
		return
	}
	if len(req.NewPassword) < 6 {
		utils.WriteError(w, http.StatusBadRequest, "new password must be at least 6 characters")
		return
	}

	user, err := h.users.GetByID(r.Context(), claims.UserID)
	if err != nil || user == nil {
		utils.WriteError(w, http.StatusNotFound, "user not found")
		return
	}

	if !utils.CheckPassword(req.CurrentPassword, user.PasswordHash) {
		utils.WriteError(w, http.StatusUnauthorized, "current password is incorrect")
		return
	}

	hash, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	if err := h.users.UpdatePassword(r.Context(), claims.UserID, hash); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	utils.WriteMessage(w, http.StatusOK, "password updated successfully")
}
