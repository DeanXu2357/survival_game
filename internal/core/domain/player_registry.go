package domain

import (
	"sync"
)

// PlayerRegistry manages the mapping between session IDs and entity IDs within a room.
// It is safe for concurrent use.
type PlayerRegistry struct {
	mu                 sync.RWMutex
	sessionToEntityMap map[string]EntityID // maps session ID to EntityID
	entityToSessionMap map[EntityID]string // maps EntityID to session ID
}

// NewPlayerRegistry creates a new PlayerRegistry.
func NewPlayerRegistry() *PlayerRegistry {
	return &PlayerRegistry{
		sessionToEntityMap: make(map[string]EntityID),
		entityToSessionMap: make(map[EntityID]string),
	}
}

// Register maps a session ID to an entity ID.
func (pr *PlayerRegistry) Register(sessionID string, entityID EntityID) {
	pr.mu.Lock()
	defer pr.mu.Unlock()
	pr.sessionToEntityMap[sessionID] = entityID
	pr.entityToSessionMap[entityID] = sessionID
}

// Unregister removes a session and its associated entity from the registry.
func (pr *PlayerRegistry) Unregister(sessionID string) {
	pr.mu.Lock()
	defer pr.mu.Unlock()
	if entityID, ok := pr.sessionToEntityMap[sessionID]; ok {
		delete(pr.entityToSessionMap, entityID)
		delete(pr.sessionToEntityMap, sessionID)
	}
}

// EntityID retrieves the entity ID for a given session ID.
func (pr *PlayerRegistry) EntityID(sessionID string) (EntityID, bool) {
	pr.mu.RLock()
	defer pr.mu.RUnlock()
	entityID, ok := pr.sessionToEntityMap[sessionID]
	return entityID, ok
}

// SessionID retrieves the session ID for a given entity ID.
func (pr *PlayerRegistry) SessionID(entityID EntityID) (string, bool) {
	pr.mu.RLock()
	defer pr.mu.RUnlock()
	sessionID, ok := pr.entityToSessionMap[entityID]
	return sessionID, ok
}

// AllSessionIDs returns a snapshot of all session IDs in the room.
func (pr *PlayerRegistry) AllSessionIDs() []string {
	pr.mu.RLock()
	defer pr.mu.RUnlock()
	ids := make([]string, 0, len(pr.sessionToEntityMap))
	for id := range pr.sessionToEntityMap {
		ids = append(ids, id)
	}
	return ids
}

// AllEntityIDs returns a snapshot of all entity IDs in the room.
func (pr *PlayerRegistry) AllEntityIDs() []EntityID {
	pr.mu.RLock()
	defer pr.mu.RUnlock()
	ids := make([]EntityID, 0, len(pr.entityToSessionMap))
	for id := range pr.entityToSessionMap {
		ids = append(ids, id)
	}
	return ids
}

// Clear removes all entries from the registry.
func (pr *PlayerRegistry) Clear() {
	pr.mu.Lock()
	defer pr.mu.Unlock()
	pr.sessionToEntityMap = make(map[string]EntityID)
	pr.entityToSessionMap = make(map[EntityID]string)
}
