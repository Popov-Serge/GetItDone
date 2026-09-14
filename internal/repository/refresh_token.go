package repository

import (
	"context"
	"getitdone/internal/apperror"
	"getitdone/internal/domain"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RefreshTokenRepository struct {
	db *pgxpool.Pool
}

func NewRefreshTokenRepository(db *pgxpool.Pool) *RefreshTokenRepository {
	return &RefreshTokenRepository{
		db: db,
	}
}

func (r *RefreshTokenRepository) Create(
	ctx context.Context,
	userID uuid.UUID,
	tokerHash string,
	expiresAt time.Time,
) error {

	const query = `
INSERT INTO refresh_tokens (
                            user_id,
                            token_hash,
                            expires_at
)
VALUES ($1, $2, $3)
`
	_, err := r.db.Exec(ctx, query, userID, tokerHash, expiresAt)

	return err
}

func (r *RefreshTokenRepository) FindByTokenHash(ctx context.Context, tokenHash string) (*domain.RefreshToken, error) {
	const query = `
SELECT
 id,
user_id,
token_hash,
expires_at,
created_at,
revoked_at
FROM refresh_tokens
WHERE token_hash = $1
`

	token := &domain.RefreshToken{}

	err := r.db.QueryRow(ctx, query, tokenHash).Scan(
		&token.ID,
		&token.UserID,
		&token.TokenHash,
		&token.ExpiresAt,
		&token.CreatedAt,
		&token.RevokedAt,
	)

	if err != nil {
		return nil, err
	}

	return token, nil
}

func (r *RefreshTokenRepository) Rotate(
	ctx context.Context,
	oldTokenHash string,
	newTokenHash string,
	userID uuid.UUID,
	expiresAt time.Time,
) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}

	defer func() {
		_ = tx.Rollback(ctx)
	}()

	const revokeQuery = `
UPDATE refresh_tokens
SET revoked_at = NOW()
WHERE token_hash = $1
  AND user_id = $2
  AND revoked_at IS NULL
  AND expires_at > NOW()
`

	result, err := tx.Exec(
		ctx,
		revokeQuery,
		oldTokenHash,
		userID,
	)

	if err != nil {
		return err
	}

	if result.RowsAffected() != 1 {
		return apperror.ErrInvalidRefreshToken
	}

	const insertQuery = `
INSERT INTO refresh_tokens (
    user_id,
    token_hash,
    expires_at
)
VALUES ($1, $2, $3)
`

	_, err = tx.Exec(
		ctx,
		insertQuery,
		userID,
		newTokenHash,
		expiresAt,
	)

	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *RefreshTokenRepository) Revoke(
	ctx context.Context,
	tokenHash string,
	userID uuid.UUID,
) error {
	const query = `

UPDATE refresh_tokens
SET revoked_at = NOW()
WHERE token_hash = $1
AND user_id = $2
AND revoked_at IS NULL`

	_, err := r.db.Exec(
		ctx,
		query,
		tokenHash,
		userID,
	)

	return err
}
