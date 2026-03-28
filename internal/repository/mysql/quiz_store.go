package mysql

import (
	"context"

	"btaskee-quiz/internal/model"

	"gorm.io/gorm"
)

type QuizStore struct {
	db *gorm.DB
}

func NewQuizStore(db *gorm.DB) *QuizStore {
	return &QuizStore{db: db}
}

func (s *QuizStore) List(ctx context.Context) ([]model.Quiz, error) {
	var quizzes []model.Quiz
	err := s.db.WithContext(ctx).
		Table("quizzes").
		Select("quizzes.*, " +
			"(SELECT COUNT(*) FROM questions WHERE questions.quiz_id = quizzes.id AND questions.deleted_at IS NULL) as question_count, " +
			"(SELECT COUNT(DISTINCT sp.user_id) FROM session_participants sp JOIN quiz_sessions qs ON sp.session_id = qs.id WHERE qs.quiz_id = quizzes.id AND qs.deleted_at IS NULL) as participant_count").
		Where("quizzes.deleted_at IS NULL").
		Scan(&quizzes).Error
	if err != nil {
		return nil, err
	}
	return quizzes, nil
}

func (s *QuizStore) Get(ctx context.Context, id uint64) (*model.Quiz, error) {
	var quiz model.Quiz
	err := s.db.WithContext(ctx).Preload("Questions.Answers").First(&quiz, id).Error
	if err != nil {
		return nil, err
	}
	return &quiz, nil
}
