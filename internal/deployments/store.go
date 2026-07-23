package deployments

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Deployment struct {
	ID                  uuid.UUID
	ProjectID           uuid.UUID
	Version             int
	Status              string
	UploadKey           string
	StoragePrefix       string
	FileCount           int
	TotalBytes          int64
	ErrorMessage        string
	ProcessingStartedAt time.Time
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

type Store struct {
	db *pgxpool.Pool
}

func NewStore(db *pgxpool.Pool) *Store {
	return &Store{
		db: db,
	}
}

func (s *Store) CreateWithID(ctx context.Context, id, projectID uuid.UUID, uploadKey string) (*Deployment, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var dummy uuid.UUID
	if err := tx.QueryRow(ctx,
		`SELECT id FROM projects WHERE id = $1 FOR UPDATE`,
		projectID,
	).Scan(&dummy); err != nil {
		return nil, err
	}

	const q = `
	INSERT INTO deployments (id, project_id, version, status, upload_key)
	VALUES (
		$1, $2,
		(SELECT COALESCE(MAX(version), 0) + 1 FROM deployments WHERE project_id = $2),
		'queued', $3
	)
	RETURNING id, project_id, version, status, upload_key, created_at`

	var d Deployment
	if err := tx.QueryRow(ctx, q, id, projectID, uploadKey).Scan(
		&d.ID, &d.ProjectID, &d.Version, &d.Status, &d.UploadKey, &d.CreatedAt,
	); err != nil {
		return nil, err
	}
	return &d, tx.Commit(ctx)
}

func (s *Store) MarkFailed(ctx context.Context, id uuid.UUID, error string) error {
	const q = `
		UPDATE deployments
		SET status = 'failed', error_message = $2
		WHERE id = $1
	`
	_, err := s.db.Exec(ctx, q, id, error)
	return err
}
