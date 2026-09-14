package repository

import (
	"context"
	"getitdone/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type FamilyRepository struct {
	db *pgxpool.Pool
}

func NewFamilyRepository(db *pgxpool.Pool) *FamilyRepository {
	return &FamilyRepository{db: db}
}

func (repo *FamilyRepository) Create(
	ctx context.Context,
	family *domain.Family,
) error {
	tx, err := repo.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	familyQuery := `
INSERT INTO families (name, created_by)
VALUES ($1, $2)
RETURNING id, created_at, updated_at
`

	err = tx.QueryRow(
		ctx,
		familyQuery,
		family.Name,
		family.CreatedBy,
	).Scan(
		&family.ID,
		&family.CreatedAt,
		&family.UpdatedAt,
	)
	if err != nil {
		return err
	}

	memberQuery := `
INSERT INTO family_members (
	family_id,
	user_id,
	name,
	relation_type_id,
	role
)
SELECT
	$1,
	u.id,
	u.name,
	fr.id,
	'owner'
FROM users u
CROSS JOIN family_relation_types fr
WHERE u.id = $2
  AND fr.code = 'other'
  AND fr.is_active = TRUE
`

	_, err = tx.Exec(
		ctx,
		memberQuery,
		family.ID,
		family.CreatedBy,
	)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (repo *FamilyRepository) GetByID(
	ctx context.Context,
	familyID uuid.UUID,
	userID uuid.UUID,
) (*domain.Family, error) {
	query := `
SELECT
	f.id,
	f.name,
	f.created_by,
	f.created_at,
	f.updated_at
FROM families f
INNER JOIN family_members fm
	ON fm.family_id = f.id
WHERE f.id = $1
  AND fm.user_id = $2
`

	family := &domain.Family{}

	err := repo.db.QueryRow(
		ctx,
		query,
		familyID,
		userID,
	).Scan(
		&family.ID,
		&family.Name,
		&family.CreatedBy,
		&family.CreatedAt,
		&family.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return family, nil
}

func (repo *FamilyRepository) Update(
	ctx context.Context,
	family *domain.Family,
) error {
	query := `
UPDATE families f
SET
	name = $1,
	updated_at = NOW()
FROM family_members fm
WHERE f.id = $2
  AND fm.family_id = f.id
  AND fm.user_id = $3
  AND fm.role = 'owner'
RETURNING
    f.id,
    f.name,
    f.created_by,
    f.created_at,
    f.updated_at
`

	err := repo.db.QueryRow(
		ctx,
		query,
		family.Name,
		family.ID,
		family.CreatedBy,
	).Scan(
		&family.ID,
		&family.Name,
		&family.CreatedBy,
		&family.CreatedAt,
		&family.UpdatedAt,
	)

	return err
}

func (repo *FamilyRepository) GetFamilies(
	ctx context.Context,
	userID uuid.UUID,
) ([]*domain.Family, error) {
	query := `
SELECT
	f.id,
	f.name,
	f.created_by,
	f.created_at,
	f.updated_at
FROM families f
INNER JOIN family_members fm
	ON f.id = fm.family_id
WHERE fm.user_id = $1
ORDER BY f.created_at DESC
`

	rows, err := repo.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	families := make([]*domain.Family, 0)

	for rows.Next() {
		family := &domain.Family{}

		err := rows.Scan(
			&family.ID,
			&family.Name,
			&family.CreatedBy,
			&family.CreatedAt,
			&family.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		families = append(families, family)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return families, nil
}

func (repo *FamilyRepository) Delete(
	ctx context.Context,
	familyID uuid.UUID,
	userID uuid.UUID,
) (bool, error) {
	query := `
DELETE FROM families f
USING family_members fm
WHERE f.id = $1
  AND fm.family_id = f.id
  AND fm.user_id = $2
  AND fm.role = 'owner'
`

	result, err := repo.db.Exec(
		ctx,
		query,
		familyID,
		userID,
	)
	if err != nil {
		return false, err
	}

	return result.RowsAffected() > 0, nil
}
