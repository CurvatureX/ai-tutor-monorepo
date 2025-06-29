package manager

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"

	"github.com/ai-tutor-monorepo/gateway/internal/model"
)

// WebSocketManager manages WebSocket connections and sessions
type WebSocketManager struct {
	sessions map[string]*model.WebSocketSession
	mutex    sync.RWMutex
	logger   *logrus.Logger
	done     chan struct{}
}

// NewWebSocketManager creates a new WebSocket manager
func NewWebSocketManager(logger *logrus.Logger) *WebSocketManager {
	return &WebSocketManager{
		sessions: make(map[string]*model.WebSocketSession),
		logger:   logger,
		done:     make(chan struct{}),
	}
}

// AddConnection adds a new WebSocket connection
func (m *WebSocketManager) AddConnection(sessionID string, conn *websocket.Conn) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	now := time.Now()
	session := &model.WebSocketSession{
		ID:           sessionID,
		Connection:   conn,
		IsRecording:  false,
		StartTime:    now,
		LastActivity: now,
		Metadata:     make(map[string]interface{}),
	}

	m.sessions[sessionID] = session
	m.logger.Infof("Added WebSocket connection for session: %s", sessionID)
}

// RemoveConnection removes a WebSocket connection
func (m *WebSocketManager) RemoveConnection(sessionID string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if session, exists := m.sessions[sessionID]; exists {
		session.Connection.Close()
		delete(m.sessions, sessionID)
		m.logger.Infof("Removed WebSocket connection for session: %s", sessionID)
	}
}

// GetSession returns a session by ID
func (m *WebSocketManager) GetSession(sessionID string) (*model.WebSocketSession, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	session, exists := m.sessions[sessionID]
	if exists {
		session.LastActivity = time.Now()
	}
	return session, exists
}

// UpdateSession updates session state
func (m *WebSocketManager) UpdateSession(sessionID string, session *model.WebSocketSession) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	session.LastActivity = time.Now()
	m.sessions[sessionID] = session
}

// SendMessage sends a JSON message to a session
func (m *WebSocketManager) SendMessage(sessionID string, message *model.WebSocketMessage) error {
	session, exists := m.GetSession(sessionID)
	if !exists {
		return NewSessionNotFoundError(sessionID)
	}

	message.Timestamp = time.Now().UnixMilli()
	data, err := json.Marshal(message)
	if err != nil {
		m.logger.Errorf("Failed to marshal message for session %s: %v", sessionID, err)
		return err
	}

	if err := session.Connection.WriteMessage(websocket.TextMessage, data); err != nil {
		m.logger.Errorf("Failed to send text message to session %s: %v", sessionID, err)
		m.RemoveConnection(sessionID)
		return err
	}

	return nil
}

// SendBinaryMessage sends binary data to a session
func (m *WebSocketManager) SendBinaryMessage(sessionID string, data []byte) error {
	session, exists := m.GetSession(sessionID)
	if !exists {
		return NewSessionNotFoundError(sessionID)
	}

	if err := session.Connection.WriteMessage(websocket.BinaryMessage, data); err != nil {
		m.logger.Errorf("Failed to send binary message to session %s: %v", sessionID, err)
		m.RemoveConnection(sessionID)
		return err
	}

	return nil
}

// BroadcastMessage sends a message to all active sessions
func (m *WebSocketManager) BroadcastMessage(message *model.WebSocketMessage) {
	m.mutex.RLock()
	sessionIDs := make([]string, 0, len(m.sessions))
	for sessionID := range m.sessions {
		sessionIDs = append(sessionIDs, sessionID)
	}
	m.mutex.RUnlock()

	for _, sessionID := range sessionIDs {
		if err := m.SendMessage(sessionID, message); err != nil {
			m.logger.Errorf("Failed to broadcast message to session %s: %v", sessionID, err)
		}
	}
}

// GetActiveSessions returns the number of active sessions
func (m *WebSocketManager) GetActiveSessions() int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return len(m.sessions)
}

// StartCleanupRoutine starts the session cleanup routine
func (m *WebSocketManager) StartCleanupRoutine(interval, maxInactivity time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.CleanupInactiveSessions(maxInactivity)
		case <-m.done:
			return
		}
	}
}

// CleanupInactiveSessions removes inactive sessions
func (m *WebSocketManager) CleanupInactiveSessions(maxInactivity time.Duration) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	now := time.Now()
	toRemove := make([]string, 0)

	for sessionID, session := range m.sessions {
		if now.Sub(session.LastActivity) > maxInactivity {
			toRemove = append(toRemove, sessionID)
		}
	}

	for _, sessionID := range toRemove {
		if session, exists := m.sessions[sessionID]; exists {
			session.Connection.Close()
			delete(m.sessions, sessionID)
			m.logger.Infof("Cleaned up inactive session: %s", sessionID)
		}
	}

	if len(toRemove) > 0 {
		m.logger.Infof("Cleaned up %d inactive sessions", len(toRemove))
	}
}

// Shutdown gracefully shuts down the manager
func (m *WebSocketManager) Shutdown() {
	close(m.done)

	m.mutex.Lock()
	defer m.mutex.Unlock()

	for sessionID, session := range m.sessions {
		session.Connection.Close()
		m.logger.Infof("Closed connection for session: %s", sessionID)
	}
	
	m.sessions = make(map[string]*model.WebSocketSession)
	m.logger.Info("WebSocket manager shutdown complete")
}

// Custom errors
type SessionNotFoundError struct {
	SessionID string
}

func (e *SessionNotFoundError) Error() string {
	return "session not found: " + e.SessionID
}

func NewSessionNotFoundError(sessionID string) *SessionNotFoundError {
	return &SessionNotFoundError{SessionID: sessionID}
}