package quiz

import (
	"context"
	"errors"

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

	tokenStore repository.TokenStore
	tokenMaker token.IMaker
}

func NewQuizServer(cfg *config.Config, rdb *goredis.Client, tokenStore repository.TokenStore, tokenMaker token.IMaker) *QuizServer {
	lb := redis.NewLeaderboardStore(rdb)

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
		Manager:    NewManager(lb),
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
	ctx.Conn().SetContext(server.NewConnContext(req.Uid, req.Token))
	s.Server.Debug("session context initialized", zap.String("uid", req.Uid))
	ctx.WriteConnack(&proto.Connack{
		Id:     req.Id,
		Status: proto.StatusOK,
	})
}

func (s *QuizServer) Start() error {
	return s.Server.Start()
}
