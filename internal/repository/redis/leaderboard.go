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

func (r *LeaderboardStore) Add(ctx context.Context, sessionCode, uid string) error {
	return r.rdb.ZAddNX(ctx, lbKey(sessionCode), redis.Z{
		Score:  0,
		Member: uid,
	}).Err()
}

func (r *LeaderboardStore) IncrBy(ctx context.Context, sessionCode, uid string, delta float64) error {
	return r.rdb.ZIncrBy(ctx, lbKey(sessionCode), delta, uid).Err()
}

func (r *LeaderboardStore) GetRanked(ctx context.Context, sessionCode string) ([]model.RankedEntry, error) {
	res, err := r.rdb.ZRevRangeWithScores(ctx, lbKey(sessionCode), 0, -1).Result()
	if err != nil {
		return nil, err
	}
	entries := make([]model.RankedEntry, len(res))
	for i, z := range res {
		entries[i] = model.RankedEntry{
			UID:   z.Member.(string),
			Score: z.Score,
		}
	}
	return entries, nil
}

func (r *LeaderboardStore) Delete(ctx context.Context, sessionCode string) error {
	return r.rdb.Del(ctx, lbKey(sessionCode)).Err()
}
