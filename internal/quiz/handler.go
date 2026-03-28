package quiz

import (
	"encoding/json"

	"btaskee-quiz/internal/server"

	"go.uber.org/zap"
)

type JoinReq struct {
	QuizID string `json:"quiz_id"`
	UID    string `json:"uid"`
	Name   string `json:"name"`
}

func (s *QuizServer) handleJoin(ctx *server.Context) {
	s.Server.Debug("handleJoin started", zap.String("body", string(ctx.Body())))
	var req JoinReq
	if err := json.Unmarshal(ctx.Body(), &req); err != nil {
		s.Server.Error("handleJoin unmarshal error", zap.Error(err))
		ctx.WriteErr(err)
		return
	}

	session := s.Manager.GetSession(req.QuizID)
	session.Join(req.UID, req.Name, ctx.Conn())
	s.Server.Metrics().JoinQuizInc()

	s.Server.Debug("handleJoin: session joined, writing OK", zap.String("uid", req.UID))
	// Using a simple nil response for successful join per previous implementation
	ctx.WriteOk()

	s.Server.Debug("handleJoin: broadcasting leaderboard")
	session.BroadcastLeaderboard()
	s.Server.Debug("handleJoin finished")
}

type AnswerReq struct {
	QuizID    string `json:"quiz_id"`
	UID       string `json:"uid"`
	IsCorrect bool   `json:"is_correct"`
}

func (s *QuizServer) handleAnswer(ctx *server.Context) {
	var req AnswerReq
	if err := json.Unmarshal(ctx.Body(), &req); err != nil {
		ctx.WriteErr(err)
		return
	}

	session := s.Manager.GetSession(req.QuizID)
	session.SubmitAnswer(req.UID, req.IsCorrect)
	s.Server.Metrics().AnswerQuizInc()

	ctx.WriteOk()

	session.BroadcastLeaderboard()
}
