package service

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type TokenBlacklist struct {
	redis *redis.Client
}

func NewTokenBlacklist(client *redis.Client) *TokenBlacklist {
	return &TokenBlacklist{
		redis: client,
	}
}

func (tokenBlacklist *TokenBlacklist) Revoke(
	ctx context.Context,
	token string,
	duration time.Duration,
) error {
	return tokenBlacklist.redis.Set(
		ctx,
		"revoked:"+token,
		"1",
		duration,
	).Err()
}

func (tokenBlacklist *TokenBlacklist) IsRevoked(
	ctx context.Context,
	token string,
) (bool, error) {
	exists, err := tokenBlacklist.redis.Exists(
		ctx,
		"revoked:"+token,
	).Result()
	if err != nil {
		return false, err
	}

	return exists == 1, nil
}
