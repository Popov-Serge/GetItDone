package service

import (
	"context"
	"getitdone/internal/apperror"
	"getitdone/internal/domain"
	"getitdone/internal/validation"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByPhone(ctx context.Context, phone string) (*domain.User, error) {
	args := m.Called(ctx, phone)
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) ExistsByPhoneOrEmail(ctx context.Context, phone string, email string) (bool, bool, error) {
	args := m.Called(ctx, phone, email)
	return args.Bool(0), args.Bool(1), args.Error(2)
}

type MockRefreshRepository struct {
	mock.Mock
}

func (m *MockRefreshRepository) FindByTokenHash(ctx context.Context, tokenHash string) (*domain.RefreshToken, error) {
	args := m.Called(ctx, tokenHash)
	return args.Get(0).(*domain.RefreshToken), args.Error(1)
}

func (m *MockRefreshRepository) Revoke(ctx context.Context, tokenHash string, userID uuid.UUID) error {
	args := m.Called(ctx, tokenHash, userID)
	return args.Error(0)
}

type MockRefreshTokenService struct {
	mock.Mock
}

func (m *MockRefreshTokenService) Generate(ctx context.Context, userID uuid.UUID) (string, error) {
	args := m.Called(ctx, userID)
	return args.String(0), args.Error(1)
}

func (m *MockRefreshTokenService) Rotate(ctx context.Context, tokenHash string, userID uuid.UUID) (string, error) {
	args := m.Called(ctx, tokenHash, userID)
	return args.String(0), args.Error(1)
}

type MockTokenBlacklist struct {
	mock.Mock
}

type MockTokenService struct {
	mock.Mock
}

func (m *MockTokenService) Generate(userID uuid.UUID) (string, error) {
	args := m.Called(userID)
	return args.String(0), args.Error(1)
}

func (m *MockTokenBlacklist) Revoke(ctx context.Context, token string, duration time.Duration) error {
	args := m.Called(ctx, token, duration)
	return args.Error(0)
}

func TestRegisterSuccess(t *testing.T) {
	input := RegisterInput{
		Name:     "test",
		Phone:    "+79184649701",
		Password: "Gillett@3000",
		Email:    "a@mail.ru",
		Timezone: "Europe/London",
	}
	usersRepository := new(MockUserRepository)
	usersRepository.On(
		"ExistsByPhoneOrEmail",
		context.Background(),
		input.Phone,
		input.Email,
	).Return(false, false, nil)

	authService := &AuthService{
		users:               usersRepository,
		refreshToken:        new(MockRefreshRepository),
		phoneValidator:      new(validation.PhoneValidator),
		tokenService:        new(MockTokenService),
		refreshTokenService: new(MockRefreshTokenService),
		tokenBlacklist:      new(MockTokenBlacklist),
	}

	usersRepository.On(
		"Create",
		context.Background(),
		mock.MatchedBy(func(user *domain.User) bool {
			return user.Name == input.Name &&
				user.Phone == input.Phone &&
				user.Email == input.Email &&
				user.Timezone == input.Timezone &&
				bcrypt.CompareHashAndPassword(
					[]byte(user.PasswordHash),
					[]byte(input.Password),
				) == nil
		}),
	).Return(nil)
	userCreated, err := authService.Register(context.Background(), input)

	require.NoError(t, err)
	assert.Equal(t, input.Name, userCreated.Name)
	assert.Equal(t, input.Phone, userCreated.Phone)
	assert.Equal(t, input.Email, userCreated.Email)
	assert.Equal(t, input.Timezone, userCreated.Timezone)
	require.NoError(
		t,
		bcrypt.CompareHashAndPassword(
			[]byte(userCreated.PasswordHash),
			[]byte(input.Password),
		),
	)
	usersRepository.AssertExpectations(t)
}

func TestRegisterInvalidPhone(t *testing.T) {
	input := RegisterInput{
		Name:     "test",
		Phone:    "123",
		Password: "Gillett@3000",
		Email:    "a@mail.ru",
		Timezone: "Europe/London",
	}
	usersRepository := new(MockUserRepository)

	authService := &AuthService{
		users:               usersRepository,
		refreshToken:        new(MockRefreshRepository),
		phoneValidator:      new(validation.PhoneValidator),
		tokenService:        new(MockTokenService),
		refreshTokenService: new(MockRefreshTokenService),
		tokenBlacklist:      new(MockTokenBlacklist),
	}

	_, err := authService.Register(context.Background(), input)

	require.Error(t, err)
	assert.Equal(t, err, apperror.ErrInvalidPhone)

	usersRepository.AssertExpectations(t)
}

func TestRegisterPhoneAmdEmailExistsError(t *testing.T) {
	input := RegisterInput{
		Name:     "test",
		Phone:    "+79184649701",
		Password: "Gillett@3000",
		Email:    "a@mail.ru",
		Timezone: "Europe/London",
	}
	usersRepository := new(MockUserRepository)

	usersRepository.On(
		"ExistsByPhoneOrEmail",
		context.Background(),
		input.Phone,
		input.Email,
	).Return(false, false, pgx.ErrTooManyRows)

	authService := &AuthService{
		users:               usersRepository,
		refreshToken:        new(MockRefreshRepository),
		phoneValidator:      new(validation.PhoneValidator),
		tokenService:        new(MockTokenService),
		refreshTokenService: new(MockRefreshTokenService),
		tokenBlacklist:      new(MockTokenBlacklist),
	}

	_, err := authService.Register(context.Background(), input)

	require.Error(t, err)
	assert.Equal(t, err, pgx.ErrTooManyRows)
	usersRepository.AssertExpectations(t)
}

func TestRegisterPhoneAlreadyExists(t *testing.T) {
	input := RegisterInput{
		Name:     "test",
		Phone:    "+79184649701",
		Password: "Gillett@3000",
		Email:    "a@mail.ru",
		Timezone: "Europe/London",
	}
	usersRepository := new(MockUserRepository)

	usersRepository.On(
		"ExistsByPhoneOrEmail",
		context.Background(),
		input.Phone,
		input.Email,
	).Return(true, false, nil)

	authService := &AuthService{
		users:               usersRepository,
		refreshToken:        new(MockRefreshRepository),
		phoneValidator:      new(validation.PhoneValidator),
		tokenService:        new(MockTokenService),
		refreshTokenService: new(MockRefreshTokenService),
		tokenBlacklist:      new(MockTokenBlacklist),
	}

	_, err := authService.Register(context.Background(), input)

	require.Equal(t, &apperror.RegistrationConflictError{map[string][]string{
		"phone": {"already_exists"},
	}}, err)
	usersRepository.AssertExpectations(t)
}
func TestRegisterEmailAlreadyExists(t *testing.T) {
	input := RegisterInput{
		Name:     "test",
		Phone:    "+79184649701",
		Password: "Gillett@3000",
		Email:    "a@mail.ru",
		Timezone: "Europe/London",
	}
	usersRepository := new(MockUserRepository)

	usersRepository.On(
		"ExistsByPhoneOrEmail",
		context.Background(),
		input.Phone,
		input.Email,
	).Return(false, true, nil)

	authService := &AuthService{
		users:               usersRepository,
		refreshToken:        new(MockRefreshRepository),
		phoneValidator:      new(validation.PhoneValidator),
		tokenService:        new(MockTokenService),
		refreshTokenService: new(MockRefreshTokenService),
		tokenBlacklist:      new(MockTokenBlacklist),
	}

	_, err := authService.Register(context.Background(), input)

	require.Equal(t, &apperror.RegistrationConflictError{map[string][]string{
		"email": {"already_exists"},
	}}, err)
	usersRepository.AssertExpectations(t)
}
func TestRegisterPhoneAndEmailAlreadyExists(t *testing.T) {
	input := RegisterInput{
		Name:     "test",
		Phone:    "+79184649701",
		Password: "Gillett@3000",
		Email:    "a@mail.ru",
		Timezone: "Europe/London",
	}
	usersRepository := new(MockUserRepository)

	usersRepository.On(
		"ExistsByPhoneOrEmail",
		context.Background(),
		input.Phone,
		input.Email,
	).Return(true, true, nil)

	authService := &AuthService{
		users:               usersRepository,
		refreshToken:        new(MockRefreshRepository),
		phoneValidator:      new(validation.PhoneValidator),
		tokenService:        new(MockTokenService),
		refreshTokenService: new(MockRefreshTokenService),
		tokenBlacklist:      new(MockTokenBlacklist),
	}

	_, err := authService.Register(context.Background(), input)

	require.Equal(t, &apperror.RegistrationConflictError{map[string][]string{
		"phone": {"already_exists"},
		"email": {"already_exists"},
	}}, err)
	usersRepository.AssertExpectations(t)
}
func TestRegisterCreateError(t *testing.T) {
	input := RegisterInput{
		Name:     "test",
		Phone:    "+79184649701",
		Password: "Gillett@3000",
		Email:    "a@mail.ru",
		Timezone: "Europe/London",
	}

	usersRepository := new(MockUserRepository)

	usersRepository.On(
		"ExistsByPhoneOrEmail",
		context.Background(),
		input.Phone,
		input.Email,
	).Return(false, false, nil)

	usersRepository.On("Create", context.Background(), mock.Anything).Return(pgx.ErrNoRows)

	authService := &AuthService{
		users:               usersRepository,
		refreshToken:        new(MockRefreshRepository),
		phoneValidator:      new(validation.PhoneValidator),
		tokenService:        new(MockTokenService),
		refreshTokenService: new(MockRefreshTokenService),
		tokenBlacklist:      new(MockTokenBlacklist),
	}

	_, err := authService.Register(context.Background(), input)

	require.Error(t, err)
	assert.Equal(t, err, pgx.ErrNoRows)
}

// Login
func TestLoginSuccess(t *testing.T) {

}
func TestLoginInvalidPhone(t *testing.T) {

}
func TestLoginUserNotFound(t *testing.T) {

}
func TestLoginRepositoryError(t *testing.T) {

}
func TestLoginInvalidPassword(t *testing.T) {

}
func TestLoginAccessTokenError(t *testing.T) {

}
func TestLoginRefreshTokenError(t *testing.T) {

}

// Refresh
func TestRefreshSuccess(t *testing.T) {

}
func TestRefreshTokenNotFound(t *testing.T) {

}
func TestRefreshRepositoryError(t *testing.T) {

}
func TestRefreshRevoked(t *testing.T) {

}
func TestRefreshExpired(t *testing.T) {

}
func TestRefreshAccessTokenError(t *testing.T) {

}
func TestRefreshRotateInvalidToken(t *testing.T) {

}
func TestRefreshRotateError(t *testing.T) {

}

// Logout
func TestLogoutSuccess(t *testing.T) {

}
func TestLogoutBlacklistError(t *testing.T) {

}
func TestLogoutRefreshRevokeError(t *testing.T) {

}
func TestLogoutExpiredAccessToken(t *testing.T) {

}
