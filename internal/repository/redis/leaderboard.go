package redis

import (
	"context"
	"fmt"

	"btaskee-quiz/internal/model"
	"btaskee-quiz/internal/repository"

	"github.com/redis/go-redis/v9"
)

type LeaderboardStore struct {
	rdb       *redis.Client
	quizStore repository.QuizStore
}

func NewLeaderboardStore(rdb *redis.Client, quizStore repository.QuizStore) *LeaderboardStore {
	return &LeaderboardStore{
		rdb:       rdb,
		quizStore: quizStore,
	}
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

	if len(zs) > 0 {
		result := make([]model.RankedEntry, 0, len(zs))
		for _, z := range zs {
			result = append(result, model.RankedEntry{
				Username: z.Member.(string),
				Score:    z.Score,
			})
		}
		return result, nil
	}

	// Lazy load from DB
	if r.quizStore == nil {
		return nil, nil
	}

	entries, err := r.quizStore.GetParticipantsWithScores(ctx, sessionCode)
	if err != nil {
		return nil, err
	}

	if len(entries) > 0 {
		// Populate Redis in background or wait? Let's wait to ensure consistency for this call
		if err := r.ReloadLeaderboard(ctx, sessionCode, entries); err != nil {
			return nil, err
		}
	}

	return entries, nil
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
