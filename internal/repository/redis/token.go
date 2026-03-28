package redis

import (
	"context"
	"fmt"
	"time"

	"btaskee-quiz/pkg/token"

	"github.com/redis/go-redis/v9"
)

const (
	AccessTokenPrefix  = "access_token:%s"
	RefreshTokenPrefix = "refresh_token:%s"
)

type TokenStore struct {
	rdb *redis.Client
}

func NewTokenStore(rdb *redis.Client) *TokenStore {
	return &TokenStore{rdb: rdb}
}

func tokenKey(t string, tokenType token.TokenType) string {
	if tokenType == token.TokenTypeRefreshToken {
		return fmt.Sprintf(RefreshTokenPrefix, t)
	}
	return fmt.Sprintf(AccessTokenPrefix, t)
}

func (r *TokenStore) Save(ctx context.Context, t string, uid string, duration time.Duration, tokenType token.TokenType) error {
	return r.rdb.Set(ctx, tokenKey(t, tokenType), uid, duration).Err()
}

func (r *TokenStore) Exists(ctx context.Context, t string, tokenType token.TokenType) (bool, error) {
	res, err := r.rdb.Exists(ctx, tokenKey(t, tokenType)).Result()
	if err != nil {
		return false, err
	}
	return res > 0, nil
}

func (r *TokenStore) Delete(ctx context.Context, t string, tokenType token.TokenType) error {
	return r.rdb.Del(ctx, tokenKey(t, tokenType)).Err()
}
