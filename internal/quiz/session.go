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

type SessionMeta struct {
	QuizID      uint64
	SessionCode string
	DBID        uint64
}

type Manager struct {
	mu           sync.RWMutex
	participants map[string]map[string]*Participant // sessionCode -> username -> Participant
	metas        map[string]*SessionMeta           // sessionCode -> SessionMeta
	lb           repository.LeaderboardStore
	us           repository.UserStore
}

func NewManager(lb repository.LeaderboardStore, us repository.UserStore) *Manager {
	return &Manager{
		participants: make(map[string]map[string]*Participant),
		metas:        make(map[string]*SessionMeta),
		lb:           lb,
		us:           us,
	}
}

func (m *Manager) Join(sessionCode string, quizID, dbID uint64, username, name string, userID uint64, score float64, conn server.Conn) {
	m.mu.Lock()
	if _, ok := m.participants[sessionCode]; !ok {
		m.participants[sessionCode] = make(map[string]*Participant)
	}
	if _, ok := m.metas[sessionCode]; !ok {
		m.metas[sessionCode] = &SessionMeta{
			QuizID:      quizID,
			SessionCode: sessionCode,
			DBID:        dbID,
		}
	}
	m.participants[sessionCode][username] = &Participant{
		Username: username,
		Name:     name,
		UserID:   userID,
		Score:    score,
		Conn:     conn,
	}
	m.mu.Unlock()

	_ = m.lb.Add(context.Background(), sessionCode, username, score)
}

func (m *Manager) SubmitAnswer(sessionCode string, username string, points int) error {
	if points > 0 {
		return m.lb.IncrBy(context.Background(), sessionCode, username, float64(points))
	}
	return nil
}

func (m *Manager) GetParticipant(sessionCode, username string) (*Participant, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if ps, ok := m.participants[sessionCode]; ok {
		p, ok := ps[username]
		return p, ok
	}
	return nil, false
}

func (m *Manager) GetSessionMeta(sessionCode string) (*SessionMeta, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	meta, ok := m.metas[sessionCode]
	return meta, ok
}

func (m *Manager) GetLeaderboard(sessionCode string) []*Participant {
	entries, err := m.lb.GetRanked(context.Background(), sessionCode)
	if err != nil {
		return nil
	}

	result := make([]*Participant, 0, len(entries))
	for _, e := range entries {
		m.mu.RLock()
		ps, ok := m.participants[sessionCode]
		var p *Participant
		if ok {
			p, ok = ps[e.Username]
		}
		m.mu.RUnlock()

		if !ok {
			// Query DB if missing from memory
			user, err := m.us.GetByUsername(context.Background(), e.Username)
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
			m.mu.Lock()
			if _, ok := m.participants[sessionCode]; !ok {
				m.participants[sessionCode] = make(map[string]*Participant)
			}
			m.participants[sessionCode][e.Username] = p
			m.mu.Unlock()
		}

		result = append(result, &Participant{
			Username: p.Username,
			Name:     p.Name,
			Score:    e.Score,
		})
	}
	return result
}

func (m *Manager) BroadcastLeaderboard(sessionCode string) {
	leaderboard := m.GetLeaderboard(sessionCode)
	meta, ok := m.GetSessionMeta(sessionCode)
	if !ok {
		return
	}

	data, err := json.Marshal(map[string]interface{}{
		"type":         "leaderboard",
		"quiz_id":      meta.QuizID,
		"session_code": sessionCode,
		"leaderboard":  leaderboard,
	})
	if err != nil {
		return
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	ps, ok := m.participants[sessionCode]
	if !ok {
		return
	}

	for _, p := range ps {
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
