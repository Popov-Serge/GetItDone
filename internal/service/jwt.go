package service

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JWTService struct {
	secret   string
	duration time.Duration
}

func NewJWTService(secret string, duration time.Duration) *JWTService {
	return &JWTService{
		secret:   secret,
		duration: duration,
	}
}

func (s *JWTService) Generate(userID uuid.UUID) (string, error) {
	now := time.Now()

	claims := jwt.RegisteredClaims{
		ID:        uuid.New().String(),
		Subject:   userID.String(),
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(s.duration)),
	}

	token, err := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		claims,
	).SignedString([]byte(s.secret))

	if err != nil {
		return "", err
	}

	return token, nil
}

func (s *JWTService) Validate(token string) (uuid.UUID, string, time.Time, error) {
	tokenParsed, err := jwt.ParseWithClaims(
		token,
		&jwt.RegisteredClaims{},
		func(token *jwt.Token) (any, error) {
			return []byte(s.secret), nil
		},
		jwt.WithValidMethods([]string{
			jwt.SigningMethodHS256.Alg(),
		}),
	)

	if err != nil {
		return uuid.Nil, "", time.Time{}, err
	}

	if !tokenParsed.Valid {
		return uuid.Nil, "", time.Time{}, errors.New("invalid token")
	}

	claims, ok := tokenParsed.Claims.(*jwt.RegisteredClaims)
	if !ok {
		return uuid.Nil, "", time.Time{}, errors.New("invalid token")
	}

	if claims.ID == "" {
		return uuid.Nil, "", time.Time{}, errors.New("invalid token")
	}

	userID, err := uuid.Parse(claims.Subject)

	if err != nil {
		return uuid.Nil, "", time.Time{}, errors.New("invalid user id")
	}

	return userID, claims.ID, claims.ExpiresAt.Time, nil
}
