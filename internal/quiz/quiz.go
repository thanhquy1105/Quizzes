package quiz

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"btaskee-quiz/internal/config"
	"btaskee-quiz/pkg/server"
	"btaskee-quiz/pkg/server/proto"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type Participant struct {
	UID   string      `json:"uid"`
	Name  string      `json:"name"`
	Score float64     `json:"score"`
	Conn  server.Conn `json:"-"`
}

type Session struct {
	QuizID       string
	mu           sync.RWMutex
	Participants map[string]*Participant
	lb           LeaderboardStore
}

func NewSession(quizID string, lb LeaderboardStore) *Session {
	return &Session{
		QuizID:       quizID,
		Participants: make(map[string]*Participant),
		lb:           lb,
	}
}

func (s *Session) Join(uid, name string, conn server.Conn) {
	s.mu.Lock()
	s.Participants[uid] = &Participant{
		UID:  uid,
		Name: name,
		Conn: conn,
	}
	s.mu.Unlock()

	_ = s.lb.Add(context.Background(), s.QuizID, uid)
}

func (s *Session) SubmitAnswer(uid string, isCorrect bool) {
	if isCorrect {
		_ = s.lb.IncrBy(context.Background(), s.QuizID, uid, 10)
	}
}

func (s *Session) GetLeaderboard() []*Participant {
	entries, err := s.lb.GetRanked(context.Background(), s.QuizID)
	if err != nil {
		return nil
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*Participant, 0, len(entries))
	for _, e := range entries {
		p, ok := s.Participants[e.UID]
		if !ok {
			continue
		}
		result = append(result, &Participant{
			UID:   p.UID,
			Name:  p.Name,
			Score: e.Score,
		})
	}
	return result
}

func (s *Session) BroadcastLeaderboard() {
	leaderboard := s.GetLeaderboard()
	data, err := json.Marshal(map[string]interface{}{
		"type":        "leaderboard",
		"quiz_id":     s.QuizID,
		"leaderboard": leaderboard,
	})
	if err != nil {
		return
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, p := range s.Participants {
		if p.Conn == nil {
			continue
		}
		msg := &proto.Message{
			MsgType:   uint32(proto.MsgTypeMessage),
			Timestamp: uint64(time.Now().UnixMilli()),
			Content:   data,
		}
		msgData, err := msg.Encode()
		if err != nil {
			continue
		}
		protoEncoding := proto.New()
		payload, err := protoEncoding.Encode(msgData, proto.MsgTypeMessage)
		if err == nil {
			_ = p.Conn.AsyncWrite(payload, nil)
		}
	}
}

type Manager struct {
	mu       sync.RWMutex
	sessions map[string]*Session
	lb       LeaderboardStore
}

func NewManager(lb LeaderboardStore) *Manager {
	return &Manager{
		sessions: make(map[string]*Session),
		lb:       lb,
	}
}

func (m *Manager) GetSession(quizID string) *Session {
	m.mu.RLock()
	s, ok := m.sessions[quizID]
	m.mu.RUnlock()

	if !ok {
		m.mu.Lock()
		s, ok = m.sessions[quizID]
		if !ok {
			s = NewSession(quizID, m.lb)
			m.sessions[quizID] = s
		}
		m.mu.Unlock()
	}
	return s
}

type QuizServer struct {
	Server  *server.Server
	Manager *Manager
}

func NewQuizServer(cfg *config.Config) *QuizServer {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	lb := NewRedisLeaderboard(rdb)
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
		Manager: NewManager(lb),
	}
	s.registerRoutes()
	return s
}

func (s *QuizServer) registerRoutes() {
	s.Server.Route("/join", s.handleJoin)
	s.Server.Route("/answer", s.handleAnswer)
}

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

func (s *QuizServer) Start() error {
	return s.Server.Start()
}
