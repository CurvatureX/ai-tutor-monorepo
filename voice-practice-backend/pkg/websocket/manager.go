package websocket

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"voice-practice-backend/internal/model"
)

// Manager handles WebSocket connections and sessions
type Manager struct {
	connections map[string]*websocket.Conn
	sessions    map[string]*model.VoiceSession
	mutex       sync.RWMutex
	logger      *logrus.Logger
}

// NewManager creates a new WebSocket manager
func NewManager(logger *logrus.Logger) *Manager {
	return &Manager{
		connections: make(map[string]*websocket.Conn),
		sessions:    make(map[string]*model.VoiceSession),
		logger:      logger,
	}
}

// AddConnection adds a new WebSocket connection
func (m *Manager) AddConnection(sessionID string, conn *websocket.Conn) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.connections[sessionID] = conn
	m.sessions[sessionID] = &model.VoiceSession{
		ID:          sessionID,
		AudioBuffer: make([]byte, 0),
		IsRecording: false,
		CreatedAt:   time.Now(),
	}

	m.logger.Infof("Added connection for session: %s", sessionID)
}

// RemoveConnection removes a WebSocket connection
func (m *Manager) RemoveConnection(sessionID string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if conn, exists := m.connections[sessionID]; exists {
		conn.Close()
		delete(m.connections, sessionID)
		delete(m.sessions, sessionID)
		m.logger.Infof("Removed connection for session: %s", sessionID)
	}
}

// GetConnection gets a WebSocket connection by session ID
func (m *Manager) GetConnection(sessionID string) (*websocket.Conn, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	conn, exists := m.connections[sessionID]
	return conn, exists
}

// GetSession gets a voice session by session ID
func (m *Manager) GetSession(sessionID string) (*model.VoiceSession, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	session, exists := m.sessions[sessionID]
	return session, exists
}

// UpdateSession updates a voice session
func (m *Manager) UpdateSession(sessionID string, session *model.VoiceSession) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.sessions[sessionID] = session
}

// SendMessage sends a message to a specific session
func (m *Manager) SendMessage(sessionID string, message *model.WebSocketMessage) error {
	conn, exists := m.GetConnection(sessionID)
	if !exists {
		return fmt.Errorf("connection not found for session: %s", sessionID)
	}

	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %v", err)
	}

	if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
		m.logger.Errorf("Failed to send message to session %s: %v", sessionID, err)
		m.RemoveConnection(sessionID)
		return err
	}

	return nil
}

// SendBinaryMessage sends binary data to a specific session
func (m *Manager) SendBinaryMessage(sessionID string, data []byte) error {
	conn, exists := m.GetConnection(sessionID)
	if !exists {
		return fmt.Errorf("connection not found for session: %s", sessionID)
	}

	if err := conn.WriteMessage(websocket.BinaryMessage, data); err != nil {
		m.logger.Errorf("Failed to send binary message to session %s: %v", sessionID, err)
		m.RemoveConnection(sessionID)
		return err
	}

	return nil
}

// BroadcastMessage broadcasts a message to all connected sessions
func (m *Manager) BroadcastMessage(message *model.WebSocketMessage) {
	m.mutex.RLock()
	sessions := make([]string, 0, len(m.connections))
	for sessionID := range m.connections {
		sessions = append(sessions, sessionID)
	}
	m.mutex.RUnlock()

	for _, sessionID := range sessions {
		if err := m.SendMessage(sessionID, message); err != nil {
			m.logger.Errorf("Failed to broadcast message to session %s: %v", sessionID, err)
		}
	}
}

// GetActiveSessionCount returns the number of active sessions
func (m *Manager) GetActiveSessionCount() int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return len(m.sessions)
}

// CleanupInactiveSessions removes sessions that have been inactive for too long
func (m *Manager) CleanupInactiveSessions(maxIdleTime time.Duration) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	now := time.Now()
	for sessionID, session := range m.sessions {
		if now.Sub(session.CreatedAt) > maxIdleTime {
			if conn, exists := m.connections[sessionID]; exists {
				conn.Close()
				delete(m.connections, sessionID)
			}
			delete(m.sessions, sessionID)
			m.logger.Infof("Cleaned up inactive session: %s", sessionID)
		}
	}
}