package quiz

import (
	"encoding/json"
	"sort"
	"sync"
	"time"

	"btaskee-quiz/pkg/wkserver"
	"btaskee-quiz/pkg/wkserver/proto"
	"go.uber.org/zap"
)

type Participant struct {
	UID   string `json:"uid"`
	Name  string `json:"name"`
	Score int    `json:"score"`
	Conn  wkserver.Conn `json:"-"`
}

type Session struct {
	QuizID       string
	mu           sync.RWMutex
	Participants map[string]*Participant
}

func NewSession(quizID string) *Session {
	return &Session{
		QuizID:       quizID,
		Participants: make(map[string]*Participant),
	}
}

func (s *Session) Join(uid, name string, conn wkserver.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Participants[uid] = &Participant{
		UID:   uid,
		Name:  name,
		Score: 0,
		Conn:  conn,
	}
}

func (s *Session) SubmitAnswer(uid string, isCorrect bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if p, ok := s.Participants[uid]; ok {
		if isCorrect {
			p.Score += 10
		}
	}
}

func (s *Session) GetLeaderboard() []*Participant {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var list []*Participant
	for _, p := range s.Participants {
		list = append(list, p)
	}

	sort.Slice(list, func(i, j int) bool {
		return list[i].Score > list[j].Score
	})

	return list
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
		if p.Conn != nil {
			// Using MsgTypeMessage for leaderboard push
			msg := &proto.Message{
				MsgType:   uint32(proto.MsgTypeMessage),
				Timestamp: uint64(time.Now().UnixMilli()),
				Content:   data,
			}
			msgData, err := msg.Encode()
			if err == nil {
				protoEncoding := proto.New()
				payload, err := protoEncoding.Encode(msgData, proto.MsgTypeMessage)
				if err == nil {
					_ = p.Conn.AsyncWrite(payload, nil)
				}
			}
		}
	}
}

type Manager struct {
	mu       sync.RWMutex
	Sessions map[string]*Session
}

func NewManager() *Manager {
	return &Manager{
		Sessions: make(map[string]*Session),
	}
}

func (m *Manager) GetSession(quizID string) *Session {
	m.mu.RLock()
	s, ok := m.Sessions[quizID]
	m.mu.RUnlock()

	if !ok {
		m.mu.Lock()
		s, ok = m.Sessions[quizID]
		if !ok {
			s = NewSession(quizID)
			m.Sessions[quizID] = s
		}
		m.mu.Unlock()
	}
	return s
}

type QuizServer struct {
	WkServer *wkserver.Server
	Manager  *Manager
}

func NewQuizServer(addr string) *QuizServer {
	s := &QuizServer{
		WkServer: wkserver.New(addr, wkserver.WithWSAddr("ws://0.0.0.0:8081")),
		Manager:  NewManager(),
	}
	s.registerRoutes()
	return s
}

func (s *QuizServer) registerRoutes() {
	s.WkServer.Route("/join", s.handleJoin)
	s.WkServer.Route("/answer", s.handleAnswer)
}

type JoinReq struct {
	QuizID string `json:"quiz_id"`
	UID    string `json:"uid"`
	Name   string `json:"name"`
}

func (s *QuizServer) handleJoin(ctx *wkserver.Context) {
	s.WkServer.Debug("handleJoin started", zap.String("body", string(ctx.Body())))
	var req JoinReq
	if err := json.Unmarshal(ctx.Body(), &req); err != nil {
		s.WkServer.Error("handleJoin unmarshal error", zap.Error(err))
		ctx.WriteErr(err)
		return
	}

	session := s.Manager.GetSession(req.QuizID)
	session.Join(req.UID, req.Name, ctx.Conn())

	s.WkServer.Debug("handleJoin: session joined, writing OK", zap.String("uid", req.UID))
	ctx.WriteOk()

	s.WkServer.Debug("handleJoin: broadcasting leaderboard")
	session.BroadcastLeaderboard()
	s.WkServer.Debug("handleJoin finished")
}

type AnswerReq struct {
	QuizID    string `json:"quiz_id"`
	UID       string `json:"uid"`
	IsCorrect bool   `json:"is_correct"`
}

func (s *QuizServer) handleAnswer(ctx *wkserver.Context) {
	var req AnswerReq
	if err := json.Unmarshal(ctx.Body(), &req); err != nil {
		ctx.WriteErr(err)
		return
	}

	session := s.Manager.GetSession(req.QuizID)
	session.SubmitAnswer(req.UID, req.IsCorrect)

	ctx.WriteOk()

	session.BroadcastLeaderboard()
}

func (s *QuizServer) Start() error {
	return s.WkServer.Start()
}
