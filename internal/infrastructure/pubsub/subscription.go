package pubsub

import (
	"errors"
	"log"
	"sync"
	"sync/atomic"
)

type IDGenerator interface {
	GenerateID() string
}

type Subscription[T any] struct {
	id        string
	manager   *Manager[T]
	ch        chan T
	handler   func(msg T)
	closed    atomic.Bool
	closeOnce sync.Once
}

func newSubscription[T any](id string, manager *Manager[T], handler func(msg T)) *Subscription[T] {
	ch := make(chan T, 100)

	sub := &Subscription[T]{
		id:      id,
		manager: manager,
		ch:      ch,
		handler: handler,
	}

	go func() {
		for msg := range ch {
			if handler != nil {
				handler(msg)
			}
		}
	}()

	return sub
}

func (s *Subscription[T]) ID() string {
	return s.id
}

func (s *Subscription[T]) Unsubscribe() {
	if s.closed.Load() {
		return
	}
	s.manager.Remove(s.id)
}

func (s *Subscription[T]) clear() {
	s.closeOnce.Do(func() {
		s.closed.Store(true)
		close(s.ch)
	})
}

func (s *Subscription[T]) Push(msg T) error {
	if s.closed.Load() {
		return errors.New("subscription is closed")
	}

	select {
	case s.ch <- msg:
		return nil
	default:
		log.Printf("Subscription %s channel full, dropping message", s.id)
		return nil
	}
}
