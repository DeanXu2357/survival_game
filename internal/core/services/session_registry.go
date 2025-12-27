package services

import (
	"sync"

	"survival/internal/core/domain/state"
)

type SessionRegistry struct {
	mu              sync.RWMutex
	sessionToEntity map[string]state.EntityID
	entityToSession map[state.EntityID]string
}

func NewSessionRegistry() *SessionRegistry {
	return &SessionRegistry{
		sessionToEntity: make(map[string]state.EntityID),
		entityToSession: make(map[state.EntityID]string),
	}
}

func (sr *SessionRegistry) Register(sessionID string, entityID state.EntityID) {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	if oldEntityID, exists := sr.sessionToEntity[sessionID]; exists {
		delete(sr.entityToSession, oldEntityID)
	}
	sr.sessionToEntity[sessionID] = entityID
	sr.entityToSession[entityID] = sessionID
}

func (sr *SessionRegistry) Unregister(sessionID string) {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	if entityID, ok := sr.sessionToEntity[sessionID]; ok {
		delete(sr.entityToSession, entityID)
		delete(sr.sessionToEntity, sessionID)
	}
}

func (sr *SessionRegistry) EntityID(sessionID string) (state.EntityID, bool) {
	sr.mu.RLock()
	defer sr.mu.RUnlock()
	entityID, ok := sr.sessionToEntity[sessionID]
	return entityID, ok
}

func (sr *SessionRegistry) SessionID(entityID state.EntityID) (string, bool) {
	sr.mu.RLock()
	defer sr.mu.RUnlock()
	sessionID, ok := sr.entityToSession[entityID]
	return sessionID, ok
}

func (sr *SessionRegistry) AllSessionIDs() []string {
	sr.mu.RLock()
	defer sr.mu.RUnlock()
	ids := make([]string, 0, len(sr.sessionToEntity))
	for id := range sr.sessionToEntity {
		ids = append(ids, id)
	}
	return ids
}

func (sr *SessionRegistry) Clear() {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	sr.sessionToEntity = make(map[string]state.EntityID)
	sr.entityToSession = make(map[state.EntityID]string)
}

func (sr *SessionRegistry) Count() int {
	sr.mu.RLock()
	defer sr.mu.RUnlock()
	return len(sr.sessionToEntity)
}
