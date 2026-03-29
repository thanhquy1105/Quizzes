package quiz

import (
	"context"
	"encoding/json"
	"errors"

	"btaskee-quiz/internal/model"
	"btaskee-quiz/internal/repository"
	"btaskee-quiz/internal/server"
	"btaskee-quiz/internal/server/proto"

	"go.uber.org/zap"
)

type JoinReq struct {
	SessionCode string `json:"session_code"`
}

func (s *QuizServer) handleJoin(ctx *server.Context) {
	s.Server.Debug("handleJoin started", zap.String("body", string(ctx.Body())))

	username, err := s.getAuthUsername(ctx)
	if err != nil {
		ctx.WriteErrorAndStatus(err, proto.StatusError)
		return
	}

	var req JoinReq
	if err := json.Unmarshal(ctx.Body(), &req); err != nil {
		s.Server.Error("handleJoin unmarshal error", zap.Error(err))
		ctx.WriteErr(err)
		return
	}

	if req.SessionCode == "" {
		ctx.WriteErrorAndStatus(errors.New("session_code is required"), proto.StatusError)
		return
	}

	// Find active session by code (time-window validated)
	dbSession, err := s.quizStore.FindActiveSessionByCode(context.Background(), req.SessionCode)
	if err != nil {
		s.Server.Error("handleJoin: error finding session", zap.Error(err))
		ctx.WriteErr(err)
		return
	}
	if dbSession == nil {
		ctx.WriteErrorAndStatus(errors.New("session not found or not active"), proto.StatusError)
		return
	}

	// Fetch quiz (QuizID comes from the session record)
	quiz, err := s.quizStore.Get(context.Background(), dbSession.QuizID)
	if err != nil {
		s.Server.Error("handleJoin: quiz not found", zap.Uint64("quizID", dbSession.QuizID), zap.Error(err))
		ctx.WriteErrorAndStatus(errors.New("quiz not found"), proto.StatusError)
		return
	}

	// Fetch user (Name comes from DB, not from client)
	user, err := s.userStore.GetByUsername(context.Background(), username)
	if err != nil {
		s.Server.Error("handleJoin: user not found", zap.String("username", username), zap.Error(err))
		ctx.WriteErrorAndStatus(errors.New("user not found"), proto.StatusError)
		return
	}

	// Check participation & resume logic
	isPart, err := s.quizStore.IsParticipant(context.Background(), dbSession.ID, user.ID)
	if err != nil {
		s.Server.Error("handleJoin: error checking participation", zap.Error(err))
		ctx.WriteErr(err)
		return
	}

	var currentScore float64

	if isPart {
		// Resume: return ALL questions, mark answered ones so frontend can lock them
		userAnswers, err := s.quizStore.GetUserAnswers(context.Background(), dbSession.ID, user.ID)
		if err != nil {
			s.Server.Error("handleJoin: error fetching user answers", zap.Error(err))
			ctx.WriteErr(err)
			return
		}

		answeredMap := make(map[uint64]bool)
		questionIDs := make([]uint64, 0, len(userAnswers))
		for _, ans := range userAnswers {
			answeredMap[ans.QuestionID] = true
			questionIDs = append(questionIDs, ans.QuestionID)
			currentScore += float64(ans.Score)
		}

		// Warm up Redis checklist for this user
		if err := s.quizStore.SyncAnsweredCache(context.Background(), dbSession.ID, user.ID, questionIDs); err != nil {
			s.Server.Error("handleJoin: failed to sync answered cache", zap.Error(err))
			// Non-critical: continue even if cache sync fails
		}

		for i := range quiz.Questions {
			if answeredMap[quiz.Questions[i].ID] {
				quiz.Questions[i].Answered = true
			}
		}

		s.Server.Info("handleJoin: resuming session", zap.Uint64("userID", user.ID), zap.Int("answered", len(userAnswers)))
	} else {
		// First join: record participant with transaction
		err = s.quizStore.Transaction(context.Background(), func(txStore repository.QuizStore) error {
			return txStore.AddParticipant(context.Background(), &model.SessionParticipant{
				SessionID: dbSession.ID,
				UserID:    user.ID,
			})
		})
		if err != nil {
			s.Server.Error("handleJoin: error adding participant", zap.Error(err))
			ctx.WriteErr(err)
			return
		}
	}

	// 7. Join in-memory session (manager handles participants and connections)
	s.Manager.Join(dbSession.SessionCode, quiz.ID, dbSession.ID, username, user.Name, user.ID, currentScore, ctx.Conn())
	s.Server.Metrics().JoinQuizInc()

	s.Server.Debug("handleJoin: joined", zap.String("username", username), zap.String("code", dbSession.SessionCode))

	quizData, err := json.Marshal(quiz)
	if err != nil {
		s.Server.Error("handleJoin: marshal quiz error", zap.Error(err))
		ctx.WriteErr(err)
		return
	}
	ctx.Write(quizData)
	s.Manager.BroadcastLeaderboard(dbSession.SessionCode)
	s.Server.Debug("handleJoin finished")
}

