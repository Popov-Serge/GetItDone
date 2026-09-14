package service

import (
	"context"
	"errors"
	"getitdone/internal/apperror"
	"getitdone/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type FamilyRepository interface {
	Create(ctx context.Context, family *domain.Family) error

	GetByID(
		ctx context.Context,
		familyID uuid.UUID,
		userID uuid.UUID,
	) (*domain.Family, error)

	Update(
		ctx context.Context,
		family *domain.Family,
	) error

	GetFamilies(
		ctx context.Context,
		userID uuid.UUID,
	) ([]*domain.Family, error)

	Delete(
		ctx context.Context,
		familyID uuid.UUID,
		userID uuid.UUID,
	) (bool, error)
}

type FamilyService struct {
	repo FamilyRepository
}

func NewFamilyService(repo FamilyRepository) *FamilyService {
	return &FamilyService{repo: repo}
}

func (f *FamilyService) Create(
	ctx context.Context,
	family *domain.Family,
) error {
	return f.repo.Create(ctx, family)
}

func (f *FamilyService) GetByID(
	ctx context.Context,
	familyID uuid.UUID,
	userID uuid.UUID,
) (*domain.Family, error) {
	family, err := f.repo.GetByID(ctx, familyID, userID)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperror.ErrFamilyNotFound
	}

	if err != nil {
		return nil, err
	}

	return family, nil
}

func (f *FamilyService) Update(
	ctx context.Context,
	family *domain.Family,
) error {
	err := f.repo.Update(ctx, family)

	if errors.Is(err, pgx.ErrNoRows) {
		return apperror.ErrFamilyNotFound
	}

	return err
}

func (f *FamilyService) GetFamilies(
	ctx context.Context,
	userID uuid.UUID,
) ([]*domain.Family, error) {
	return f.repo.GetFamilies(ctx, userID)
}

func (f *FamilyService) Delete(
	ctx context.Context,
	familyID uuid.UUID,
	userID uuid.UUID,
) error {
	deleted, err := f.repo.Delete(ctx, familyID, userID)
	if err != nil {
		return err
	}

	if !deleted {
		return apperror.ErrFamilyNotFound
	}

	return nil
}
