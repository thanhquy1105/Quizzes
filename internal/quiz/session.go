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
	Username string      `json:"username"`
	Name     string      `json:"name"`
	UserID   uint64      `json:"user_id"`
	Score    float64     `json:"score"`
	Conn     server.Conn `json:"-"`
}

type Session struct {
	QuizID       uint64
	SessionCode  string
	DBID         uint64
	mu           sync.RWMutex
	Participants map[string]*Participant
	lb           repository.LeaderboardStore
	us           repository.UserStore
}

func NewSession(quizID uint64, sessionCode string, dbID uint64, lb repository.LeaderboardStore, us repository.UserStore) *Session {
	return &Session{
		QuizID:       quizID,
		SessionCode:  sessionCode,
		DBID:         dbID,
		Participants: make(map[string]*Participant),
		lb:           lb,
		us:           us,
	}
}

func (s *Session) Join(username, name string, userID uint64, conn server.Conn) {
	s.mu.Lock()
	s.Participants[username] = &Participant{
		Username: username,
		Name:     name,
		UserID:   userID,
		Conn:     conn,
	}
	s.mu.Unlock()

	_ = s.lb.Add(context.Background(), s.SessionCode, username)
}

func (s *Session) SubmitAnswer(username string, points int) {
	if points > 0 {
		_ = s.lb.IncrBy(context.Background(), s.SessionCode, username, float64(points))
	}
}

// GetParticipant retrieves a joined participant by their Username from memory.
func (s *Session) GetParticipant(username string) (*Participant, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.Participants[username]
	return p, ok
}

func (s *Session) GetLeaderboard() []*Participant {
	entries, err := s.lb.GetRanked(context.Background(), s.SessionCode)
	if err != nil {
		return nil
	}

	result := make([]*Participant, 0, len(entries))
	for _, e := range entries {
		s.mu.RLock()
		p, ok := s.Participants[e.Username]
		s.mu.RUnlock()

		if !ok {
			// Query DB if missing from memory
			user, err := s.us.GetByUsername(context.Background(), e.Username)
			if err != nil {
				continue
			}
			p = &Participant{
				Username: e.Username,
				Name:     user.Name,
				UserID:   user.ID,
				Score:    e.Score,
			}
			// Cache in memory (inactive participant)
			s.mu.Lock()
			s.Participants[e.Username] = p
			s.mu.Unlock()
		}

		result = append(result, &Participant{
			Username: p.Username,
			Name:     p.Name,
			Score:    e.Score,
		})
	}
	return result
}

func (s *Session) BroadcastLeaderboard() {
	leaderboard := s.GetLeaderboard()
	data, err := json.Marshal(map[string]interface{}{
		"type":         "leaderboard",
		"quiz_id":      s.QuizID,
		"session_code": s.SessionCode,
		"leaderboard":  leaderboard,
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
	us       repository.UserStore
}

func NewManager(lb repository.LeaderboardStore, us repository.UserStore) *Manager {
	return &Manager{
		sessions: make(map[string]*Session),
		lb:       lb,
		us:       us,
	}
}

func (m *Manager) GetSession(sessionCode string, quizID uint64, dbID uint64) *Session {
	m.mu.RLock()
	s, ok := m.sessions[sessionCode]
	m.mu.RUnlock()

	if !ok {
		m.mu.Lock()
		s, ok = m.sessions[sessionCode]
		if !ok {
			s = NewSession(quizID, sessionCode, dbID, m.lb, m.us)
			m.sessions[sessionCode] = s
		}
		m.mu.Unlock()
	}
	return s
}

// GetSessionByCode looks up an existing in-memory session without creating one.
// Returns nil if the session has not been initialized yet.
func (m *Manager) GetSessionByCode(sessionCode string) *Session {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.sessions[sessionCode]
}
