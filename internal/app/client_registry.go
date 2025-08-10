package app

import (
	"fmt"
	"sync"
	"time"
)

var ErrClientSessionValidationFailed = fmt.Errorf("client session validation failed")

// SessionInfo contains information about a client session
type SessionInfo struct {
	SessionID string
	ClientID  string
	GameName  string
	Client    Client
	LastSeen  time.Time
}

// ClientRegistry manages the set of active clients.
// It is safe for concurrent use.
type ClientRegistry struct {
	mu             sync.RWMutex
	clients        map[string]Client       // clientID -> Client
	sessions       map[string]*SessionInfo // sessionID -> SessionInfo
	clientSessions map[string]string       // clientID -> sessionID
	idGen          IDGenerator
}

// NewClientRegistry creates a new ClientRegistry.
func NewClientRegistry(idGen IDGenerator) *ClientRegistry {
	return &ClientRegistry{
		clients:        make(map[string]Client),
		sessions:       make(map[string]*SessionInfo),
		clientSessions: make(map[string]string),
		idGen:          idGen,
	}
}

type AddEvent func(client Client, sessionID, gameName string)

// Add adds a client to the registry with full reconnection logic.
// onNewConnection is called for new connections only.
// onReconnection is called for successful reconnections only.
func (cr *ClientRegistry) Add(client Client, providedSessionID, gameName string, onNewConnection AddEvent, onReconnection AddEvent) error {
	if client.IsClosed() {
		return fmt.Errorf("cannot add closed client %s", client.ID())
	}

	queryClient, exist := cr.Get(client.ID())
	if exist {
		if queryClient.SessionID() != providedSessionID {
			return fmt.Errorf("client %s already exists with a different session ID", client.ID())
		}

		// reconnection logic
		cr.Remove(queryClient.ID())
		queryClient.Close()
	}

	cr.mu.Lock()
	defer cr.mu.Unlock()

	var sessionInfo *SessionInfo
	var isReconnection bool
	if providedSessionID != "" {
		if info, exists := cr.sessions[providedSessionID]; exists &&
			info.ClientID == client.ID() && info.GameName == gameName {
			info.Client.Close()
			info.Client = client

			sessionInfo = info
			isReconnection = true
		} else {
			return fmt.Errorf("%w: sessionID=%s, clientID=%s, gameName=%ss", ErrClientSessionValidationFailed, providedSessionID, client.ID(), gameName)
		}
	}

	if sessionInfo == nil {
		sessionInfo = &SessionInfo{
			SessionID: cr.idGen.GenerateID(),
			ClientID:  client.ID(),
			GameName:  gameName,
			Client:    client,
			LastSeen:  time.Now(),
		}
	}

	if err := client.SetSessionID(sessionInfo.SessionID); err != nil {
		return fmt.Errorf("failed to set session ID for client %s: %w", client.ID(), err)
	}

	cr.clients[client.ID()] = client
	cr.sessions[sessionInfo.SessionID] = sessionInfo
	cr.clientSessions[client.ID()] = sessionInfo.SessionID

	if !isReconnection && onNewConnection != nil {
		onNewConnection(client, sessionInfo.SessionID, gameName)
	}

	if isReconnection && onReconnection != nil {
		onReconnection(client, sessionInfo.SessionID, gameName)
	}

	return nil
}

// Remove removes a client from the registry but preserves session info for potential reconnection.
func (cr *ClientRegistry) Remove(clientID string) {
	cr.mu.Lock()
	defer cr.mu.Unlock()
	delete(cr.clients, clientID)
	// Note: We intentionally keep session info for reconnection
	// Sessions will be cleaned up by CleanupExpiredSessions
}

// Get retrieves a client by its ID.
func (cr *ClientRegistry) Get(clientID string) (Client, bool) {
	cr.mu.RLock()
	defer cr.mu.RUnlock()
	client, ok := cr.clients[clientID]
	return client, ok
}

// All returns a snapshot of all clients.
func (cr *ClientRegistry) All() []Client {
	cr.mu.RLock()
	defer cr.mu.RUnlock()
	list := make([]Client, 0, len(cr.clients))
	for _, client := range cr.clients {
		list = append(list, client)
	}
	return list
}

// Clear removes all clients from the registry.
func (cr *ClientRegistry) Clear() {
	cr.mu.Lock()
	defer cr.mu.Unlock()
	cr.clients = make(map[string]Client)
	cr.sessions = make(map[string]*SessionInfo)
	cr.clientSessions = make(map[string]string)
}
