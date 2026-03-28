package model

import (
	"time"

	"gorm.io/gorm"
)

type RankedEntry struct {
	UID   string
	Score float64
}

type User struct {
	ID        uint64         `gorm:"primaryKey;autoIncrement"`
	Name      string         `gorm:"size:255"`
	Username  string         `gorm:"size:100;uniqueIndex"`
	CreatedAt time.Time      `gorm:"default:CURRENT_TIMESTAMP"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type Quiz struct {
	ID               uint64         `gorm:"primaryKey;autoIncrement"`
	Title            string         `gorm:"size:255"`
	Description      string         `gorm:"type:text"`
	CreatedAt        time.Time      `gorm:"default:CURRENT_TIMESTAMP"`
	Questions        []Question     `gorm:"foreignKey:QuizID"`
	QuestionCount    int64          `json:"question_count"`
	ParticipantCount int64          `json:"participant_count"`
	DeletedAt        gorm.DeletedAt `gorm:"index"`
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
	Answered  bool           `gorm:"-" json:"answered"`
}

type Answer struct {
	ID         uint64         `gorm:"primaryKey;autoIncrement"`
	QuestionID uint64         `gorm:"index"`
	Content    string         `gorm:"type:text"`
	IsCorrect  bool           `gorm:"default:false"`
	DeletedAt  gorm.DeletedAt `gorm:"index"`
}

type QuizSession struct {
	ID          uint64 `gorm:"primaryKey;autoIncrement"`
	QuizID      uint64 `gorm:"index"`
	SessionCode string `gorm:"size:20;uniqueIndex"`
	Name        string `gorm:"size:255"`
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
	IsCorrect  bool      `gorm:"default:false" json:"-"`
	Score      int       `gorm:"default:0"`
	AnsweredAt time.Time `gorm:"default:CURRENT_TIMESTAMP"`
}
