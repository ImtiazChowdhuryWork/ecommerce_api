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

type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	UpdatePassword(ctx context.Context, id uuid.UUID, hash string) error
	EmailExists(ctx context.Context, email string) (bool, error)
}

type userRepo struct{ db *pgxpool.Pool }

func NewUserRepository(db *pgxpool.Pool) UserRepository { return &userRepo{db: db} }

func (r *userRepo) Create(ctx context.Context, u *models.User) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO users (id, email, password_hash, name, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		u.ID, u.Email, u.PasswordHash, u.Name, u.Role, u.CreatedAt, u.UpdatedAt)
	return err
}

func (r *userRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	u := &models.User{}
	err := r.db.QueryRow(ctx, `
		SELECT id, email, password_hash, name, role, created_at, updated_at
		FROM users WHERE id = $1`, id).
		Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.Role, &u.CreatedAt, &u.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return u, err
}

func (r *userRepo) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	u := &models.User{}
	err := r.db.QueryRow(ctx, `
		SELECT id, email, password_hash, name, role, created_at, updated_at
		FROM users WHERE email = $1`, email).
		Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.Role, &u.CreatedAt, &u.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return u, err
}

func (r *userRepo) Update(ctx context.Context, u *models.User) error {
	u.UpdatedAt = time.Now()
	_, err := r.db.Exec(ctx, `
		UPDATE users SET email = $1, name = $2, updated_at = $3 WHERE id = $4`,
		u.Email, u.Name, u.UpdatedAt, u.ID)
	return err
}

func (r *userRepo) UpdatePassword(ctx context.Context, id uuid.UUID, hash string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE users SET password_hash = $1, updated_at = NOW() WHERE id = $2`, hash, id)
	return err
}

func (r *userRepo) EmailExists(ctx context.Context, email string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`, email).Scan(&exists)
	return exists, err
}
