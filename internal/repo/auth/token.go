package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/faisalhardin/medilink/internal/entity/constant"

	"github.com/go-redis/redis/v8"
)

type TokenStorage interface {
	Set(ctx context.Context, token string, staffID interface{}, expiredAt time.Duration) error
	Get(ctx context.Context, key string) (string, error)
}

type MockRedisClient map[string]interface{}

// type MockRedisClient struct {
// 	store map[string]interface{}
// }

func (r MockRedisClient) Set(ctx context.Context, key string, value interface{}, expireAt time.Duration) error {
	r[key] = value
	return nil
}

func (r MockRedisClient) Get(ctx context.Context, key string) (string, error) {
	if val, ok := r[key]; ok {
		if strVal, ok := val.(string); ok {
			return strVal, nil
		}
		return "", fmt.Errorf("value is not a string: %v", val)
	}
	return "", constant.ErrorNotFound
}

type RedisClient struct {
	red *redis.Client
}

func (r RedisClient) Set(ctx context.Context, key string, value interface{}, expireIn time.Duration) error {
	return r.red.SetNX(ctx, key, value, expireIn).Err()
}

func (r RedisClient) Get(ctx context.Context, key string) (string, error) {
	return r.red.Get(ctx, key).Result()
}

type TokenStore struct {
	Str TokenStorage
}

// func NewTokenStore(store *TokenStore) *TokenStore {
// 	return store
// }

func GenerateOpaqueToken() (string, error) {
	tokenBytes := make([]byte, 32)
	_, err := rand.Read(tokenBytes)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(tokenBytes), nil
}
