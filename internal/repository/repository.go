package repository

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        uint64         `gorm:"primaryKey;autoIncrement"`
	Name      string         `gorm:"size:255"`
	Username  string         `gorm:"size:100;uniqueIndex"`
	CreatedAt time.Time      `gorm:"default:CURRENT_TIMESTAMP"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type UserStore interface {
	Save(ctx context.Context, user *User) error
	Get(ctx context.Context, uid string) (*User, error)
}

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

type Quiz struct {
	ID          uint64         `gorm:"primaryKey;autoIncrement"`
	Title       string         `gorm:"size:255"`
	Description string         `gorm:"type:text"`
	CreatedAt   time.Time      `gorm:"default:CURRENT_TIMESTAMP"`
	Questions   []Question     `gorm:"foreignKey:QuizID"`
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

type Question struct {
	ID        uint64         `gorm:"primaryKey;autoIncrement"`
	QuizID    uint64         `gorm:"index"`
	Content   string         `gorm:"type:text"`
	Point     int            `gorm:"default:10"`
	TimeLimit int            `gorm:"default:10"`
	CreatedAt time.Time      `gorm:"default:CURRENT_TIMESTAMP"`
	Answers   []Answer       `gorm:"foreignKey:QuestionID"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type Answer struct {
	ID         uint64         `gorm:"primaryKey;autoIncrement"`
	QuestionID uint64         `gorm:"index"`
	Content    string         `gorm:"type:text"`
	IsCorrect  bool           `gorm:"default:false"`
	DeletedAt  gorm.DeletedAt `gorm:"index"`
}

type QuizSession struct {
	ID          uint64         `gorm:"primaryKey;autoIncrement"`
	QuizID      uint64         `gorm:"index"`
	SessionCode string         `gorm:"size:20;uniqueIndex"`
	Status      string         `gorm:"size:20;default:'waiting'"` // waiting, running, finished
	StartedAt   *time.Time
	EndedAt     *time.Time
	CreatedAt   time.Time      `gorm:"default:CURRENT_TIMESTAMP"`
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

type SessionParticipant struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement"`
	SessionID uint64    `gorm:"uniqueIndex:idx_session_user"`
	UserID    uint64    `gorm:"uniqueIndex:idx_session_user"`
	Score     int       `gorm:"default:0"`
	JoinedAt  time.Time `gorm:"default:CURRENT_TIMESTAMP"`
}

type UserAnswer struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement"`
	SessionID  uint64    `gorm:"uniqueIndex:idx_answer"`
	UserID     uint64    `gorm:"uniqueIndex:idx_answer"`
	QuestionID uint64    `gorm:"uniqueIndex:idx_answer"`
	AnswerID   uint64    `gorm:"index"`
	IsCorrect  bool      `gorm:"default:false"`
	Score      int       `gorm:"default:0"`
	AnsweredAt time.Time `gorm:"default:CURRENT_TIMESTAMP"`
}
