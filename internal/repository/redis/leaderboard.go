package redis

import (
	"context"
	"fmt"

	"btaskee-quiz/internal/model"

	"github.com/redis/go-redis/v9"
)

type LeaderboardStore struct {
	rdb *redis.Client
}

func NewLeaderboardStore(rdb *redis.Client) *LeaderboardStore {
	return &LeaderboardStore{rdb: rdb}
}

func lbKey(sessionCode string) string {
	return fmt.Sprintf("leaderboard:session:%s", sessionCode)
}

func (s *LeaderboardStore) Add(ctx context.Context, sessionCode, username string) error {
	key := fmt.Sprintf("leaderboard:session:%s", sessionCode)
	return s.rdb.ZAddNX(ctx, key, redis.Z{
		Score:  0,
		Member: username,
	}).Err()
}

func (s *LeaderboardStore) IncrBy(ctx context.Context, sessionCode, username string, delta float64) error {
	key := fmt.Sprintf("leaderboard:session:%s", sessionCode)
	return s.rdb.ZIncrBy(ctx, key, delta, username).Err()
}

func (r *LeaderboardStore) GetRanked(ctx context.Context, sessionCode string) ([]model.RankedEntry, error) {
	zs, err := r.rdb.ZRevRangeWithScores(ctx, lbKey(sessionCode), 0, -1).Result()
	if err != nil {
		return nil, err
	}
	result := make([]model.RankedEntry, 0, len(zs))
	for _, z := range zs {
		result = append(result, model.RankedEntry{
			Username: z.Member.(string),
			Score:    z.Score,
		})
	}
	return result, nil
}

func (r *LeaderboardStore) Delete(ctx context.Context, sessionCode string) error {
	return r.rdb.Del(ctx, lbKey(sessionCode)).Err()
}

func (s *LeaderboardStore) ReloadLeaderboard(ctx context.Context, sessionCode string, entries []model.RankedEntry) error {
	key := lbKey(sessionCode)
	// 1. Clear existing
	if err := s.rdb.Del(ctx, key).Err(); err != nil {
		return err
	}

	// 2. Bulk add
	if len(entries) == 0 {
		return nil
	}

	zs := make([]redis.Z, 0, len(entries))
	for _, e := range entries {
		zs = append(zs, redis.Z{
			Score:  e.Score,
			Member: e.Username,
		})
	}

	return s.rdb.ZAdd(ctx, key, zs...).Err()
}
