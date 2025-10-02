package game

import (
	"sync"
)

// PlayerRegistry manages the mapping between session IDs and player IDs within a room.
// It is safe for concurrent use.
type PlayerRegistry struct {
	mu                 sync.RWMutex
	sessionToPlayerMap map[string]string // maps session ID to Player ID
	playerToSessionMap map[string]string // maps Player ID to session ID
}

// NewPlayerRegistry creates a new PlayerRegistry.
func NewPlayerRegistry() *PlayerRegistry {
	return &PlayerRegistry{
		sessionToPlayerMap: make(map[string]string),
		playerToSessionMap: make(map[string]string),
	}
}

// Register maps a session ID to a player ID.
func (pr *PlayerRegistry) Register(sessionID, playerID string) {
	pr.mu.Lock()
	defer pr.mu.Unlock()
	pr.sessionToPlayerMap[sessionID] = playerID
	pr.playerToSessionMap[playerID] = sessionID
}

// Unregister removes a session and its associated player from the registry.
func (pr *PlayerRegistry) Unregister(sessionID string) {
	pr.mu.Lock()
	defer pr.mu.Unlock()
	if playerID, ok := pr.sessionToPlayerMap[sessionID]; ok {
		delete(pr.playerToSessionMap, playerID)
		delete(pr.sessionToPlayerMap, sessionID)
	}
}

// PlayerID retrieves the player ID for a given session ID.
func (pr *PlayerRegistry) PlayerID(sessionID string) (string, bool) {
	pr.mu.RLock()
	defer pr.mu.RUnlock()
	playerID, ok := pr.sessionToPlayerMap[sessionID]
	return playerID, ok
}

// SessionID retrieves the session ID for a given player ID.
func (pr *PlayerRegistry) SessionID(playerID string) (string, bool) {
	pr.mu.RLock()
	defer pr.mu.RUnlock()
	sessionID, ok := pr.playerToSessionMap[playerID]
	return sessionID, ok
}

// AllSessionIDs returns a snapshot of all session IDs in the room.
func (pr *PlayerRegistry) AllSessionIDs() []string {
	pr.mu.RLock()
	defer pr.mu.RUnlock()
	ids := make([]string, 0, len(pr.sessionToPlayerMap))
	for id := range pr.sessionToPlayerMap {
		ids = append(ids, id)
	}
	return ids
}

// Clear removes all entries from the registry.
func (pr *PlayerRegistry) Clear() {
	pr.mu.Lock()
	defer pr.mu.Unlock()
	pr.sessionToPlayerMap = make(map[string]string)
	pr.playerToSessionMap = make(map[string]string)
}
