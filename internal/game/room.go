package game

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"survival/internal/protocol"
)

type Sender interface {
	Send(ctx context.Context, envelope protocol.ResponseEnvelope) error
	Close(ctx context.Context) error
	ID() string
}

type OutgoingMessage struct {
	TargetPlayerIDs []string // 要發送給哪些玩家，若為 nil 表示廣播
	Data            interface{}
}

type Room struct {
	ID       string
	state    *State
	logic    *Logic
	commands chan protocol.Command
	outgoing chan OutgoingMessage

	clients           map[string]Sender // maps client ID to Sender interface
	clientToPlayerMap map[string]string // maps client ID to Player ID
	playerToClientMap map[string]string // maps Player ID to client ID
	mu                sync.RWMutex
}

func NewRoom(id string) *Room {
	return &Room{
		ID:                id,
		state:             NewGameState(),
		logic:             NewGameLogic(),
		commands:          make(chan protocol.Command, 200),
		outgoing:          make(chan OutgoingMessage, 400),
		mu:                sync.RWMutex{},
		clients:           make(map[string]Sender),
		clientToPlayerMap: make(map[string]string), // maps client ID to Player ID
		playerToClientMap: make(map[string]string), // maps Player ID to client ID
	}
}

func (r *Room) CommandsChannel() chan protocol.Command {
	return r.commands
}

func (r *Room) Run() {
	defer func() {
		if rcv := recover(); rcv != nil {
			log.Printf("Room %s panic: %v", r.ID, rcv)
		}
	}()

	ticker := time.NewTicker(time.Second / targetTickRate)
	defer ticker.Stop()

	currentInputs := make(map[string]protocol.PlayerInput)

	for {
		select {
		case cmd := <-r.commands:
			currentInputs[r.clientToPlayerMap[cmd.ClientID]] = cmd.Input
		case <-ticker.C:
			r.logic.Update(r.state, currentInputs, deltaTime)

			currentInputs = make(map[string]protocol.PlayerInput) // Reset inputs after processing

			// todo: survey how to do incremental updates
			// Broadcast game update to all clients in this room
		}
	}
}

func (r *Room) Shutdown(ctx context.Context) error {
	_, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// save game state if needed
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, client := range r.clients {
		if err := client.Close(ctx); err != nil {
			log.Printf("Error closing client %s: %v", client.ID(), err)
		}
	}
	r.clients = make(map[string]Sender)           // Clear clients
	r.clientToPlayerMap = make(map[string]string) // Clear client to player map
	r.playerToClientMap = make(map[string]string) // Clear player to client map

	return nil
}

func (r *Room) AddClient(_ context.Context, client Sender) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.clients[client.ID()]; exists {
		r.clients[client.ID()] = client // Update existing client
		return nil
	}

	player, err := r.state.newPlayer()
	if err != nil {
		return fmt.Errorf("failed to create new player: %w", err)
	}
	r.clients[client.ID()] = client
	r.clientToPlayerMap[client.ID()] = player.ID
	r.playerToClientMap[player.ID] = client.ID()

	return nil
}

func (r *Room) RemoveClientID(ctx context.Context, clientID string) {
	// todo: how to remove client?
	panic("Implete me")
}
