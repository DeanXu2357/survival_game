package game

import "sync"

// PlayerRegistry manages the mapping between client IDs and player IDs within a room.
// It is safe for concurrent use.
type PlayerRegistry struct {
	mu                sync.RWMutex
	clientToPlayerMap map[string]string // maps client ID to Player ID
	playerToClientMap map[string]string // maps Player ID to client ID
}

// NewPlayerRegistry creates a new PlayerRegistry.
func NewPlayerRegistry() *PlayerRegistry {
	return &PlayerRegistry{
		clientToPlayerMap: make(map[string]string),
		playerToClientMap: make(map[string]string),
	}
}

// Register maps a client ID to a player ID.
func (pr *PlayerRegistry) Register(clientID, playerID string) {
	pr.mu.Lock()
	defer pr.mu.Unlock()
	pr.clientToPlayerMap[clientID] = playerID
	pr.playerToClientMap[playerID] = clientID
}

// Unregister removes a client and its associated player from the registry.
func (pr *PlayerRegistry) Unregister(clientID string) {
	pr.mu.Lock()
	defer pr.mu.Unlock()
	if playerID, ok := pr.clientToPlayerMap[clientID]; ok {
		delete(pr.playerToClientMap, playerID)
		delete(pr.clientToPlayerMap, clientID)
	}
}

// GetPlayerID retrieves the player ID for a given client ID.
func (pr *PlayerRegistry) GetPlayerID(clientID string) (string, bool) {
	pr.mu.RLock()
	defer pr.mu.RUnlock()
	playerID, ok := pr.clientToPlayerMap[clientID]
	return playerID, ok
}

// GetClientID retrieves the client ID for a given player ID.
func (pr *PlayerRegistry) GetClientID(playerID string) (string, bool) {
	pr.mu.RLock()
	defer pr.mu.RUnlock()
	clientID, ok := pr.playerToClientMap[playerID]
	return clientID, ok
}

// AllClientIDs returns a snapshot of all client IDs in the room.
func (pr *PlayerRegistry) AllClientIDs() []string {
	pr.mu.RLock()
	defer pr.mu.RUnlock()
	ids := make([]string, 0, len(pr.clientToPlayerMap))
	for id := range pr.clientToPlayerMap {
		ids = append(ids, id)
	}
	return ids
}

// Clear removes all entries from the registry.
func (pr *PlayerRegistry) Clear() {
	pr.mu.Lock()
	defer pr.mu.Unlock()
	pr.clientToPlayerMap = make(map[string]string)
	pr.playerToClientMap = make(map[string]string)
}
