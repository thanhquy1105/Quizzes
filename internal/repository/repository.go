package repository

import (
	"context"
	"time"

	"btaskee-quiz/internal/model"
	"btaskee-quiz/pkg/token"
)

type LeaderboardStore interface {
	Add(ctx context.Context, quizID, uid string) error
	IncrBy(ctx context.Context, quizID, uid string, delta float64) error
	GetRanked(ctx context.Context, quizID string) ([]model.RankedEntry, error)
	Delete(ctx context.Context, quizID string) error
}

type UserStore interface {
	Save(ctx context.Context, user *model.User) error
	Get(ctx context.Context, uid string) (*model.User, error)
	GetByUsername(ctx context.Context, username string) (*model.User, error)
}

type QuizStore interface {
	List(ctx context.Context) ([]model.Quiz, error)
	Get(ctx context.Context, id uint64) (*model.Quiz, error)
}

type TokenStore interface {
	Save(ctx context.Context, token string, uid string, duration time.Duration, tokenType token.TokenType) error
	Exists(ctx context.Context, token string, tokenType token.TokenType) (bool, error)
	Delete(ctx context.Context, token string, tokenType token.TokenType) error
}
