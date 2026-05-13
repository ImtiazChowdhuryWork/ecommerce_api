package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"ecommerce-api/internal/config"
	"ecommerce-api/internal/models"
	"ecommerce-api/internal/repository"
	"ecommerce-api/internal/utils"
)

type AuthHandler struct {
	users repository.UserRepository
	cfg   *config.Config
}

func NewAuthHandler(users repository.UserRepository, cfg *config.Config) *AuthHandler {
	return &AuthHandler{users: users, cfg: cfg}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	req.Name = strings.TrimSpace(req.Name)

	if req.Email == "" || req.Password == "" || req.Name == "" {
		utils.WriteError(w, http.StatusBadRequest, "email, password, and name are required")
		return
	}
	if len(req.Password) < 6 {
		utils.WriteError(w, http.StatusBadRequest, "password must be at least 6 characters")
		return
	}

	exists, err := h.users.EmailExists(r.Context(), req.Email)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if exists {
		utils.WriteError(w, http.StatusConflict, "email already registered")
		return
	}

	hash, err := utils.HashPassword(req.Password)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	now := time.Now()
	user := &models.User{
		ID:           uuid.New(),
		Email:        req.Email,
		PasswordHash: hash,
		Name:         req.Name,
		Role:         models.RoleCustomer,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := h.users.Create(r.Context(), user); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	tokens, err := h.issueTokens(user)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	utils.WriteSuccess(w, http.StatusCreated, tokens)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	user, err := h.users.GetByEmail(r.Context(), req.Email)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if user == nil || !utils.CheckPassword(req.Password, user.PasswordHash) {
		utils.WriteError(w, http.StatusUnauthorized, "invalid email or password")
		return
	}

	tokens, err := h.issueTokens(user)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	utils.WriteSuccess(w, http.StatusOK, tokens)
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req models.RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	claims, err := utils.ValidateToken(req.RefreshToken, h.cfg.RefreshSecret)
	if err != nil || claims.Type != "refresh" {
		utils.WriteError(w, http.StatusUnauthorized, "invalid or expired refresh token")
		return
	}

	user, err := h.users.GetByID(r.Context(), claims.UserID)
	if err != nil || user == nil {
		utils.WriteError(w, http.StatusUnauthorized, "user not found")
		return
	}

	accessToken, err := utils.GenerateAccessToken(user, h.cfg.JWTSecret, h.cfg.JWTExpiry)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	utils.WriteSuccess(w, http.StatusOK, map[string]string{"access_token": accessToken})
}

func (h *AuthHandler) issueTokens(user *models.User) (*models.AuthResponse, error) {
	access, err := utils.GenerateAccessToken(user, h.cfg.JWTSecret, h.cfg.JWTExpiry)
	if err != nil {
		return nil, err
	}
	refresh, err := utils.GenerateRefreshToken(user, h.cfg.RefreshSecret, h.cfg.RefreshExpiry)
	if err != nil {
		return nil, err
	}
	return &models.AuthResponse{
		AccessToken:  access,
		RefreshToken: refresh,
		User:         user,
	}, nil
}
