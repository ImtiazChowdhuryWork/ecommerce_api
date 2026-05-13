package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"ecommerce-api/internal/models"
)

type ProductRepository interface {
	Create(ctx context.Context, p *models.Product) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Product, error)
	List(ctx context.Context, params models.ProductListParams) ([]*models.Product, int, error)
	Update(ctx context.Context, p *models.Product) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type productRepo struct{ db *pgxpool.Pool }

func NewProductRepository(db *pgxpool.Pool) ProductRepository { return &productRepo{db: db} }

func (r *productRepo) Create(ctx context.Context, p *models.Product) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO products (id, name, description, price, stock, category_id, image_url, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		p.ID, p.Name, p.Description, p.Price, p.Stock,
		nullUUID(p.CategoryID), p.ImageURL, p.CreatedAt, p.UpdatedAt)
	return err
}

func (r *productRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.Product, error) {
	p := &models.Product{}
	var catID pgtype.UUID
	var catName, catDesc *string
	var catCreatedAt, catUpdatedAt *time.Time

	err := r.db.QueryRow(ctx, `
		SELECT p.id, p.name, p.description, p.price, p.stock, p.category_id, p.image_url, p.created_at, p.updated_at,
		       c.id, c.name, c.description, c.created_at, c.updated_at
		FROM products p
		LEFT JOIN categories c ON p.category_id = c.id
		WHERE p.id = $1`, id).
		Scan(
			&p.ID, &p.Name, &p.Description, &p.Price, &p.Stock,
			&catID, &p.ImageURL, &p.CreatedAt, &p.UpdatedAt,
			new(pgtype.UUID), &catName, &catDesc, &catCreatedAt, &catUpdatedAt,
		)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if catID.Valid {
		uid := uuid.UUID(catID.Bytes)
		p.CategoryID = &uid
		p.Category = &models.Category{
			ID:          uid,
			Name:        *catName,
			Description: *catDesc,
			CreatedAt:   *catCreatedAt,
			UpdatedAt:   *catUpdatedAt,
		}
	}
	return p, nil
}

func (r *productRepo) List(ctx context.Context, params models.ProductListParams) ([]*models.Product, int, error) {
	conditions := []string{"1=1"}
	args := []interface{}{}
	i := 1

	if params.Search != "" {
		conditions = append(conditions, fmt.Sprintf(
			"(p.name ILIKE $%d OR p.description ILIKE $%d)", i, i))
		args = append(args, "%"+params.Search+"%")
		i++
	}
	if params.CategoryID != nil {
		conditions = append(conditions, fmt.Sprintf("p.category_id = $%d", i))
		args = append(args, *params.CategoryID)
		i++
	}

	where := "WHERE " + strings.Join(conditions, " AND ")

	var total int
	if err := r.db.QueryRow(ctx,
		"SELECT COUNT(*) FROM products p "+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	sortCol := map[string]string{
		"price": "p.price", "name": "p.name",
	}[params.SortBy]
	if sortCol == "" {
		sortCol = "p.created_at"
	}
	sortDir := "DESC"
	if strings.ToLower(params.SortOrder) == "asc" {
		sortDir = "ASC"
	}

	query := fmt.Sprintf(`
		SELECT p.id, p.name, p.description, p.price, p.stock, p.category_id, p.image_url, p.created_at, p.updated_at,
		       c.name
		FROM products p
		LEFT JOIN categories c ON p.category_id = c.id
		%s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d`,
		where, sortCol, sortDir, i, i+1)

	offset := (params.Page - 1) * params.Limit
	args = append(args, params.Limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var products []*models.Product
	for rows.Next() {
		p := &models.Product{}
		var catID pgtype.UUID
		var catName *string

		if err := rows.Scan(
			&p.ID, &p.Name, &p.Description, &p.Price, &p.Stock,
			&catID, &p.ImageURL, &p.CreatedAt, &p.UpdatedAt,
			&catName,
		); err != nil {
			return nil, 0, err
		}
		if catID.Valid {
			uid := uuid.UUID(catID.Bytes)
			p.CategoryID = &uid
			if catName != nil {
				p.Category = &models.Category{ID: uid, Name: *catName}
			}
		}
		products = append(products, p)
	}
	if products == nil {
		products = []*models.Product{}
	}
	return products, total, rows.Err()
}

func (r *productRepo) Update(ctx context.Context, p *models.Product) error {
	p.UpdatedAt = time.Now()
	_, err := r.db.Exec(ctx, `
		UPDATE products
		SET name = $1, description = $2, price = $3, stock = $4,
		    category_id = $5, image_url = $6, updated_at = $7
		WHERE id = $8`,
		p.Name, p.Description, p.Price, p.Stock,
		nullUUID(p.CategoryID), p.ImageURL, p.UpdatedAt, p.ID)
	return err
}

func (r *productRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM products WHERE id = $1`, id)
	return err
}

// nullUUID converts *uuid.UUID to pgtype.UUID for nullable columns.
func nullUUID(u *uuid.UUID) pgtype.UUID {
	if u == nil {
		return pgtype.UUID{}
	}
	return pgtype.UUID{Bytes: *u, Valid: true}
}
