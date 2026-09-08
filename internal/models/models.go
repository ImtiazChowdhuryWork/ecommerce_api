package models

import (
	"time"

	"github.com/google/uuid"
)

// ------- Enums -------

type Role string

const (
	RoleCustomer Role = "customer"
	RoleAdmin    Role = "admin"
)

type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusConfirmed OrderStatus = "confirmed"
	OrderStatusShipped   OrderStatus = "shipped"
	OrderStatusDelivered OrderStatus = "delivered"
	OrderStatusCancelled OrderStatus = "cancelled"
)

// ------- Domain models -------

type User struct {
	ID           uuid.UUID `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Name         string    `json:"name"`
	Role         Role      `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Category struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Product struct {
	ID          uuid.UUID  `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Price       float64    `json:"price"`
	Stock       int        `json:"stock"`
	CategoryID  *uuid.UUID `json:"category_id"`
	ImageURL    string     `json:"image_url"`
	Category    *Category  `json:"category,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type Cart struct {
	ID        uuid.UUID  `json:"id"`
	UserID    uuid.UUID  `json:"user_id"`
	Items     []CartItem `json:"items"`
	Total     float64    `json:"total"`
	ItemCount int        `json:"item_count"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type CartItem struct {
	ID        uuid.UUID `json:"id"`
	CartID    uuid.UUID `json:"cart_id"`
	ProductID uuid.UUID `json:"product_id"`
	Product   *Product  `json:"product,omitempty"`
	Quantity  int       `json:"quantity"`
	Subtotal  float64   `json:"subtotal"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Order struct {
	ID              uuid.UUID   `json:"id"`
	UserID          uuid.UUID   `json:"user_id"`
	Status          OrderStatus `json:"status"`
	TotalAmount     float64     `json:"total_amount"`
	ShippingAddress string      `json:"shipping_address"`
	Notes           string      `json:"notes,omitempty"`
	Items           []OrderItem `json:"items,omitempty"`
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`
}

type OrderItem struct {
	ID        uuid.UUID `json:"id"`
	OrderID   uuid.UUID `json:"order_id"`
	ProductID uuid.UUID `json:"product_id"`
	Product   *Product  `json:"product,omitempty"`
	Quantity  int       `json:"quantity"`
	Price     float64   `json:"price"`
	Subtotal  float64   `json:"subtotal"`
	CreatedAt time.Time `json:"created_at"`
}

type Review struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	ProductID uuid.UUID `json:"product_id"`
	User      *UserPublic `json:"user,omitempty"`
	Rating    int       `json:"rating"`
	Comment   string    `json:"comment"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UserPublic is a safe subset of User for embedding in other responses
type UserPublic struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

// ------- Request types -------

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type UpdateProfileRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

type CreateCategoryRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type UpdateCategoryRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

type CreateProductRequest struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Price       float64    `json:"price"`
	Stock       int        `json:"stock"`
	CategoryID  *uuid.UUID `json:"category_id"`
	ImageURL    string     `json:"image_url"`
}

type UpdateProductRequest struct {
	Name        *string    `json:"name"`
	Description *string    `json:"description"`
	Price       *float64   `json:"price"`
	Stock       *int       `json:"stock"`
	CategoryID  *uuid.UUID `json:"category_id"`
	ImageURL    *string    `json:"image_url"`
}

type AddCartItemRequest struct {
	ProductID uuid.UUID `json:"product_id"`
	Quantity  int       `json:"quantity"`
}

type UpdateCartItemRequest struct {
	Quantity int `json:"quantity"`
}

type CreateOrderRequest struct {
	ShippingAddress string `json:"shipping_address"`
	Notes           string `json:"notes"`
}

type UpdateOrderStatusRequest struct {
	Status OrderStatus `json:"status"`
}

type CreateReviewRequest struct {
	Rating  int    `json:"rating"`
	Comment string `json:"comment"`
}

type UpdateReviewRequest struct {
	Rating  *int    `json:"rating"`
	Comment *string `json:"comment"`
}

// ------- Shared response types -------

type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	User         *User  `json:"user"`
}

type ProductListParams struct {
	Page       int
	Limit      int
	Search     string
	CategoryID *uuid.UUID
	SortBy     string
	SortOrder  string
}

type PaginatedResponse struct {
	Items      interface{} `json:"items"`
	Total      int         `json:"total"`
	Page       int         `json:"page"`
	Limit      int         `json:"limit"`
	TotalPages int         `json:"total_pages"`
}
