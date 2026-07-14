package rediscache

import (
	"context"
	"crypto/subtle"
	"time"

	"github.com/redis/go-redis/v9"
)

type PasswordResetStore struct {
	client *redis.Client
}

func NewPasswordResetStore(address, password string) *PasswordResetStore {
	return &PasswordResetStore{client: redis.NewClient(&redis.Options{
		Addr:     address,
		Password: password,
	})}
}

func (store *PasswordResetStore) Close() error {
	return store.client.Close()
}

func (store *PasswordResetStore) Save(ctx context.Context, email, code string, ttl time.Duration) error {
	return store.client.Set(ctx, resetKey(email), code, ttl).Err()
}

func (store *PasswordResetStore) VerifyAndConsume(ctx context.Context, email, code string) (bool, error) {
	storedCode, err := store.client.Get(ctx, resetKey(email)).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if subtle.ConstantTimeCompare([]byte(storedCode), []byte(code)) != 1 {
		return false, nil
	}
	return true, store.client.Del(ctx, resetKey(email)).Err()
}

func resetKey(email string) string {
	return "password-reset:" + email
}
