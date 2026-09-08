package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"ecommerce-api/internal/models"
)

type OrderRepository interface {
	Create(ctx context.Context, userID uuid.UUID, req *models.CreateOrderRequest, cart *models.Cart) (*models.Order, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.Order, error)
	ListByUser(ctx context.Context, userID uuid.UUID, page, limit int) ([]*models.Order, int, error)
	ListAll(ctx context.Context, page, limit int) ([]*models.Order, int, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status models.OrderStatus) error
	Cancel(ctx context.Context, id, userID uuid.UUID, isAdmin bool) error
}

type orderRepo struct{ db *pgxpool.Pool }

func NewOrderRepository(db *pgxpool.Pool) OrderRepository { return &orderRepo{db: db} }

func (r *orderRepo) Create(ctx context.Context, userID uuid.UUID, req *models.CreateOrderRequest, cart *models.Cart) (*models.Order, error) {
	if len(cart.Items) == 0 {
		return nil, errors.New("cart is empty")
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	order := &models.Order{
		ID:              uuid.New(),
		UserID:          userID,
		Status:          models.OrderStatusPending,
		TotalAmount:     cart.Total,
		ShippingAddress: req.ShippingAddress,
		Notes:           req.Notes,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO orders (id, user_id, status, total_amount, shipping_address, notes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		order.ID, order.UserID, order.Status, order.TotalAmount,
		order.ShippingAddress, order.Notes, order.CreatedAt, order.UpdatedAt)
	if err != nil {
		return nil, err
	}

	for _, item := range cart.Items {
		_, err = tx.Exec(ctx, `
			INSERT INTO order_items (id, order_id, product_id, quantity, price, created_at)
			VALUES ($1, $2, $3, $4, $5, NOW())`,
			uuid.New(), order.ID, item.ProductID, item.Quantity, item.Product.Price)
		if err != nil {
			return nil, err
		}

		res, err := tx.Exec(ctx, `
			UPDATE products SET stock = stock - $1, updated_at = NOW()
			WHERE id = $2 AND stock >= $1`,
			item.Quantity, item.ProductID)
		if err != nil {
			return nil, err
		}
		if res.RowsAffected() == 0 {
			return nil, fmt.Errorf("insufficient stock for product %s", item.Product.Name)
		}
	}

	// Clear the cart
	_, err = tx.Exec(ctx, `DELETE FROM cart_items WHERE cart_id = $1`, cart.ID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return order, nil
}

func (r *orderRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.Order, error) {
	order := &models.Order{}
	err := r.db.QueryRow(ctx, `
		SELECT id, user_id, status, total_amount, shipping_address, COALESCE(notes, ''), created_at, updated_at
		FROM orders WHERE id = $1`, id).
		Scan(&order.ID, &order.UserID, &order.Status, &order.TotalAmount,
			&order.ShippingAddress, &order.Notes, &order.CreatedAt, &order.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	// Load items
	rows, err := r.db.Query(ctx, `
		SELECT oi.id, oi.order_id, oi.product_id, oi.quantity, oi.price, oi.created_at,
		       p.name, p.image_url
		FROM order_items oi
		JOIN products p ON oi.product_id = p.id
		WHERE oi.order_id = $1`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	order.Items = []models.OrderItem{}
	for rows.Next() {
		item := models.OrderItem{Product: &models.Product{}}
		err := rows.Scan(
			&item.ID, &item.OrderID, &item.ProductID, &item.Quantity, &item.Price, &item.CreatedAt,
			&item.Product.Name, &item.Product.ImageURL,
		)
		if err != nil {
			return nil, err
		}
		item.Product.ID = item.ProductID
		item.Subtotal = item.Price * float64(item.Quantity)
		order.Items = append(order.Items, item)
	}
	return order, rows.Err()
}

func (r *orderRepo) ListByUser(ctx context.Context, userID uuid.UUID, page, limit int) ([]*models.Order, int, error) {
	var total int
	r.db.QueryRow(ctx, `SELECT COUNT(*) FROM orders WHERE user_id = $1`, userID).Scan(&total)

	rows, err := r.db.Query(ctx, `
		SELECT id, user_id, status, total_amount, shipping_address, COALESCE(notes, ''), created_at, updated_at
		FROM orders WHERE user_id = $1
		ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		userID, limit, (page-1)*limit)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	return scanOrders(rows, total)
}

func (r *orderRepo) ListAll(ctx context.Context, page, limit int) ([]*models.Order, int, error) {
	var total int
	r.db.QueryRow(ctx, `SELECT COUNT(*) FROM orders`).Scan(&total)

	rows, err := r.db.Query(ctx, `
		SELECT id, user_id, status, total_amount, shipping_address, COALESCE(notes, ''), created_at, updated_at
		FROM orders
		ORDER BY created_at DESC LIMIT $1 OFFSET $2`,
		limit, (page-1)*limit)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	return scanOrders(rows, total)
}

func (r *orderRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status models.OrderStatus) error {
	res, err := r.db.Exec(ctx,
		`UPDATE orders SET status = $1, updated_at = NOW() WHERE id = $2`, status, id)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return errors.New("order not found")
	}
	return nil
}

func (r *orderRepo) Cancel(ctx context.Context, id, userID uuid.UUID, isAdmin bool) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var status models.OrderStatus
	var ownerID uuid.UUID
	err = tx.QueryRow(ctx,
		`SELECT status, user_id FROM orders WHERE id = $1`, id).
		Scan(&status, &ownerID)
	if errors.Is(err, pgx.ErrNoRows) {
		return errors.New("order not found")
	}
	if err != nil {
		return err
	}
	if !isAdmin && ownerID != userID {
		return errors.New("order not found")
	}
	if status != models.OrderStatusPending {
		return fmt.Errorf("cannot cancel an order with status '%s'", status)
	}

	// Restore stock
	_, err = tx.Exec(ctx, `
		UPDATE products p
		SET stock = p.stock + oi.quantity, updated_at = NOW()
		FROM order_items oi
		WHERE oi.order_id = $1 AND p.id = oi.product_id`, id)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx,
		`UPDATE orders SET status = 'cancelled', updated_at = NOW() WHERE id = $1`, id)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func scanOrders(rows pgx.Rows, total int) ([]*models.Order, int, error) {
	var orders []*models.Order
	for rows.Next() {
		o := &models.Order{}
		if err := rows.Scan(&o.ID, &o.UserID, &o.Status, &o.TotalAmount,
			&o.ShippingAddress, &o.Notes, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, 0, err
		}
		orders = append(orders, o)
	}
	if orders == nil {
		orders = []*models.Order{}
	}
	return orders, total, rows.Err()
}
