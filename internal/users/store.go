package users

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type User struct {
	ID              uuid.UUID
	Email           string
	PasswordHash    string
	EmailVerifiedAt *time.Time
	CreatedAt       time.Time
}

func (u *User) isVerified() bool {
	return u.EmailVerifiedAt != nil
}

var (
	ErrNotFound   = errors.New("user not found")
	ErrEmailTaken = errors.New("email already taken")
)

type Store struct {
	db *pgxpool.Pool
}

func NewStore(db *pgxpool.Pool) *Store {
	return &Store{
		db: db,
	}
}

func (s *Store) Create(ctx context.Context, email, passwordHash string) (*User, error) {
	const q = `
		INSERT INTO users (email, password_hash)
		VALUES ($1, $2)
		RETURNING id, email_verified_at, created_at
	`

	var u User
	u.Email = email
	u.PasswordHash = passwordHash
	err := s.db.QueryRow(ctx, q, email, passwordHash).Scan(
		&u.ID, &u.EmailVerifiedAt, &u.CreatedAt,
	)

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return nil, ErrEmailTaken
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (s *Store) GetByEmail(ctx context.Context, email string) (*User,
	error) {
	const q = `
		SELECT id, email, password_hash, email_verified_at, created_at
		FROM users WHERE email = $1
	`
	var u User
	err := s.db.QueryRow(ctx, q, email).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.EmailVerifiedAt, &u.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return &u, err
}

func (s *Store) GetByID(ctx context.Context, id uuid.UUID) (*User,
	error) {
	const q = `
		SELECT id, email, password_hash, email_verified_at, created_at
		FROM users WHERE id = $1
	`
	var u User
	err := s.db.QueryRow(ctx, q, id).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.EmailVerifiedAt, &u.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return &u, err
}
