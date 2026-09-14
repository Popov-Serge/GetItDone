package repository

import (
	"context"
	"errors"
	"getitdone/internal/apperror"
	"getitdone/internal/domain"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

func (ur *UserRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
        INSERT INTO users (
             name,
             phone,
             email,
             password_hash,
             timezone              
        )
        VALUES ($1, $2, $3, $4, $5)
        RETURNING id, timezone, created_at, updated_at
        `

	err := ur.db.QueryRow(
		ctx,
		query,
		user.Name,
		user.Phone,
		user.Email,
		user.PasswordHash,
		user.Timezone,
	).Scan(
		&user.ID,
		&user.Timezone,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			switch pgErr.ConstraintName {
			case "users_phone_key":
				return apperror.ErrPhoneAlreadyExists
			case "users_email_key":
				return apperror.ErrEmailAlreadyExists
			}
		}

		return err
	}
	return nil
}

func (ur *UserRepository) GetByPhone(ctx context.Context, phone string) (*domain.User, error) {
	query := `
        SELECT
            id,
            name,
            phone,
            email,
            password_hash,
            timezone,
            created_at,
            updated_at
        FROM users
        WHERE phone = $1
    `

	user := domain.User{}

	err := ur.db.QueryRow(
		ctx,
		query,
		phone,
	).Scan(
		&user.ID,
		&user.Name,
		&user.Phone,
		&user.Email,
		&user.PasswordHash,
		&user.Timezone,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (ur *UserRepository) ExistsByPhoneOrEmail(
	ctx context.Context,
	phone string,
	email string,
) (bool, bool, error) {
	query := `
        SELECT
             EXISTS (
                  SELECT 1
                  FROM users
                  WHERE phone = $1
         ),
             EXISTS (
                  SELECT 1
                  FROM users
                  WHERE email = $2
                  )

`
	var phoneExists bool
	var emailExists bool

	err := ur.db.QueryRow(
		ctx,
		query,
		phone,
		email,
	).Scan(&phoneExists, &emailExists)

	if err != nil {
		return false, false, err
	}

	return phoneExists, emailExists, nil
}
