package quiz

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type RankedEntry struct {
	UID   string
	Score float64
}

type LeaderboardStore interface {
	Add(ctx context.Context, quizID, uid string) error
	IncrBy(ctx context.Context, quizID, uid string, delta float64) error
	GetRanked(ctx context.Context, quizID string) ([]RankedEntry, error)
	Delete(ctx context.Context, quizID string) error
}

type RedisLeaderboard struct {
	rdb *redis.Client
}

func NewRedisLeaderboard(rdb *redis.Client) *RedisLeaderboard {
	return &RedisLeaderboard{rdb: rdb}
}

func lbKey(quizID string) string {
	return fmt.Sprintf("leaderboard:%s", quizID)
}

func (r *RedisLeaderboard) Add(ctx context.Context, quizID, uid string) error {
	return r.rdb.ZAddNX(ctx, lbKey(quizID), redis.Z{
		Score:  0,
		Member: uid,
	}).Err()
}

func (r *RedisLeaderboard) IncrBy(ctx context.Context, quizID, uid string, delta float64) error {
	return r.rdb.ZIncrBy(ctx, lbKey(quizID), delta, uid).Err()
}

func (r *RedisLeaderboard) GetRanked(ctx context.Context, quizID string) ([]RankedEntry, error) {
	res, err := r.rdb.ZRevRangeWithScores(ctx, lbKey(quizID), 0, -1).Result()
	if err != nil {
		return nil, err
	}
	entries := make([]RankedEntry, len(res))
	for i, z := range res {
		entries[i] = RankedEntry{
			UID:   z.Member.(string),
			Score: z.Score,
		}
	}
	return entries, nil
}

func (r *RedisLeaderboard) Delete(ctx context.Context, quizID string) error {
	return r.rdb.Del(ctx, lbKey(quizID)).Err()
}
