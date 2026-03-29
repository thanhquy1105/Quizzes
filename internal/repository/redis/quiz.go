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

type QuizCache struct {
	rdb   *redis.Client
	store repository.QuizStore
}

func NewQuizCache(rdb *redis.Client, store repository.QuizStore) *QuizCache {
	return &QuizCache{
		rdb:   rdb,
		store: store,
	}
}

func (c *QuizCache) quizKey(id uint64) string {
	return fmt.Sprintf("quiz:detail:%d", id)
}

func (c *QuizCache) validationKey(id uint64) string {
	return fmt.Sprintf("quiz:v:%d", id)
}

func (c *QuizCache) sessionKey(code string) string {
	return fmt.Sprintf("session:detail:%s", code)
}

func (c *QuizCache) sessionListKey() string {
	return "session:list"
}

func (c *QuizCache) Get(ctx context.Context, id uint64) (*model.Quiz, error) {
	key := c.quizKey(id)

	// 1. Try Redis
	val, err := c.rdb.Get(ctx, key).Result()
	if err == nil {
		var quiz model.Quiz
		if err := json.Unmarshal([]byte(val), &quiz); err == nil {
			return &quiz, nil
		}
	}

	// 2. Fallback to DB
	quiz, err := c.store.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	// 3. Cache it (24h TTL)
	if data, err := json.Marshal(quiz); err == nil {
		_ = c.rdb.Set(ctx, key, data, 24*time.Hour).Err()
	}

	// 4. Also cache validation map for O(1) lookup
	vKey := c.validationKey(id)
	vMap := make(map[string]interface{})
	for _, q := range quiz.Questions {
		for _, a := range q.Answers {
			if a.IsCorrect {
				vMap[fmt.Sprintf("q:%d", q.ID)] = fmt.Sprintf("%d:%d", a.ID, q.Point)
				break
			}
		}
	}
	if len(vMap) > 0 {
		_ = c.rdb.HSet(ctx, vKey, vMap).Err()
		_ = c.rdb.Expire(ctx, vKey, 24*time.Hour).Err()
	}

	return quiz, nil
}

func (c *QuizCache) ValidateAnswer(ctx context.Context, quizID, questionID, answerID uint64) (int, bool, error) {
	vKey := c.validationKey(quizID)
	field := fmt.Sprintf("q:%d", questionID)

	val, err := c.rdb.HGet(ctx, vKey, field).Result()
	if err == nil {
		var correctID uint64
		var points int
		if _, err := fmt.Sscanf(val, "%d:%d", &correctID, &points); err == nil {
			if correctID == answerID {
				return points, true, nil
			}
			return 0, false, nil
		}
	}

	// Fallback to underlying store and refresh cache in background
	points, correct, err := c.store.ValidateAnswer(ctx, quizID, questionID, answerID)
	if err == nil {
		// Just trigger a Get() to refresh everything if it was missing
		go func() {
			_, _ = c.Get(context.Background(), quizID)
		}()
	}
	return points, correct, err
}

// Delegation methods
func (c *QuizCache) List(ctx context.Context) ([]model.Quiz, error) {
	return c.store.List(ctx)
}

func (c *QuizCache) FindActiveSessionByCode(ctx context.Context, code string) (*model.QuizSession, error) {
	// Use cached GetSessionByCode
	session, err := c.GetSessionByCode(ctx, code)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, nil
	}

	// Double check active status in Go logic (more efficient than complex Redis keys)
	now := time.Now()
	if (session.StartedAt == nil || session.StartedAt.Before(now) || session.StartedAt.Equal(now)) &&
		(session.EndedAt == nil || session.EndedAt.After(now) || session.EndedAt.Equal(now)) {
		return session, nil
	}

	return nil, nil
}

func (c *QuizCache) CreateSession(ctx context.Context, session *model.QuizSession) error {
	if err := c.store.CreateSession(ctx, session); err != nil {
		return err
	}
	// Invalidate session list
	_ = c.rdb.Del(ctx, c.sessionListKey()).Err()
	return nil
}

func (c *QuizCache) AddParticipant(ctx context.Context, participant *model.SessionParticipant) error {
	return c.store.AddParticipant(ctx, participant)
}

func (c *QuizCache) SaveUserAnswer(ctx context.Context, answer *model.UserAnswer) error {
	return c.store.SaveUserAnswer(ctx, answer)
}

