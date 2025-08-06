package app

import (
	"sync"

	"survival/internal/protocol"
)

type clientSubscription struct {
	id      string
	handler CommandHandler
	manager *subscriptionManager
	once    sync.Once
}

// Unsubscribe terminates the subscription.
func (s *clientSubscription) Unsubscribe() error {
	s.manager.Remove(s.id)
	return nil
}

func (s *clientSubscription) DeliveryChannel(source <-chan protocol.Command) {
	s.once.Do(func() {
		go func() {
			for cmd := range source {
				if s.handler != nil {
					s.handler(cmd) // Call the handler with the command
				}
			}
		}()
	})
}

type subscriptionManager struct {
	mu            sync.RWMutex
	subscriptions map[string]*clientSubscription
	channels      map[string]chan protocol.Command
}

func (sm *subscriptionManager) Add(subscription *clientSubscription) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if subscription == nil {
		return nil
	}

	if _, exists := sm.subscriptions[subscription.id]; exists {
		return ErrClientSubscriptionExists
	}

	sm.subscriptions[subscription.id] = subscription

	subscription.manager = sm

	source := make(chan protocol.Command, 100) // Buffered channel for commands
	subscription.DeliveryChannel(source)
	sm.channels[subscription.id] = source

	return nil
}

func (sm *subscriptionManager) Remove(subscriptionID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if _, exists := sm.subscriptions[subscriptionID]; exists {
		delete(sm.subscriptions, subscriptionID)
	}
	if ch, exists := sm.channels[subscriptionID]; exists {
		delete(sm.channels, subscriptionID)
		close(ch) // Close the channel to stop receiving commands
	}

	return
}

func (sm *subscriptionManager) RemoveAll() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	for id := range sm.subscriptions {
		delete(sm.subscriptions, id)
		if ch, exists := sm.channels[id]; exists {
			close(ch) // Close the channel to stop receiving commands
			delete(sm.channels, id)
		}
	}
}

func (sm *subscriptionManager) All() []*clientSubscription {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	subscriptions := make([]*clientSubscription, 0, len(sm.subscriptions))
	for _, sub := range sm.subscriptions {
		subscriptions = append(subscriptions, sub)
	}
	return subscriptions
}

func (sm *subscriptionManager) AllChannels() []chan<- protocol.Command {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	channels := make([]chan<- protocol.Command, 0, len(sm.channels))
	for _, ch := range sm.channels {
		channels = append(channels, ch)
	}

	return channels
}

func newSubscriptionManager() *subscriptionManager {
	return &subscriptionManager{
		subscriptions: make(map[string]*clientSubscription),
		channels:      make(map[string]chan protocol.Command),
		mu:            sync.RWMutex{},
	}
}
