package redis

import (
	"context"
	"fmt"

	"btaskee-quiz/internal/repository"

	"github.com/redis/go-redis/v9"
)

type LeaderboardStore struct {
	rdb *redis.Client
}

func NewLeaderboardStore(rdb *redis.Client) *LeaderboardStore {
	return &LeaderboardStore{rdb: rdb}
}

func lbKey(quizID string) string {
	return fmt.Sprintf("leaderboard:%s", quizID)
}

func (r *LeaderboardStore) Add(ctx context.Context, quizID, uid string) error {
	return r.rdb.ZAddNX(ctx, lbKey(quizID), redis.Z{
		Score:  0,
		Member: uid,
	}).Err()
}

func (r *LeaderboardStore) IncrBy(ctx context.Context, quizID, uid string, delta float64) error {
	return r.rdb.ZIncrBy(ctx, lbKey(quizID), delta, uid).Err()
}

func (r *LeaderboardStore) GetRanked(ctx context.Context, quizID string) ([]repository.RankedEntry, error) {
	res, err := r.rdb.ZRevRangeWithScores(ctx, lbKey(quizID), 0, -1).Result()
	if err != nil {
		return nil, err
	}
	entries := make([]repository.RankedEntry, len(res))
	for i, z := range res {
		entries[i] = repository.RankedEntry{
			UID:   z.Member.(string),
			Score: z.Score,
		}
	}
	return entries, nil
}

func (r *LeaderboardStore) Delete(ctx context.Context, quizID string) error {
	return r.rdb.Del(ctx, lbKey(quizID)).Err()
}
