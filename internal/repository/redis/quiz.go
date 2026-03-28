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
	return c.store.FindActiveSessionByCode(ctx, code)
}

func (c *QuizCache) CreateSession(ctx context.Context, session *model.QuizSession) error {
	return c.store.CreateSession(ctx, session)
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
	return c.store.ListSessions(ctx)
}

func (c *QuizCache) GetSessionByCode(ctx context.Context, code string) (*model.QuizSession, error) {
	return c.store.GetSessionByCode(ctx, code)
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
