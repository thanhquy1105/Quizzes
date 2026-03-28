package quiz

import (
	"context"
	"encoding/json"
	"errors"

	"btaskee-quiz/internal/model"
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

	// 3. Find active session by code (time-window validated)
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

	// 4. Fetch quiz (QuizID comes from the session record)
	quiz, err := s.quizStore.Get(context.Background(), dbSession.QuizID)
	if err != nil {
		s.Server.Error("handleJoin: quiz not found", zap.Uint64("quizID", dbSession.QuizID), zap.Error(err))
		ctx.WriteErrorAndStatus(errors.New("quiz not found"), proto.StatusError)
		return
	}

	// 5. Fetch user (Name comes from DB, not from client)
	user, err := s.userStore.GetByUsername(context.Background(), username)
	if err != nil {
		s.Server.Error("handleJoin: user not found", zap.String("username", username), zap.Error(err))
		ctx.WriteErrorAndStatus(errors.New("user not found"), proto.StatusError)
		return
	}

	// 6. Check participation & resume logic
	isPart, err := s.quizStore.IsParticipant(context.Background(), dbSession.ID, user.ID)
	if err != nil {
		s.Server.Error("handleJoin: error checking participation", zap.Error(err))
		ctx.WriteErr(err)
		return
	}

	if isPart {
		// Resume: return ALL questions, mark answered ones so frontend can lock them
		userAnswers, err := s.quizStore.GetUserAnswers(context.Background(), dbSession.ID, user.ID)
		if err != nil {
			s.Server.Error("handleJoin: error fetching user answers", zap.Error(err))
			ctx.WriteErr(err)
			return
		}

		answeredMap := make(map[uint64]bool)
		for _, ans := range userAnswers {
			answeredMap[ans.QuestionID] = true
		}

		for i := range quiz.Questions {
			if answeredMap[quiz.Questions[i].ID] {
				quiz.Questions[i].Answered = true
			}
		}

		s.Server.Info("handleJoin: resuming session", zap.Uint64("userID", user.ID), zap.Int("answered", len(userAnswers)))
	} else {
		// First join: record participant
		err = s.quizStore.AddParticipant(context.Background(), &model.SessionParticipant{
			SessionID: dbSession.ID,
			UserID:    user.ID,
		})
		if err != nil {
			s.Server.Error("handleJoin: error adding participant", zap.Error(err))
			ctx.WriteErr(err)
			return
		}
	}

	// 7. Join in-memory session (keyed by session_code)
	session := s.Manager.GetSession(dbSession.SessionCode, quiz.ID, dbSession.ID)
	session.Join(username, user.Name, user.ID, ctx.Conn())
	s.Server.Metrics().JoinQuizInc()

	s.Server.Debug("handleJoin: joined", zap.String("username", username), zap.String("code", dbSession.SessionCode))

	quizData, err := json.Marshal(quiz)
	if err != nil {
		s.Server.Error("handleJoin: marshal quiz error", zap.Error(err))
		ctx.WriteErr(err)
		return
	}
	ctx.Write(quizData)
	session.BroadcastLeaderboard()
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

	// 2. Load user
	user, err := s.userStore.GetByUsername(context.Background(), username)
	if err != nil {
		s.Server.Error("handleAnswer: user not found", zap.String("username", username), zap.Error(err))
		ctx.WriteErrorAndStatus(errors.New("user not found"), proto.StatusError)
		return
	}

	// 3. Get in-memory session by session_code
	session := s.Manager.GetSessionByCode(req.SessionCode)
	if session == nil {
		s.Server.Error("handleAnswer: session not found", zap.String("code", req.SessionCode))
		ctx.WriteErrorAndStatus(errors.New("session not found"), proto.StatusError)
		return
	}

	// check if user already answered this question
	userAnswer, err := s.quizStore.GetUserAnswer(context.Background(), session.DBID, user.ID, req.QuestionID)
	if err != nil {
		s.Server.Error("handleAnswer: user answer not found", zap.Uint64("userID", user.ID), zap.Uint64("questionID", req.QuestionID), zap.Error(err))
		ctx.WriteErrorAndStatus(errors.New("user answer not found"), proto.StatusError)
		return
	}

	if userAnswer != nil {
		s.Server.Warn("handleAnswer: user already answered this question", zap.Uint64("userID", user.ID), zap.Uint64("questionID", req.QuestionID))
		ctx.WriteErrorAndStatus(errors.New("user already answered this question"), proto.StatusAlreadyAnswered)
		return
	}

	// 4. Verify answer server-side via distributed Redis validation
	points, isCorrect, err := s.quizStore.ValidateAnswer(context.Background(), session.QuizID, req.QuestionID, req.AnswerID)
	if err != nil {
		s.Server.Error("handleAnswer: validation error", zap.Error(err))
		ctx.WriteErrorAndStatus(errors.New("validation failed"), proto.StatusError)
		return
	}

	answer := &model.UserAnswer{
		SessionID:  session.DBID,
		UserID:     user.ID,
		QuestionID: req.QuestionID,
		AnswerID:   req.AnswerID,
		IsCorrect:  isCorrect,
		Score:      points,
	}

	if err := s.quizStore.SaveUserAnswer(context.Background(), answer); err != nil {
		s.Server.Error("handleAnswer: failed to save answer", zap.Error(err))
		ctx.WriteErrorAndStatus(errors.New("failed to save answer"), proto.StatusError)
		return
	}

	// 5. Update score in DB
	if points > 0 {
		if err := s.quizStore.UpdateParticipantScore(context.Background(), session.DBID, user.ID, points); err != nil {
			s.Server.Error("handleAnswer: failed to update score", zap.Error(err))
			ctx.WriteErrorAndStatus(errors.New("failed to update score"), proto.StatusError)
			return
		}

		// 6. Update in-memory leaderboard
		session.SubmitAnswer(username, points)
		session.BroadcastLeaderboard()
	}

	s.Server.Metrics().AnswerQuizInc()

	ctx.WriteOk()
}
