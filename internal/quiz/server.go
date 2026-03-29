package quiz

import (
	"context"
	"errors"
	"time"

	"btaskee-quiz/internal/config"
	"btaskee-quiz/internal/repository"
	"btaskee-quiz/internal/repository/redis"
	"btaskee-quiz/internal/server"
	"btaskee-quiz/internal/server/proto"
	"btaskee-quiz/pkg/token"

	goredis "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type QuizServer struct {
	Server  *server.Server
	Manager *Manager

	quizStore  repository.QuizStore
	lbStore    repository.LeaderboardStore
	userStore  repository.UserStore
	tokenStore repository.TokenStore
	tokenMaker token.IMaker
}

func NewQuizServer(cfg *config.Config, rdb *goredis.Client, tokenStore repository.TokenStore, quizStore repository.QuizStore, userStore repository.UserStore, tokenMaker token.IMaker) *QuizServer {
	lb := redis.NewLeaderboardStore(rdb, quizStore)

	s := &QuizServer{
		Server: server.New(cfg.Server.TCPAddr,
			server.WithWSAddr(cfg.Server.WSAddr),
			server.WithGorillaWSAddr(cfg.Server.GorillaWSAddr),
			server.WithRequestPoolSize(cfg.Server.RequestPoolSize),
			server.WithMessagePoolSize(cfg.Server.MessagePoolSize),
			server.WithMaxIdle(cfg.Server.MaxIdle),
			server.WithRequestTimeout(cfg.Server.RequestTimeout),
			server.WithLogDetail(cfg.Server.LogDetail),
		),
		Manager:    NewManager(lb, userStore),
		quizStore:  redis.NewQuizCache(rdb, quizStore),
		lbStore:    lb,
		userStore:  redis.NewUserCache(rdb, userStore),
		tokenStore: tokenStore,
		tokenMaker: tokenMaker,
	}
	s.registerRoutes()
	return s
}

func (s *QuizServer) registerRoutes() {
	s.Server.Route(s.Server.Options().ConnPath, s.handleConnect)
	s.Server.Route("/join", s.requireAuth(s.handleJoin))
	s.Server.Route("/answer", s.requireAuth(s.handleAnswer))
}

func (s *QuizServer) requireAuth(next server.Handler) server.Handler {
	return func(ctx *server.Context) {
		connCtx, ok := ctx.Conn().Context().(*server.ConnContext)
		if !ok || connCtx == nil || connCtx.Token() == "" {
			s.Server.Debug("unauthorized: missing token in socket session")
			ctx.WriteErrorAndStatus(errors.New("unauthorized"), proto.StatusError)
			_ = ctx.Conn().Close()
			return
		}

		next(ctx)
	}
}

func (s *QuizServer) handleConnect(ctx *server.Context) {
	req := ctx.ConnReq()

	// 1. Verify token cryptographically
	_, err := s.tokenMaker.VerifyToken(req.Token, token.TokenTypeSessionToken)
	if err != nil {
		s.Server.Debug("invalid JWT token", zap.Error(err))
		ctx.WriteConnack(&proto.Connack{
			Id:     req.Id,
			Status: proto.StatusError,
		})
		_ = ctx.Conn().Close()
		return
	}

	// 2. Verify token exists in Redis
	exists, err := s.tokenStore.Exists(context.Background(), req.Token, token.TokenTypeSessionToken)
	if err != nil || !exists {
		s.Server.Debug("token not found in redis", zap.Error(err))
		ctx.WriteConnack(&proto.Connack{
			Id:     req.Id,
			Status: proto.StatusError,
		})
		_ = ctx.Conn().Close()
		return
	}

	// Valid session
	ctx.Conn().SetContext(server.NewConnContext(req.Username, req.Token))
	s.Server.Debug("session context initialized", zap.String("username", req.Username))
	ctx.WriteConnack(&proto.Connack{
		Id:     req.Id,
		Status: proto.StatusOK,
	})
}

func (s *QuizServer) Start() error {
	go s.startReconciliation(context.Background())
	return s.Server.Start()
}

func (s *QuizServer) startReconciliation(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	s.Server.Info("starting periodic leaderboard reconciliation worker")

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.performReconciliation(ctx)
		}
	}
}

func (s *QuizServer) performReconciliation(ctx context.Context) {
	activeSessions, err := s.quizStore.ListActiveSessions(ctx)
	if err != nil {
		s.Server.Error("reconciliation: failed to list active sessions", zap.Error(err))
		return
	}

	for _, session := range activeSessions {
		s.Server.Debug("reconciliation: processing session", zap.String("code", session.SessionCode))

		// 1. Fetch from DB
		entries, err := s.quizStore.GetParticipantsWithScores(ctx, session.SessionCode)
		if err != nil {
			s.Server.Error("reconciliation: failed to fetch participants", zap.String("code", session.SessionCode), zap.Error(err))
			continue
		}

		// 2. Reload into Redis
		if err := s.lbStore.ReloadLeaderboard(ctx, session.SessionCode, entries); err != nil {
			s.Server.Error("reconciliation: failed to reload leaderboard", zap.String("code", session.SessionCode), zap.Error(err))
			continue
		}

		s.Server.Info("reconciliation: success", zap.String("code", session.SessionCode), zap.Int("count", len(entries)))
	}
}
