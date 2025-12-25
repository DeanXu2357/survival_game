package services

import (
	"sync"

	"survival/internal/core/domain"
)

type SessionRegistry struct {
	mu               sync.RWMutex
	sessionToEntity  map[string]domain.EntityID
	entityToSession  map[domain.EntityID]string
}

func NewSessionRegistry() *SessionRegistry {
	return &SessionRegistry{
		sessionToEntity: make(map[string]domain.EntityID),
		entityToSession: make(map[domain.EntityID]string),
	}
}

func (sr *SessionRegistry) Register(sessionID string, entityID domain.EntityID) {
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

func (sr *SessionRegistry) EntityID(sessionID string) (domain.EntityID, bool) {
	sr.mu.RLock()
	defer sr.mu.RUnlock()
	entityID, ok := sr.sessionToEntity[sessionID]
	return entityID, ok
}

func (sr *SessionRegistry) SessionID(entityID domain.EntityID) (string, bool) {
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
	sr.sessionToEntity = make(map[string]domain.EntityID)
	sr.entityToSession = make(map[domain.EntityID]string)
}

func (sr *SessionRegistry) Count() int {
	sr.mu.RLock()
	defer sr.mu.RUnlock()
	return len(sr.sessionToEntity)
}
