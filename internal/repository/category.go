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

type CategoryRepository interface {
	Create(ctx context.Context, c *models.Category) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Category, error)
	List(ctx context.Context) ([]*models.Category, error)
	Update(ctx context.Context, c *models.Category) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type categoryRepo struct{ db *pgxpool.Pool }

func NewCategoryRepository(db *pgxpool.Pool) CategoryRepository { return &categoryRepo{db: db} }

func (r *categoryRepo) Create(ctx context.Context, c *models.Category) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO categories (id, name, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)`,
		c.ID, c.Name, c.Description, c.CreatedAt, c.UpdatedAt)
	return err
}

func (r *categoryRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.Category, error) {
	c := &models.Category{}
	err := r.db.QueryRow(ctx, `
		SELECT id, name, description, created_at, updated_at
		FROM categories WHERE id = $1`, id).
		Scan(&c.ID, &c.Name, &c.Description, &c.CreatedAt, &c.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return c, err
}

func (r *categoryRepo) List(ctx context.Context) ([]*models.Category, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, name, description, created_at, updated_at
		FROM categories ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cats []*models.Category
	for rows.Next() {
		c := &models.Category{}
		if err := rows.Scan(&c.ID, &c.Name, &c.Description, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		cats = append(cats, c)
	}
	if cats == nil {
		cats = []*models.Category{}
	}
	return cats, rows.Err()
}

func (r *categoryRepo) Update(ctx context.Context, c *models.Category) error {
	c.UpdatedAt = time.Now()
	_, err := r.db.Exec(ctx, `
		UPDATE categories SET name = $1, description = $2, updated_at = $3 WHERE id = $4`,
		c.Name, c.Description, c.UpdatedAt, c.ID)
	return err
}

func (r *categoryRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM categories WHERE id = $1`, id)
	return err
}
