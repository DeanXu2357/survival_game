package services

import (
	"sync"
)

type Manager[T any] struct {
	mu    sync.RWMutex
	subs  map[string]*Subscription[T]
	idGen IDGenerator
}

func NewManager[T any](generator IDGenerator) *Manager[T] {
	return &Manager[T]{
		subs:  make(map[string]*Subscription[T]),
		idGen: generator,
	}
}

func (m *Manager[T]) Remove(subID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if sub, ok := m.subs[subID]; ok {
		delete(m.subs, subID)
		sub.clear()
	}
}

func (m *Manager[T]) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, sub := range m.subs {
		sub.clear()
	}

	m.subs = make(map[string]*Subscription[T])
}

func (m *Manager[T]) All() []*Subscription[T] {
	m.mu.RLock()
	defer m.mu.RUnlock()

	subs := make([]*Subscription[T], 0, len(m.subs))
	for _, sub := range m.subs {
		subs = append(subs, sub)
	}
	return subs
}

func (m *Manager[T]) Add(handler func(msg T)) (*Subscription[T], error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	sub := newSubscription(m.idGen.GenerateID(), m, handler)
	m.subs[sub.ID()] = sub
	return sub, nil
}
