package app

import (
	"sync"
)

// ClientRegistry manages the set of active clients.
// It is safe for concurrent use.
type ClientRegistry struct {
	mu      sync.RWMutex
	clients map[string]Client
}

// NewClientRegistry creates a new ClientRegistry.
func NewClientRegistry() *ClientRegistry {
	return &ClientRegistry{
		clients: make(map[string]Client),
	}
}

// Add adds a client to the registry.
func (cr *ClientRegistry) Add(client Client) {
	cr.mu.Lock()
	defer cr.mu.Unlock()
	cr.clients[client.ID()] = client
}

// Remove removes a client from the registry.
func (cr *ClientRegistry) Remove(clientID string) {
	cr.mu.Lock()
	defer cr.mu.Unlock()
	delete(cr.clients, clientID)
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
}
