package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"time"

	"github.com/google/uuid"
)

type RefreshTokenRepository interface {
	Create(
		ctx context.Context,
		userID uuid.UUID,
		tokenHash string,
		expiresAt time.Time,
	) error
	Rotate(
		ctx context.Context,
		oldTokenHash string,
		newTokenHash string,
		userID uuid.UUID,
		expiresAt time.Time,
	) error
	Revoke(
		ctx context.Context,
		tokenHash string,
		userID uuid.UUID,
	) error
}

type RefreshTokenService struct {
	repository RefreshTokenRepository
	duration   time.Duration
}

func NewRefreshTokenService(repository RefreshTokenRepository, duration time.Duration) *RefreshTokenService {
	return &RefreshTokenService{
		repository: repository,
		duration:   duration,
	}
}

func (s *RefreshTokenService) generate() (
	string,
	string,
	time.Time,
	error,
) {
	bytes := make([]byte, 32)

	if _, err := rand.Read(bytes); err != nil {
		return "", "", time.Time{}, err
	}

	token := base64.RawURLEncoding.EncodeToString(bytes)

	hash := sha256.Sum256([]byte(token))
	tokenHash := hex.EncodeToString(hash[:])

	expiresAt := time.Now().Add(s.duration)

	return token, tokenHash, expiresAt, nil
}

func (s *RefreshTokenService) Generate(
	ctx context.Context,
	userID uuid.UUID,
) (string, error) {
	token, tokenHash, expiresAt, err := s.generate()
	if err != nil {
		return "", err
	}

	err = s.repository.Create(
		ctx,
		userID,
		tokenHash,
		expiresAt,
	)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (s *RefreshTokenService) Rotate(
	ctx context.Context,
	oldTokenHash string,
	userID uuid.UUID,
) (string, error) {
	token, tokenHash, expiresAt, err := s.generate()
	if err != nil {
		return "", err
	}

	err = s.repository.Rotate(
		ctx,
		oldTokenHash,
		tokenHash,
		userID,
		expiresAt,
	)
	if err != nil {
		return "", err
	}

	return token, nil
}