type AnswerReq struct {
	SessionCode string `json:"session_code"`
	QuestionID  uint64 `json:"question_id"`
	AnswerID    uint64 `json:"answer_id"`
}

func (s *QuizServer) handleAnswer(ctx *server.Context) {
	username, err := s.getAuthUsername(ctx)
	if err != nil {
		s.Server.Error("handleAnswer: unauthorized", zap.Error(err))
		ctx.WriteErrorAndStatus(err, proto.StatusError)
		return
	}

	var req AnswerReq
	if err := json.Unmarshal(ctx.Body(), &req); err != nil {
		s.Server.Error("handleAnswer unmarshal error", zap.Error(err))
		ctx.WriteErr(err)
		return
	}

	// Load user
	user, err := s.userStore.GetByUsername(context.Background(), username)
	if err != nil {
		s.Server.Error("handleAnswer: user not found", zap.String("username", username), zap.Error(err))
		ctx.WriteErrorAndStatus(errors.New("user not found"), proto.StatusError)
		return
	}

	// Get session metadata
	meta, ok := s.Manager.GetSessionMeta(req.SessionCode)
	if !ok {
		s.Server.Error("handleAnswer: session metadata not found", zap.String("code", req.SessionCode))
		ctx.WriteErrorAndStatus(errors.New("session not found"), proto.StatusError)
		return
	}

	// Optimize: Checklist bằng Redis (O(1)) - Tránh MySQL SELECT hoàn toàn cho việc kiểm tra trùng lặp
	isNew, err := s.quizStore.CheckAndSetAnswered(context.Background(), meta.DBID, user.ID, req.QuestionID)
	if err != nil {
		s.Server.Error("handleAnswer: redis checklist error", zap.Error(err))
		ctx.WriteErr(err)
		return
	}
	if !isNew {
		s.Server.Warn("handleAnswer: user already answered this question", zap.Uint64("userID", user.ID), zap.Uint64("questionID", req.QuestionID))
		ctx.WriteErrorAndStatus(errors.New("user already answered this question"), proto.StatusAlreadyAnswered)
		return
	}

	// Verify answer server-side (Đưa ra ngoài transaction để tối ưu lock time)
	points, isCorrect, err := s.quizStore.ValidateAnswer(context.Background(), meta.QuizID, req.QuestionID, req.AnswerID)
	if err != nil {
		s.Server.Error("handleAnswer: validation error", zap.Error(err))
		ctx.WriteErrorAndStatus(errors.New("validation failed"), proto.StatusError)
		return
	}

	answer := &model.UserAnswer{
		SessionID:  meta.DBID,
		UserID:     user.ID,
		QuestionID: req.QuestionID,
		AnswerID:   req.AnswerID,
		IsCorrect:  isCorrect,
		Score:      points,
	}

	// Wrap ONLY database write operations in a transaction
	err = s.quizStore.Transaction(context.Background(), func(txStore repository.QuizStore) error {
		if err := txStore.SaveUserAnswer(context.Background(), answer); err != nil {
			return err
		}

		// Update score in DB
		if points > 0 {
			if err := txStore.UpdateParticipantScore(context.Background(), meta.DBID, user.ID, points); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		// Rollback Redis checklist if DB transaction failed so user can retry
		_ = s.quizStore.RemoveAnsweredCache(context.Background(), meta.DBID, user.ID, req.QuestionID)

		s.Server.Error("handleAnswer: transaction failed", zap.Error(err))
		ctx.WriteErrorAndStatus(errors.New("failed to save answer"), proto.StatusError)
		return
	}

	// 5. Update in-memory leaderboard (after transaction success)
	if points > 0 {
		if err := s.Manager.SubmitAnswer(req.SessionCode, username, points); err != nil {
			s.Server.Error("handleAnswer: failed to update leaderboard", zap.Error(err))
		} else {
			s.Manager.BroadcastLeaderboard(req.SessionCode)
		}
	}

	s.Server.Metrics().AnswerQuizInc()
	ctx.WriteOk()
}
