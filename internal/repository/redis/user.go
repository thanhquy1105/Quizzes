package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"btaskee-quiz/internal/model"
	"btaskee-quiz/internal/repository"

	"github.com/redis/go-redis/v9"
)

type UserCache struct {
	rdb   *redis.Client
	store repository.UserStore
}

func NewUserCache(rdb *redis.Client, store repository.UserStore) *UserCache {
	return &UserCache{
		rdb:   rdb,
		store: store,
	}
}

func (c *UserCache) profileKey(username string) string {
	return fmt.Sprintf("user:profile:%s", username)
}

func (c *UserCache) Save(ctx context.Context, user *model.User) error {
	if err := c.store.Save(ctx, user); err != nil {
		return err
	}
	// Invalidate cache on save
	return c.rdb.Del(ctx, c.profileKey(user.Username)).Err()
}

func (c *UserCache) Get(ctx context.Context, username string) (*model.User, error) {
	return c.GetByUsername(ctx, username)
}

func (c *UserCache) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	key := c.profileKey(username)

	// 1. Try Redis
	val, err := c.rdb.Get(ctx, key).Result()
	if err == nil {
		var user model.User
		if err := json.Unmarshal([]byte(val), &user); err == nil {
			return &user, nil
		}
	}

	// 2. Fallback to DB
	user, err := c.store.GetByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	// 3. Cache it (1h TTL)
	if data, err := json.Marshal(user); err == nil {
		_ = c.rdb.Set(ctx, key, data, 1*time.Hour).Err()
	}

	return user, nil
}
