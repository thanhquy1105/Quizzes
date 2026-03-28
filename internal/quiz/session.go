package quiz

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"btaskee-quiz/internal/repository"
	"btaskee-quiz/internal/server"
	"btaskee-quiz/internal/server/proto"
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
	lb           repository.LeaderboardStore
}

func NewSession(quizID string, lb repository.LeaderboardStore) *Session {
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
	lb       repository.LeaderboardStore
}

func NewManager(lb repository.LeaderboardStore) *Manager {
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
