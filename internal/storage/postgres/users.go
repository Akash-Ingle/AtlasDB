package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

type User struct {
	UserID       uuid.UUID `json:"user_id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type UserStore struct {
	pool   *pgxpool.Pool
	logger zerolog.Logger
}

func NewUserStore(pool *pgxpool.Pool, logger zerolog.Logger) *UserStore {
	return &UserStore{pool: pool, logger: logger}
}

func (s *UserStore) Create(ctx context.Context, email, passwordHash, role string) (*User, error) {
	user := &User{
		UserID:       uuid.New(),
		Email:        email,
		PasswordHash: passwordHash,
		Role:         role,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	_, err := s.pool.Exec(ctx, `
		INSERT INTO users (user_id, email, password_hash, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, user.UserID, user.Email, user.PasswordHash, user.Role, user.CreatedAt, user.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	return user, nil
}

func (s *UserStore) GetByEmail(ctx context.Context, email string) (*User, error) {
	return s.scanUser(s.pool.QueryRow(ctx, `
		SELECT user_id, email, password_hash, role, created_at, updated_at
		FROM users WHERE email = $1
	`, email))
}

func (s *UserStore) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
	return s.scanUser(s.pool.QueryRow(ctx, `
		SELECT user_id, email, password_hash, role, created_at, updated_at
		FROM users WHERE user_id = $1
	`, id))
}

func (s *UserStore) scanUser(row pgx.Row) (*User, error) {
	var u User
	err := row.Scan(&u.UserID, &u.Email, &u.PasswordHash, &u.Role, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}
