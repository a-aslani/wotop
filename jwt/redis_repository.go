package jwt

import (
	"context"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"strconv"
	"strings"
	"time"
)

// RedisRepository is an implementation of the Repository interface
// that uses Redis as the storage backend.
//
// Fields:
//   - rdb: The Redis client used for interacting with the Redis database.
type RedisRepository struct {
	rdb *redis.Client
}

// Ensure RedisRepository implements the Repository interface.
var _ Repository = (*RedisRepository)(nil)

// NewRedisRepository creates a new instance of RedisRepository.
//
// Parameters:
//   - rdb: The Redis client used for interacting with the Redis database.
//
// Returns:
//   - A pointer to a RedisRepository instance.
func NewRedisRepository(rdb *redis.Client) *RedisRepository {
	return &RedisRepository{rdb}
}

// StoreRefreshToken stores a refresh token in Redis.
//
// Parameters:
//   - ctx: The context for the operation.
//   - sub: The subject (user ID) associated with the token.
//   - jti: The unique identifier for the token.
//
// Returns:
//   - An error if the operation fails.
func (r RedisRepository) StoreRefreshToken(ctx context.Context, sub, jti string) error {
	return r.rdb.Set(ctx, fmt.Sprintf("%s:%s", RefreshTokenTableName, jti), sub, 0).Err()
}

// DeleteRefreshToken deletes a refresh token from Redis.
//
// Parameters:
//   - ctx: The context for the operation.
//   - jti: The unique identifier for the token to be deleted.
//
// Returns:
//   - An error if the operation fails.
func (r RedisRepository) DeleteRefreshToken(ctx context.Context, jti string) error {
	return r.rdb.Del(ctx, fmt.Sprintf("%s:%s", RefreshTokenTableName, jti)).Err()
}

// FindRefreshToken retrieves a refresh token from Redis.
//
// Parameters:
//   - ctx: The context for the operation.
//   - jti: The unique identifier for the token to be retrieved.
//
// Returns:
//   - The subject (user ID) associated with the token.
//   - An error if the token is not found or the operation fails.
func (r RedisRepository) FindRefreshToken(ctx context.Context, jti string) (sub string, err error) {
	sub, err = r.rdb.Get(ctx, fmt.Sprintf("%s:%s", RefreshTokenTableName, jti)).Result()
	if errors.Is(err, redis.Nil) {
		err = ErrTokenAlreadyRefreshed
		return
	}
	return
}

// FindAllRefreshTokens retrieves all refresh tokens from Redis.
//
// Parameters:
//   - ctx: The context for the operation.
//
// Returns:
//   - A slice of RefreshToken objects.
//   - An error if the operation fails.
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

// StoreBlockedToken stores a blocked token in Redis.
//
// Parameters:
//   - ctx: The context for the operation.
//   - sub: The subject (user ID) associated with the token.
//   - token: The token to be blocked.
//   - expiresAt: The expiration time of the token in Unix timestamp format.
//
// Returns:
//   - An error if the operation fails.
func (r RedisRepository) StoreBlockedToken(ctx context.Context, sub, token string, expiresAt int64) error {
	return r.rdb.Set(ctx, fmt.Sprintf("%s:%s:%d", BlockedTokenTableName, sub, expiresAt), token, 0).Err()
}

// FindAllBlockedTokens retrieves all blocked tokens from Redis.
//
// Parameters:
//   - ctx: The context for the operation.
//
// Returns:
//   - A slice of blocked tokens as strings.
//   - An error if the operation fails.
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
