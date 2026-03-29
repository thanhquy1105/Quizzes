package repository

import (
	"context"
	"time"

	"btaskee-quiz/internal/model"
	"btaskee-quiz/pkg/token"
)

type LeaderboardStore interface {
	Add(ctx context.Context, sessionCode, username string) error
	IncrBy(ctx context.Context, sessionCode, username string, delta float64) error
	GetRanked(ctx context.Context, sessionCode string) ([]model.RankedEntry, error)
	Delete(ctx context.Context, sessionCode string) error
	ReloadLeaderboard(ctx context.Context, sessionCode string, entries []model.RankedEntry) error
}

type UserStore interface {
	Save(ctx context.Context, user *model.User) error
	Get(ctx context.Context, username string) (*model.User, error)
	GetByUsername(ctx context.Context, username string) (*model.User, error)
}

type QuizStore interface {
	List(ctx context.Context) ([]model.Quiz, error)
	Get(ctx context.Context, id uint64) (*model.Quiz, error)
	FindActiveSessionByCode(ctx context.Context, code string) (*model.QuizSession, error)
	CreateSession(ctx context.Context, session *model.QuizSession) error
	AddParticipant(ctx context.Context, participant *model.SessionParticipant) error
	SaveUserAnswer(ctx context.Context, answer *model.UserAnswer) error
	UpdateParticipantScore(ctx context.Context, sessionID, userID uint64, score int) error
	ListSessions(ctx context.Context) ([]model.QuizSession, error)
	GetSessionByCode(ctx context.Context, code string) (*model.QuizSession, error)
	IsParticipant(ctx context.Context, sessionID, userID uint64) (bool, error)
	GetUserAnswers(ctx context.Context, sessionID, userID uint64) ([]model.UserAnswer, error)
	GetUserAnswer(ctx context.Context, sessionID, userID, questionID uint64) (*model.UserAnswer, error)
	ValidateAnswer(ctx context.Context, quizID, questionID, answerID uint64) (int, bool, error)
	GetParticipantsWithScores(ctx context.Context, sessionCode string) ([]model.RankedEntry, error)
	ListActiveSessions(ctx context.Context) ([]model.QuizSession, error)
}

type TokenStore interface {
	Save(ctx context.Context, token string, uid string, duration time.Duration, tokenType token.TokenType) error
	Exists(ctx context.Context, token string, tokenType token.TokenType) (bool, error)
	Delete(ctx context.Context, token string, tokenType token.TokenType) error
}
