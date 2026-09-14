package service

import (
	"getitdone/internal/apperror"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"context"
	"getitdone/internal/domain"
	"testing"
)

type MockFamilyRepository struct {
	mock.Mock
}

func (m *MockFamilyRepository) GetByID(ctx context.Context, familyID uuid.UUID, userID uuid.UUID) (*domain.Family, error) {
	args := m.Called(ctx, familyID, userID)

	var family *domain.Family

	if args.Get(0) != nil {
		family = args.Get(0).(*domain.Family)
	}

	return family, args.Error(1)
}

func (m *MockFamilyRepository) Update(ctx context.Context, family *domain.Family) error {
	args := m.Called(ctx, family)

	return args.Error(0)
}

func (m *MockFamilyRepository) GetFamilies(ctx context.Context, userID uuid.UUID) ([]*domain.Family, error) {
	args := m.Called(ctx, userID)

	var families []*domain.Family

	if args.Get(0) != nil {
		families = args.Get(0).([]*domain.Family)
	}

	return families, args.Error(1)
}

func (m *MockFamilyRepository) Delete(ctx context.Context, familyID uuid.UUID, userID uuid.UUID) (bool, error) {
	args := m.Called(ctx, familyID, userID)
	var deleted bool
	if args.Get(0) != nil {
		deleted = args.Get(0).(bool)
	}
	return deleted, args.Error(1)
}

func (m *MockFamilyRepository) Create(
	ctx context.Context,
	family *domain.Family,
) error {
	args := m.Called(ctx, family)
	return args.Error(0)
}

func TestCreate(t *testing.T) {
	repo := new(MockFamilyRepository)
	service := NewFamilyService(repo)
	family := &domain.Family{
		Name:      "test",
		CreatedBy: uuid.New(),
	}

	repo.On("Create", context.Background(), family).Return(nil)
	err := service.Create(context.Background(), family)

	require.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestGetByID(t *testing.T) {
	repo := new(MockFamilyRepository)
	service := NewFamilyService(repo)
	id := uuid.New()
	userID := uuid.New()
	family := &domain.Family{
		Name:      "test",
		CreatedBy: userID,
		ID:        id,
	}
	repo.On("GetByID", context.Background(), id, userID).Return(family, nil)
	expectedFamily := &domain.Family{
		Name:      "test",
		CreatedBy: userID,
		ID:        id,
	}
	actualFamily, err := service.GetByID(context.Background(), id, userID)
	require.NoError(t, err)
	require.Equal(t, expectedFamily, actualFamily)
	repo.AssertExpectations(t)
}

func TestGetByIDNoFam(t *testing.T) {
	repo := new(MockFamilyRepository)
	service := NewFamilyService(repo)
	id := uuid.New()
	userID := uuid.New()
	repo.On("GetByID", context.Background(), id, userID).Return(nil, pgx.ErrNoRows)
	_, err := service.GetByID(context.Background(), id, userID)
	require.ErrorIs(t, err, apperror.ErrFamilyNotFound)
	repo.AssertExpectations(t)

}

func TestGetByIdDbError(t *testing.T) {
	repo := new(MockFamilyRepository)
	service := NewFamilyService(repo)
	id := uuid.New()
	userID := uuid.New()
	repo.On("GetByID", context.Background(), id, userID).Return(nil, pgx.ErrTooManyRows)
	_, err := service.GetByID(context.Background(), id, userID)
	require.ErrorIs(t, err, pgx.ErrTooManyRows)
	repo.AssertExpectations(t)
}

func TestFamilyServiceUpdate(t *testing.T) {
	repo := new(MockFamilyRepository)
	service := NewFamilyService(repo)
	family := &domain.Family{
		Name: "test",
	}
	repo.On("Update", context.Background(), family).Return(nil)
	err := service.Update(context.Background(), family)
	require.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestFamilyServiceUpdateWithError(t *testing.T) {
	repo := new(MockFamilyRepository)
	service := NewFamilyService(repo)
	family := &domain.Family{
		Name: "test",
	}
	repo.On("Update", context.Background(), family).Return(pgx.ErrNoRows)
	err := service.Update(context.Background(), family)
	require.ErrorIs(t, err, apperror.ErrFamilyNotFound)
	repo.AssertExpectations(t)
}

func TestFamilyServiceUpdateWithDbError(t *testing.T) {
	repo := new(MockFamilyRepository)
	service := NewFamilyService(repo)

	family := &domain.Family{
		Name: "test",
	}

	repo.On(
		"Update",
		context.Background(),
		family,
	).Return(pgx.ErrTooManyRows)

	err := service.Update(
		context.Background(),
		family,
	)

	require.ErrorIs(t, err, pgx.ErrTooManyRows)
	repo.AssertExpectations(t)
}

func TestGetFamilies(t *testing.T) {
	repo := new(MockFamilyRepository)
	userID := uuid.New()
	family1 := &domain.Family{
		Name: "test",
	}
	family2 := &domain.Family{
		Name: "test2",
	}
	service := NewFamilyService(repo)
	repo.On("GetFamilies", context.Background(), userID).Return([]*domain.Family{family1, family2}, nil)
	families, err := service.GetFamilies(context.Background(), userID)
	require.NoError(t, err)
	assert.Equal(t, []*domain.Family{family1, family2}, families)
	repo.AssertExpectations(t)
}

func TestGetFamiliesError(t *testing.T) {
	repo := new(MockFamilyRepository)
	userID := uuid.New()
	service := NewFamilyService(repo)
	repo.On("GetFamilies", context.Background(), userID).Return(nil, pgx.ErrNoRows)
	_, err := service.GetFamilies(context.Background(), userID)
	require.Error(t, err)
	repo.AssertExpectations(t)
}

func TestDelete(t *testing.T) {
	repo := new(MockFamilyRepository)
	familyID := uuid.New()
	userId := uuid.New()
	service := NewFamilyService(repo)
	repo.On("Delete", context.Background(), familyID, userId).Return(true, nil)
	err := service.Delete(context.Background(), familyID, userId)
	require.NoError(t, err)
}

func TestDeleteError(t *testing.T) {
	repo := new(MockFamilyRepository)
	familyID := uuid.New()
	userId := uuid.New()
	service := NewFamilyService(repo)
	repo.On("Delete", context.Background(), familyID, userId).Return(false, pgx.ErrTooManyRows)
	err := service.Delete(context.Background(), familyID, userId)
	require.ErrorIs(t, err, pgx.ErrTooManyRows)
}

func TestDeleteNotFound(t *testing.T) {
	repo := new(MockFamilyRepository)
	familyID := uuid.New()
	userId := uuid.New()
	service := NewFamilyService(repo)

	repo.On("Delete", context.Background(), familyID, userId).Return(false, nil)
	err := service.Delete(context.Background(), familyID, userId)
	require.ErrorIs(t, err, apperror.ErrFamilyNotFound)
	repo.AssertExpectations(t)
}
