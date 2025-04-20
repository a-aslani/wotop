package wotop_jwt

import (
	"context"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"strconv"
	"strings"
	"time"
)

type RedisRepository struct {
	rdb *redis.Client
}

var _ Repository = (*RedisRepository)(nil)

func NewRedisRepository(rdb *redis.Client) *RedisRepository {
	return &RedisRepository{rdb}
}

func (r RedisRepository) StoreRefreshToken(ctx context.Context, sub, jti string) error {
	return r.rdb.Set(ctx, fmt.Sprintf("%s:%s", RefreshTokenTableName, jti), sub, 0).Err()
}

func (r RedisRepository) DeleteRefreshToken(ctx context.Context, jti string) error {
	return r.rdb.Del(ctx, fmt.Sprintf("%s:%s", RefreshTokenTableName, jti)).Err()
}

func (r RedisRepository) FindRefreshToken(ctx context.Context, jti string) (sub string, err error) {
	sub, err = r.rdb.Get(ctx, fmt.Sprintf("%s:%s", RefreshTokenTableName, jti)).Result()
	if errors.Is(err, redis.Nil) {
		err = ErrTokenAlreadyRefreshed
		return
	}
	return
}

func (r RedisRepository) FindAllRefreshTokens(ctx context.Context) ([]RefreshToken, error) {

	tokens := make([]RefreshToken, 0)

	keys, err := r.rdb.Keys(ctx, fmt.Sprintf("%s:*", RefreshTokenTableName)).Result()
	if err != nil {
		return tokens, err
	}

	for _, key := range keys {

		sub, err := r.rdb.Get(ctx, key).Result()
		if err != nil {
			return tokens, err
		}

		jti := strings.Split(key, ":")[1]
		tokens = append(tokens, RefreshToken{
			Subject: sub,
			JTI:     jti,
		})
	}

	return tokens, nil
}

func (r RedisRepository) StoreBlockedToken(ctx context.Context, sub, token string, expiresAt int64) error {
	return r.rdb.Set(ctx, fmt.Sprintf("%s:%s:%d", BlockedTokenTableName, sub, expiresAt), token, 0).Err()
}

func (r RedisRepository) FindAllBlockedTokens(ctx context.Context) ([]string, error) {

	tokens := make([]string, 0)

	keys, err := r.rdb.Keys(ctx, fmt.Sprintf("%s:*:*", BlockedTokenTableName)).Result()
	if err != nil {
		return tokens, err
	}

	for _, key := range keys {

		spKeys := strings.Split(key, ":")
		expiredAtStr := spKeys[len(spKeys)-1]

		if expiredAtStr != "" {
			expiredAt, err := strconv.ParseInt(expiredAtStr, 10, 64)
			if err != nil {
				continue
			}

			if expiredAt <= time.Now().Unix() {
				r.rdb.Del(ctx, key)
				continue
			}
		}

		token, err := r.rdb.Get(ctx, key).Result()
		if err != nil {
			return tokens, err
		}

		tokens = append(tokens, token)
	}

	return tokens, nil
}
