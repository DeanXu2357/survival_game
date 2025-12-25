package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"survival/internal/core/domain"
	"survival/internal/core/ports"
	"survival/internal/utils"
)

type UpdateMessage struct {
	ToSessions []string // an empty slice means broadcast to all
	Envelope   ports.ResponseEnvelope
}

type Room struct {
	ID         string
	mapConfig  *domain.MapConfig
	game       *domain.Game
	sessions   *SessionRegistry
	subManager *Manager[UpdateMessage]

	commands chan ports.Command
	outgoing chan UpdateMessage

	ctx    context.Context
	cancel context.CancelFunc
}

func NewRoom(ctx context.Context, id string) *Room {
	roomCTX, cancel := context.WithCancel(ctx)

	return &Room{
		ID:         id,
		mapConfig:  nil,
		game:       domain.NewGame(),
		sessions:   NewSessionRegistry(),
		subManager: NewManager[UpdateMessage](utils.NewSequentialIDGenerator(fmt.Sprintf("room%s-sub-", id))),

		commands: make(chan ports.Command, 200),
		outgoing: make(chan UpdateMessage, 400),

		ctx:    roomCTX,
		cancel: cancel,
	}
}

func NewRoomWithMap(ctx context.Context, id string, mapConfig *domain.MapConfig) *Room {
	roomCTX, cancel := context.WithCancel(ctx)

	return &Room{
		ID:         id,
		mapConfig:  mapConfig,
		game:       domain.NewGame(), // TODO: initialize game with mapConfig
		sessions:   NewSessionRegistry(),
		subManager: NewManager[UpdateMessage](utils.NewSequentialIDGenerator(fmt.Sprintf("room%s-sub-", id))),

		commands: make(chan ports.Command, 200),
		outgoing: make(chan UpdateMessage, 400),

		ctx:    roomCTX,
		cancel: cancel,
	}
}

func (r *Room) SendCommand(cmd ports.Command) {
	select {
	case r.commands <- cmd:
	default:
		log.Printf("Room %s command channel full, dropping command from client %s", r.ID, cmd.SessionID)
	}
}

func (r *Room) SubscribeResponse(handler func(msg UpdateMessage)) error {
	// todo: check room is running

	if handler == nil {
		return fmt.Errorf("handler cannot be nil")
	}

	_, err := r.subManager.Add(handler)
	if err != nil {
		return fmt.Errorf("failed to add subscription to room %s: %w", r.ID, err)
	}

	return nil
}

// OutgoingChannel returns a read-only channel for outgoing messages.
func (r *Room) OutgoingChannel() <-chan UpdateMessage {
	return r.outgoing
}

func (r *Room) responsePump() {
	for msg := range r.outgoing {

		for _, sub := range r.subManager.All() {
			if err := sub.Push(msg); err != nil {
				log.Printf("Failed to push message to subscription %s: %v", sub.ID(), err)
			}
		}

		select {
		case <-r.ctx.Done():
			return
		default:
		}
	}
}

func (r *Room) Run() {
	defer func() {
		if rcv := recover(); rcv != nil {
			log.Printf("Room %s panic: %v", r.ID, rcv)
		}
		log.Printf("Room %s stopped", r.ID)
	}()

	// TODO: should extract  game run and member management to separate methods.

	go r.responsePump()

	ticker := time.NewTicker(time.Second / ports.TargetTickRate)
	defer ticker.Stop()

	currentInputs := make(map[domain.EntityID]ports.PlayerInput)

	for {
		select {
		case cmd := <-r.commands:
			entityID, ok := r.sessions.EntityID(cmd.SessionID)
			if !ok {
				log.Printf("Warning: No entity ID found for session %s", cmd.SessionID)
				continue
			}
			currentInputs[entityID] = cmd.Input
		case <-ticker.C:
			r.game.UpdateInLoop(ports.DeltaTime, currentInputs)
			currentInputs = make(map[domain.EntityID]ports.PlayerInput) // Reset inputs after processing
			r.broadcastGameUpdate()
		case <-r.ctx.Done():
			return
		}
	}
}

func (r *Room) Shutdown(ctx context.Context) error {
	defer r.cancel()

	log.Printf("Room %s shutdown initiated", r.ID)

	r.sessions.Clear()
	r.subManager.Clear()

	log.Printf("Room %s shutdown completed", r.ID)
	return nil
}

// AddPlayer creates a new player or reconnects existing players and registers them to the room.
func (r *Room) AddPlayer(client Client) error {
	sessionID := client.SessionID()
	if _, exist := r.sessions.EntityID(sessionID); exist {
		return nil
	}

	entityID, err := r.game.JoinPlayer()
	if err != nil {
		return fmt.Errorf("failed to create new player: %w", err)
	}

	r.sessions.Register(sessionID, entityID)

	log.Printf("Player created and registered - Session: %s, EntityID: %d", sessionID, entityID)
	log.Printf("Total players in room: %d", r.PlayerCount())

	return nil
}

// SendStaticData sends static map data (walls, dimensions) to specific clients
func (r *Room) SendStaticData(sessionIDs []string) {
	staticData := r.game.Statics()

	payloadBytes, err := json.Marshal(staticData)
	if err != nil {
		return
	}

	envelope := ports.ResponseEnvelope{
		EnvelopeType: ports.StaticDataEnvelope,
		Payload:      json.RawMessage(payloadBytes),
	}

	r.outgoing <- UpdateMessage{
		ToSessions: sessionIDs,
		Envelope:   envelope,
	}
}

// RemovePlayer removes a player from the game state and the registry.
func (r *Room) RemovePlayer(sessionID string) {
	if entityID, ok := r.sessions.EntityID(sessionID); ok {
		r.sessions.Unregister(sessionID)
		log.Printf("Player EntityID %d (Session %s) removed from room %s", entityID, sessionID, r.ID)
	}
}

func (r *Room) broadcastGameUpdate() {
	gameUpdate := r.game.Statics() // TODO: check if this should be static data or delta data after updating.

	payloadBytes, err := json.Marshal(gameUpdate)
	if err != nil {
		log.Printf("Failed to marshal game update: %v", err)
		return
	}

	envelope := ports.ResponseEnvelope{
		EnvelopeType: ports.GameUpdateEnvelope,
		Payload:      json.RawMessage(payloadBytes),
	}

	r.outgoing <- UpdateMessage{
		ToSessions: r.sessions.AllSessionIDs(),
		Envelope:   envelope,
	}
}

func (r *Room) PlayerCount() int {
	return r.sessions.Count()
}

func (r *Room) Name() string {
	if r.mapConfig != nil {
		return r.mapConfig.Name
	}
	return r.ID
}

func (r *Room) MaxPlayers() int {
	return 0
}
