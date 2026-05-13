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

type ReviewRepository interface {
	Create(ctx context.Context, r *models.Review) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Review, error)
	ListByProduct(ctx context.Context, productID uuid.UUID, page, limit int) ([]*models.Review, int, error)
	Update(ctx context.Context, r *models.Review) error
	Delete(ctx context.Context, id, userID uuid.UUID, isAdmin bool) error
	GetByUserAndProduct(ctx context.Context, userID, productID uuid.UUID) (*models.Review, error)
}

type reviewRepo struct{ db *pgxpool.Pool }

func NewReviewRepository(db *pgxpool.Pool) ReviewRepository { return &reviewRepo{db: db} }

func (r *reviewRepo) Create(ctx context.Context, rev *models.Review) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO reviews (id, user_id, product_id, rating, comment, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		rev.ID, rev.UserID, rev.ProductID, rev.Rating, rev.Comment, rev.CreatedAt, rev.UpdatedAt)
	return err
}

func (r *reviewRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.Review, error) {
	rev := &models.Review{}
	err := r.db.QueryRow(ctx, `
		SELECT id, user_id, product_id, rating, comment, created_at, updated_at
		FROM reviews WHERE id = $1`, id).
		Scan(&rev.ID, &rev.UserID, &rev.ProductID, &rev.Rating, &rev.Comment, &rev.CreatedAt, &rev.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return rev, err
}

func (r *reviewRepo) ListByProduct(ctx context.Context, productID uuid.UUID, page, limit int) ([]*models.Review, int, error) {
	var total int
	r.db.QueryRow(ctx, `SELECT COUNT(*) FROM reviews WHERE product_id = $1`, productID).Scan(&total)

	rows, err := r.db.Query(ctx, `
		SELECT r.id, r.user_id, r.product_id, r.rating, r.comment, r.created_at, r.updated_at,
		       u.name
		FROM reviews r
		JOIN users u ON r.user_id = u.id
		WHERE r.product_id = $1
		ORDER BY r.created_at DESC
		LIMIT $2 OFFSET $3`,
		productID, limit, (page-1)*limit)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var reviews []*models.Review
	for rows.Next() {
		rev := &models.Review{User: &models.UserPublic{}}
		err := rows.Scan(
			&rev.ID, &rev.UserID, &rev.ProductID, &rev.Rating, &rev.Comment,
			&rev.CreatedAt, &rev.UpdatedAt,
			&rev.User.Name,
		)
		if err != nil {
			return nil, 0, err
		}
		rev.User.ID = rev.UserID
		reviews = append(reviews, rev)
	}
	if reviews == nil {
		reviews = []*models.Review{}
	}
	return reviews, total, rows.Err()
}

func (r *reviewRepo) Update(ctx context.Context, rev *models.Review) error {
	rev.UpdatedAt = time.Now()
	_, err := r.db.Exec(ctx, `
		UPDATE reviews SET rating = $1, comment = $2, updated_at = $3 WHERE id = $4`,
		rev.Rating, rev.Comment, rev.UpdatedAt, rev.ID)
	return err
}

func (r *reviewRepo) Delete(ctx context.Context, id, userID uuid.UUID, isAdmin bool) error {
	var query string
	var args []interface{}
	if isAdmin {
		query = `DELETE FROM reviews WHERE id = $1`
		args = []interface{}{id}
	} else {
		query = `DELETE FROM reviews WHERE id = $1 AND user_id = $2`
		args = []interface{}{id, userID}
	}
	res, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return errors.New("review not found or not owned by user")
	}
	return nil
}

func (r *reviewRepo) GetByUserAndProduct(ctx context.Context, userID, productID uuid.UUID) (*models.Review, error) {
	rev := &models.Review{}
	err := r.db.QueryRow(ctx, `
		SELECT id, user_id, product_id, rating, comment, created_at, updated_at
		FROM reviews WHERE user_id = $1 AND product_id = $2`, userID, productID).
		Scan(&rev.ID, &rev.UserID, &rev.ProductID, &rev.Rating, &rev.Comment, &rev.CreatedAt, &rev.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return rev, err
}
