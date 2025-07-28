package game

import (
	"context"
	"encoding/json"
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

	ctx    context.Context
	cancel context.CancelFunc

	clients           map[string]Sender // maps client ID to Sender interface
	clientToPlayerMap map[string]string // maps client ID to Player ID
	playerToClientMap map[string]string // maps Player ID to client ID
	mu                sync.RWMutex
}

func NewRoom(ctx context.Context, id string) *Room {
	roomCTX, cancel := context.WithCancel(ctx)

	return &Room{
		ID:       id,
		state:    NewGameState(),
		logic:    NewGameLogic(),
		commands: make(chan protocol.Command, 200),
		outgoing: make(chan OutgoingMessage, 400),

		mu:                sync.RWMutex{},
		clients:           make(map[string]Sender),
		clientToPlayerMap: make(map[string]string), // maps client ID to Player ID
		playerToClientMap: make(map[string]string), // maps Player ID to client ID

		ctx:    roomCTX,
		cancel: cancel,
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
			r.mu.RLock() // TODO: use private method package logics of reading clientToPlayerMap
			defer r.mu.RUnlock()
			playerID := r.clientToPlayerMap[cmd.ClientID]
			if playerID == "" {
				log.Printf("Warning: No player ID found for client %s", cmd.ClientID)
				continue
			}
			currentInputs[playerID] = cmd.Input
			log.Printf("Received input from client %s (player %s): %+v", cmd.ClientID, playerID, cmd.Input)
		case <-ticker.C:
			r.logic.Update(r.state, currentInputs, deltaTime)

			currentInputs = make(map[string]protocol.PlayerInput) // Reset inputs after processing

			// todo: survey how to do incremental updates
			// Broadcast game update to all clients in this room
			r.broadcastGameUpdate()
		case <-r.ctx.Done():
			return
		}
	}
}

func (r *Room) Shutdown(ctx context.Context) error {
	defer r.cancel()

	_, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	defer func() {
		log.Printf("Room %s shutdown completed", r.ID)
	}()

	// save game state if needed
	r.mu.Lock()
	defer r.mu.Unlock()

	log.Println("clients", r.clients)

	for _, client := range r.clients {
		log.Printf("Closing client %s in room %s", client.ID(), r.ID)
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

	log.Printf("Player created successfully - Client: %s, Player: %s, Position: %+v",
		client.ID(), player.ID, player.Position)
	log.Printf("Total players in room: %d", len(r.state.Players))

	return nil
}

func (r *Room) broadcastGameUpdate() {
	// Create game state update
	gameUpdate := r.state.ToClientState()

	// Marshal to JSON
	payloadBytes, err := json.Marshal(gameUpdate)
	if err != nil {
		log.Printf("Failed to marshal game update: %v", err)
		return
	}

	envelope := protocol.ResponseEnvelope{
		Type:    protocol.GameUpdateEnvelope,
		Payload: json.RawMessage(payloadBytes),
	}

	r.mu.RLock()

	// Send to all clients (create a snapshot to avoid modification during iteration)
	clientSnapshot := make(map[string]Sender)
	for k, v := range r.clients {
		clientSnapshot[k] = v
	}

	r.mu.RUnlock()

	for clientID, client := range clientSnapshot {
		go func(cID string, c Sender) {
			defer func() {
				if panicVal := recover(); panicVal != nil {
					log.Printf("Recovered from panic while sending to client %s: %v", cID, panicVal)
					// TODO: Remove failed client from room
				}
			}()

			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()

			if err := c.Send(ctx, envelope); err != nil {
				log.Printf("Failed to send game update to client %s: %v", c.ID(), err)
				// TODO: Consider removing disconnected clients from the room
			}
		}(clientID, client)
	}
}
