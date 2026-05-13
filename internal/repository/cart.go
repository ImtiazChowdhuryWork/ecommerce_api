package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"ecommerce-api/internal/models"
)

type CartRepository interface {
	GetOrCreateByUserID(ctx context.Context, userID uuid.UUID) (*models.Cart, error)
	AddItem(ctx context.Context, cartID, productID uuid.UUID, quantity int) (*models.CartItem, error)
	UpdateItem(ctx context.Context, itemID uuid.UUID, quantity int) (*models.CartItem, error)
	RemoveItem(ctx context.Context, itemID, cartID uuid.UUID) error
	ClearItems(ctx context.Context, cartID uuid.UUID) error
	GetItemByID(ctx context.Context, itemID uuid.UUID) (*models.CartItem, error)
}

type cartRepo struct{ db *pgxpool.Pool }

func NewCartRepository(db *pgxpool.Pool) CartRepository { return &cartRepo{db: db} }

func (r *cartRepo) GetOrCreateByUserID(ctx context.Context, userID uuid.UUID) (*models.Cart, error) {
	cart := &models.Cart{}
	err := r.db.QueryRow(ctx, `
		SELECT id, user_id, created_at, updated_at FROM carts WHERE user_id = $1`, userID).
		Scan(&cart.ID, &cart.UserID, &cart.CreatedAt, &cart.UpdatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		cart.ID = uuid.New()
		cart.UserID = userID
		cart.CreatedAt = time.Now()
		cart.UpdatedAt = time.Now()
		_, err = r.db.Exec(ctx, `
			INSERT INTO carts (id, user_id, created_at, updated_at) VALUES ($1, $2, $3, $4)`,
			cart.ID, cart.UserID, cart.CreatedAt, cart.UpdatedAt)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	// Load items with product details
	rows, err := r.db.Query(ctx, `
		SELECT ci.id, ci.cart_id, ci.product_id, ci.quantity, ci.created_at, ci.updated_at,
		       p.id, p.name, p.description, p.price, p.stock, p.image_url
		FROM cart_items ci
		JOIN products p ON ci.product_id = p.id
		WHERE ci.cart_id = $1
		ORDER BY ci.created_at`, cart.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var total float64
	cart.Items = []models.CartItem{}
	for rows.Next() {
		item := models.CartItem{Product: &models.Product{}}
		err := rows.Scan(
			&item.ID, &item.CartID, &item.ProductID, &item.Quantity,
			&item.CreatedAt, &item.UpdatedAt,
			&item.Product.ID, &item.Product.Name, &item.Product.Description,
			&item.Product.Price, &item.Product.Stock, &item.Product.ImageURL,
		)
		if err != nil {
			return nil, err
		}
		item.Subtotal = item.Product.Price * float64(item.Quantity)
		total += item.Subtotal
		cart.Items = append(cart.Items, item)
	}
	cart.Total = total
	cart.ItemCount = len(cart.Items)
	return cart, rows.Err()
}

func (r *cartRepo) AddItem(ctx context.Context, cartID, productID uuid.UUID, quantity int) (*models.CartItem, error) {
	// Upsert: if the product is already in cart, add to quantity
	item := &models.CartItem{}
	err := r.db.QueryRow(ctx, `
		INSERT INTO cart_items (id, cart_id, product_id, quantity, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		ON CONFLICT (cart_id, product_id)
		DO UPDATE SET quantity = cart_items.quantity + EXCLUDED.quantity, updated_at = NOW()
		RETURNING id, cart_id, product_id, quantity, created_at, updated_at`,
		uuid.New(), cartID, productID, quantity).
		Scan(&item.ID, &item.CartID, &item.ProductID, &item.Quantity, &item.CreatedAt, &item.UpdatedAt)
	return item, err
}

func (r *cartRepo) UpdateItem(ctx context.Context, itemID uuid.UUID, quantity int) (*models.CartItem, error) {
	item := &models.CartItem{}
	err := r.db.QueryRow(ctx, `
		UPDATE cart_items SET quantity = $1, updated_at = NOW()
		WHERE id = $2
		RETURNING id, cart_id, product_id, quantity, created_at, updated_at`,
		quantity, itemID).
		Scan(&item.ID, &item.CartID, &item.ProductID, &item.Quantity, &item.CreatedAt, &item.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return item, err
}

func (r *cartRepo) RemoveItem(ctx context.Context, itemID, cartID uuid.UUID) error {
	_, err := r.db.Exec(ctx,
		`DELETE FROM cart_items WHERE id = $1 AND cart_id = $2`, itemID, cartID)
	return err
}

func (r *cartRepo) ClearItems(ctx context.Context, cartID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM cart_items WHERE cart_id = $1`, cartID)
	return err
}

func (r *cartRepo) GetItemByID(ctx context.Context, itemID uuid.UUID) (*models.CartItem, error) {
	item := &models.CartItem{}
	err := r.db.QueryRow(ctx, `
		SELECT id, cart_id, product_id, quantity, created_at, updated_at
		FROM cart_items WHERE id = $1`, itemID).
		Scan(&item.ID, &item.CartID, &item.ProductID, &item.Quantity, &item.CreatedAt, &item.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return item, err
}
