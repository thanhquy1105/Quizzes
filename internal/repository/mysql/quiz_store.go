package mysql

import (
	"context"
	"errors"
	"time"

	"btaskee-quiz/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

func (s *QuizStore) FindActiveSessionByCode(ctx context.Context, code string) (*model.QuizSession, error) {
	var session model.QuizSession
	now := time.Now()
	err := s.db.WithContext(ctx).
		Where("session_code = ? AND (started_at IS NULL OR started_at <= ?) AND (ended_at IS NULL OR ended_at >= ?)", code, now, now).
		First(&session).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &session, nil
}

func (s *QuizStore) CreateSession(ctx context.Context, session *model.QuizSession) error {
	return s.db.WithContext(ctx).Create(session).Error
}

func (s *QuizStore) AddParticipant(ctx context.Context, participant *model.SessionParticipant) error {
	return s.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "session_id"}, {Name: "user_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"joined_at", "score"}),
		}).
		Create(participant).Error
}

func (s *QuizStore) SaveUserAnswer(ctx context.Context, answer *model.UserAnswer) error {
	return s.db.WithContext(ctx).Create(answer).Error
}

func (s *QuizStore) UpdateParticipantScore(ctx context.Context, sessionID, userID uint64, score int) error {
	return s.db.WithContext(ctx).
		Model(&model.SessionParticipant{}).
		Where("session_id = ? AND user_id = ?", sessionID, userID).
		Update("score", gorm.Expr("score + ?", score)).Error
}

func (s *QuizStore) ListSessions(ctx context.Context) ([]model.QuizSession, error) {
	var sessions []model.QuizSession
	err := s.db.WithContext(ctx).
		Where("deleted_at IS NULL").
		Find(&sessions).Error
	return sessions, err
}

func (s *QuizStore) GetSessionByCode(ctx context.Context, code string) (*model.QuizSession, error) {
	var session model.QuizSession
	err := s.db.WithContext(ctx).Where("session_code = ?", code).First(&session).Error
	return &session, err
}

func (s *QuizStore) IsParticipant(ctx context.Context, sessionID, userID uint64) (bool, error) {
	var count int64
	err := s.db.WithContext(ctx).Model(&model.SessionParticipant{}).
		Where("session_id = ? AND user_id = ?", sessionID, userID).
		Count(&count).Error
	return count > 0, err
}

func (s *QuizStore) GetUserAnswers(ctx context.Context, sessionID, userID uint64) ([]model.UserAnswer, error) {
	var answers []model.UserAnswer
	err := s.db.WithContext(ctx).
		Where("session_id = ? AND user_id = ?", sessionID, userID).
		Find(&answers).Error
	return answers, err
}

func (s *QuizStore) GetUserAnswer(ctx context.Context, sessionID, userID, questionID uint64) (*model.UserAnswer, error) {
	var answer model.UserAnswer
	err := s.db.WithContext(ctx).
		Where("session_id = ? AND user_id = ? AND question_id = ?", sessionID, userID, questionID).
		First(&answer).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &answer, nil
}

func (s *QuizStore) ValidateAnswer(ctx context.Context, quizID, questionID, answerID uint64) (int, bool, error) {
	var answer model.Answer
	err := s.db.WithContext(ctx).
		Joins("JOIN questions ON questions.id = answers.question_id").
		Where("questions.quiz_id = ? AND answers.question_id = ? AND answers.id = ?", quizID, questionID, answerID).
		First(&answer).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, false, nil
		}
		return 0, false, err
	}

	if !answer.IsCorrect {
		return 0, false, nil
	}

	var question model.Question
	if err := s.db.WithContext(ctx).First(&question, questionID).Error; err != nil {
		return 0, false, err
	}

	return question.Point, true, nil
}

func (s *QuizStore) GetParticipantsWithScores(ctx context.Context, sessionCode string) ([]model.RankedEntry, error) {
	var results []model.RankedEntry
	err := s.db.WithContext(ctx).
		Table("session_participants").
		Select("users.username, session_participants.score").
		Joins("JOIN quiz_sessions ON quiz_sessions.id = session_participants.session_id").
		Joins("JOIN users ON users.id = session_participants.user_id").
		Where("quiz_sessions.session_code = ?", sessionCode).
		Scan(&results).Error

	return results, err
}