func (c *QuizCache) UpdateParticipantScore(ctx context.Context, sessionID, userID uint64, score int) error {
	return c.store.UpdateParticipantScore(ctx, sessionID, userID, score)
}

func (c *QuizCache) ListSessions(ctx context.Context) ([]model.QuizSession, error) {
	key := c.sessionListKey()

	// 1. Try Redis
	val, err := c.rdb.Get(ctx, key).Result()
	if err == nil {
		var sessions []model.QuizSession
		if err := json.Unmarshal([]byte(val), &sessions); err == nil {
			return sessions, nil
		}
	}

	// 2. Fallback to DB
	sessions, err := c.store.ListSessions(ctx)
	if err != nil {
		return nil, err
	}

	// 3. Cache it (1h TTL)
	if data, err := json.Marshal(sessions); err == nil {
		_ = c.rdb.Set(ctx, key, data, 1*time.Hour).Err()
	}

	return sessions, nil
}

func (c *QuizCache) GetSessionByCode(ctx context.Context, code string) (*model.QuizSession, error) {
	key := c.sessionKey(code)

	// 1. Try Redis
	val, err := c.rdb.Get(ctx, key).Result()
	if err == nil {
		var session model.QuizSession
		if err := json.Unmarshal([]byte(val), &session); err == nil {
			return &session, nil
		}
	}

	// 2. Fallback to DB
	session, err := c.store.GetSessionByCode(ctx, code)
	if err != nil {
		return nil, err
	}

	// 3. Cache it (24h TTL)
	if data, err := json.Marshal(session); err == nil {
		_ = c.rdb.Set(ctx, key, data, 24*time.Hour).Err()
	}

	return session, nil
}

func (c *QuizCache) IsParticipant(ctx context.Context, sessionID, userID uint64) (bool, error) {
	return c.store.IsParticipant(ctx, sessionID, userID)
}

func (c *QuizCache) GetUserAnswers(ctx context.Context, sessionID, userID uint64) ([]model.UserAnswer, error) {
	return c.store.GetUserAnswers(ctx, sessionID, userID)
}

func (c *QuizCache) GetUserAnswer(ctx context.Context, sessionID, userID, questionID uint64) (*model.UserAnswer, error) {
	return c.store.GetUserAnswer(ctx, sessionID, userID, questionID)
}

func (c *QuizCache) GetParticipantsWithScores(ctx context.Context, sessionCode string) ([]model.RankedEntry, error) {
	return c.store.GetParticipantsWithScores(ctx, sessionCode)
}

func (c *QuizCache) ListActiveSessions(ctx context.Context) ([]model.QuizSession, error) {
	return c.store.ListActiveSessions(ctx)
}

func (c *QuizCache) Transaction(ctx context.Context, fn func(repository.QuizStore) error) error {
	return c.store.Transaction(ctx, func(txStore repository.QuizStore) error {
		txCache := &QuizCache{
			rdb:   c.rdb,
			store: txStore,
		}
		return fn(txCache)
	})
}

func (c *QuizCache) answeredKey(sessionID, userID uint64) string {
	return fmt.Sprintf("session:answered:%d:%d", sessionID, userID)
}

func (c *QuizCache) CheckAndSetAnswered(ctx context.Context, sessionID, userID, questionID uint64) (bool, error) {
	key := c.answeredKey(sessionID, userID)
	added, err := c.rdb.SAdd(ctx, key, questionID).Result()
	if err != nil {
		return false, err
	}
	// Expire after 24h to cleanup
	_ = c.rdb.Expire(ctx, key, 24*time.Hour).Err()
	return added > 0, nil
}

func (c *QuizCache) RemoveAnsweredCache(ctx context.Context, sessionID, userID, questionID uint64) error {
	key := c.answeredKey(sessionID, userID)
	return c.rdb.SRem(ctx, key, questionID).Err()
}

func (c *QuizCache) SyncAnsweredCache(ctx context.Context, sessionID, userID uint64, questionIDs []uint64) error {
	if len(questionIDs) == 0 {
		return nil
	}
	key := c.answeredKey(sessionID, userID)
	// SAdd accepts multiple values
	interfaces := make([]interface{}, len(questionIDs))
	for i, v := range questionIDs {
		interfaces[i] = v
	}
	err := c.rdb.SAdd(ctx, key, interfaces...).Err()
	if err != nil {
		return err
	}
	return c.rdb.Expire(ctx, key, 24*time.Hour).Err()
}
