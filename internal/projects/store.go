package projects

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Project struct {
	ID                 uuid.UUID `json:"id"`
	UserID             uuid.UUID `json:"user_id"`
	Name               string    `json:"name"`
	Subdomain          string    `json:"subdomain"`
	ActiveDeploymentID uuid.UUID `json:"active_deployment_id"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

var (
	ErrSubdomainTaken  = errors.New("this subdomain is already taken")
	ErrNameTaken       = errors.New("already have an project with this name")
	ErrProjectNotFound = errors.New("project not found")
)

type Store struct {
	db *pgxpool.Pool
}

func NewStore(db *pgxpool.Pool) *Store {
	return &Store{
		db: db,
	}
}

func (s *Store) Create(ctx context.Context, userID uuid.UUID, name, subdomain string) (*Project, error) {
	const q = `
		INSERT INTO projects (user_id, name, subdomain)
		VALUES ($1, $2, $3)
		RETURNING id, active_deployment_id, created_at
	`
	var p Project
	p.UserID = userID
	p.Name = name
	p.Subdomain = subdomain
	err := s.db.QueryRow(ctx, q, userID, name, subdomain).Scan(
		&p.ID, &p.ActiveDeploymentID, &p.CreatedAt,
	)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		switch pgErr.ConstraintName {
		case "projects_subdomain_key":
			return nil, ErrSubdomainTaken
		case "projects_user_id_name_key":
			return nil, ErrNameTaken
		}
		return nil, ErrSubdomainTaken
	}
	return &p, err
}

func (s *Store) GetOwnedByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*Project, error) {
	const q = `
		select id, user_id, name, subdomain, active_deployment_id, created_at
		FROM projects
		WHERE id = $1 AND user_id = $2;
	`

	var p Project
	err := s.db.QueryRow(ctx, q, id, userID).Scan(
		&p.ID, &p.UserID, &p.Name, &p.Subdomain, &p.ActiveDeploymentID, &p.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrProjectNotFound
		}
		return nil, err
	}
	return &p, nil
}
