package users

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (s *Store) CreateVerificationToken(ctx context.Context, userID uuid.UUID, hash string, ttl time.Duration) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx,
		`DELETE FROM email_verification_tokens WHERE user_id = $1 AND used_at IS NULL`,
		userID,
	); err != nil {
		return err
	}

	if _, err := tx.Exec(
		ctx,
		`INSERT INTO email_verification_tokens (user_id, token_hash, expires_at)
		 VALUES ($1, $2, $3)`,
		userID, hash, time.Now().Add(ttl),
	); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s *Store) ConsumeVerificationToken(ctx context.Context, rawHash string) (uuid.UUID, error) {
	const q = `
		WITH consumed AS (
			UPDATE email_verification_tokens
				SET used_at = now()
			WHERE token_hash = $1
				AND used_at IS NULL
				AND expires_at > now()
			RETURNING user_id
		)
		UPDATE users
			SET email_verified_at = COALESCE(email_verified_at, now())
		WHERE id = (SELECT user_id FROM consumed)
		RETURNING id
	`
	var id uuid.UUID

	err := s.db.QueryRow(ctx, q, rawHash).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, ErrVerificationTokenInvalid
	}

	return id, err
}
