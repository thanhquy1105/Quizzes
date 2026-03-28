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

	// 1. Get authenticated UID from connection context
	connCtx, ok := ctx.Conn().Context().(*server.ConnContext)
	if !ok || connCtx == nil {
		ctx.WriteErrorAndStatus(errors.New("unauthorized"), proto.StatusError)
		return
	}
	uid := connCtx.UID()

	// 2. Parse request — only session_code needed
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
	user, err := s.userStore.GetByUsername(context.Background(), uid)
	if err != nil {
		s.Server.Error("handleJoin: user not found", zap.String("uid", uid), zap.Error(err))
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
	session.Join(uid, user.Name, ctx.Conn())
	s.Server.Metrics().JoinQuizInc()

	s.Server.Debug("handleJoin: joined", zap.String("uid", uid), zap.String("code", dbSession.SessionCode))

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

// AnswerReq - session_code identifies the session; UID from connection context.
// IsCorrect is NOT trusted from client — backend verifies against DB.
type AnswerReq struct {
	SessionCode string `json:"session_code"`
	QuestionID  uint64 `json:"question_id"`
	AnswerID    uint64 `json:"answer_id"`
}

func (s *QuizServer) handleAnswer(ctx *server.Context) {
	// 1. Get authenticated UID from connection context
	connCtx, ok := ctx.Conn().Context().(*server.ConnContext)
	if !ok || connCtx == nil {
		ctx.WriteErrorAndStatus(errors.New("unauthorized"), proto.StatusError)
		return
	}
	uid := connCtx.UID()

	var req AnswerReq
	if err := json.Unmarshal(ctx.Body(), &req); err != nil {
		s.Server.Error("handleAnswer unmarshal error", zap.Error(err))
		ctx.WriteErr(err)
		return
	}

	// 2. Find user in DB
	user, err := s.userStore.GetByUsername(context.Background(), uid)
	if err != nil {
		s.Server.Error("handleAnswer: user not found", zap.String("uid", uid), zap.Error(err))
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
		ctx.WriteErrorAndStatus(errors.New("user already answered this question"), proto.StatusError)
		return
	}

	quiz, err := s.quizStore.Get(context.Background(), session.QuizID)
	if err != nil {
		s.Server.Error("handleAnswer: quiz not found", zap.Uint64("quizID", session.QuizID), zap.Error(err))
		ctx.WriteErrorAndStatus(errors.New("quiz not found"), proto.StatusError)
		return
	}

	// 4. Verify answer server-side — do NOT trust client's is_correct
	points := 0
	isCorrect := false
	for _, q := range quiz.Questions {
		if q.ID == req.QuestionID {
			for _, a := range q.Answers {
				if a.ID == req.AnswerID && a.IsCorrect {
					isCorrect = true
					points = q.Point
					break
				}
			}
			break
		}
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
		session.SubmitAnswer(uid, points)

		session.BroadcastLeaderboard()
	}

	s.Server.Metrics().AnswerQuizInc()

	ctx.WriteOk()
}
