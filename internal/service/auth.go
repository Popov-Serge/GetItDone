package service

import (
	"context"
	"errors"
	"getitdone/internal/apperror"
	"getitdone/internal/domain"
	"getitdone/internal/validation"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByPhone(ctx context.Context, phone string) (*domain.User, error)
	ExistsByPhoneOrEmail(
		ctx context.Context,
		phone string,
		email string,
	) (bool, bool, error)
}

type TokenService interface {
	Generate(userID uuid.UUID) (string, error)
}

type RefreshRepository interface {
	FindByTokenHash(ctx context.Context, tokenHash string) (*domain.RefreshToken, error)
	Revoke(
		ctx context.Context,
		tokenHash string,
		userID uuid.UUID,
	) error
}

type TokenBlacklistInterface interface {
	Revoke(
		ctx context.Context,
		token string,
		duration time.Duration,
	) error
}

type RefreshTokenServiceInterface interface {
	Generate(
		ctx context.Context,
		userID uuid.UUID,
	) (string, error)

	Rotate(
		ctx context.Context,
		tokenHash string,
		userID uuid.UUID,
	) (string, error)
}

type RegisterInput struct {
	Name     string
	Phone    string
	Email    string
	Password string
	Timezone string
}

type LoginInput struct {
	Phone    string
	Password string
}

type AuthService struct {
	users               UserRepository
	refreshToken        RefreshRepository
	phoneValidator      *validation.PhoneValidator
	tokenService        TokenService
	refreshTokenService RefreshTokenServiceInterface
	tokenBlacklist      TokenBlacklistInterface
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

func NewAuthService(
	users UserRepository,
	refreshToken RefreshRepository,
	phoneValidator *validation.PhoneValidator,
	tokenService TokenService,
	refreshTokenService RefreshTokenServiceInterface,
	tokenBlacklist TokenBlacklistInterface,
) *AuthService {
	return &AuthService{
		users:               users,
		refreshToken:        refreshToken,
		phoneValidator:      phoneValidator,
		tokenService:        tokenService,
		refreshTokenService: refreshTokenService,
		tokenBlacklist:      tokenBlacklist,
	}
}

func (s *AuthService) Register(
	ctx context.Context,
	input RegisterInput,
) (*domain.User, error) {

	normalizedPhone, valid := s.phoneValidator.Validate(input.Phone)

	if !valid {
		return nil, apperror.ErrInvalidPhone
	}

	input.Phone = normalizedPhone

	phoneExists, emailExists, err := s.users.ExistsByPhoneOrEmail(
		ctx,
		input.Phone,
		input.Email,
	)

	if err != nil {
		return nil, err
	}

	if phoneExists || emailExists {
		fields := make(map[string][]string)

		if phoneExists {
			fields["phone"] = []string{"already_exists"}
		}

		if emailExists {
			fields["email"] = []string{"already_exists"}
		}

		return nil, &apperror.RegistrationConflictError{
			Fields: fields,
		}
	}

	passwordHash, err := bcrypt.GenerateFromPassword(
		[]byte(input.Password),
		10,
	)

	if err != nil {
		return nil, err
	}

	user := &domain.User{
		Name:         input.Name,
		Phone:        input.Phone,
		Email:        input.Email,
		PasswordHash: string(passwordHash),
		Timezone:     input.Timezone,
	}

	if err := s.users.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *AuthService) Login(
	ctx context.Context,
	input LoginInput,
) (TokenPair, error) {
	normalizedPhone, valid := s.phoneValidator.Validate(input.Phone)

	if !valid {
		return TokenPair{}, apperror.ErrInvalidCredentials
	}

	user, err := s.users.GetByPhone(ctx, normalizedPhone)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return TokenPair{}, apperror.ErrInvalidCredentials
		}

		return TokenPair{}, err
	}

	err = bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash),
		[]byte(input.Password),
	)

	if err != nil {
		return TokenPair{}, apperror.ErrInvalidCredentials
	}

	accessToken, err := s.tokenService.Generate(user.ID)
	if err != nil {
		return TokenPair{}, err
	}

	refreshToken, err := s.refreshTokenService.Generate(ctx, user.ID)
	if err != nil {
		return TokenPair{}, err
	}

	return TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthService) Refresh(
	ctx context.Context,
	hash string,
) (TokenPair, error) {
	token, err := s.refreshToken.FindByTokenHash(
		ctx,
		hash,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return TokenPair{}, apperror.ErrInvalidRefreshToken
		}

		return TokenPair{}, err
	}

	if token.RevokedAt != nil {
		return TokenPair{}, apperror.ErrInvalidRefreshToken
	}

	if time.Now().After(token.ExpiresAt) {
		return TokenPair{}, apperror.ErrInvalidRefreshToken
	}

	accessToken, err := s.tokenService.Generate(
		token.UserID,
	)
	if err != nil {
		return TokenPair{}, err
	}

	refreshToken, err := s.refreshTokenService.Rotate(
		ctx,
		hash,
		token.UserID,
	)

	if err != nil {
		if errors.Is(err, apperror.ErrInvalidRefreshToken) {
			return TokenPair{}, apperror.ErrInvalidRefreshToken
		}

		return TokenPair{}, err
	}

	return TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthService) Logout(
	ctx context.Context,
	userID uuid.UUID,
	jti string,
	expiresAt time.Time,
	refreshTokenHash string,
) error {
	duration := time.Until(expiresAt)

	if duration > 0 {
		err := s.tokenBlacklist.Revoke(
			ctx,
			jti,
			duration,
		)
		if err != nil {
			return err
		}
	}

	err := s.refreshToken.Revoke(
		ctx,
		refreshTokenHash,
		userID,
	)
	if err != nil {
		return err
	}

	return nil
}
