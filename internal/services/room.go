package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"survival/internal/engine"
	"survival/internal/engine/ports"
	"survival/internal/utils"
)

type UpdateMessage struct {
	ToSessions []string // an empty slice means broadcast to all
	Envelope   ports.ResponseEnvelope
}

type Room struct {
	ID         string
	mapConfig  *engine.MapConfig
	game       *engine.Game
	sessions   *SessionRegistry
	subManager *Manager[UpdateMessage]

	joinClientCh chan Client
	commands     chan ports.Command
	outgoing     chan UpdateMessage

	ctx    context.Context
	cancel context.CancelFunc
}

func NewRoom(ctx context.Context, id string) (*Room, error) {
	return NewRoomWithMap(ctx, id, engine.DefaultMapConfig())
}

func NewRoomWithMap(ctx context.Context, id string, mapConfig *engine.MapConfig) (*Room, error) {
	roomCTX, cancel := context.WithCancel(ctx)

	game, err := engine.NewGame(mapConfig)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create game: %w", err)
	}

	return &Room{
		ID:         id,
		mapConfig:  mapConfig,
		game:       game,
		sessions:   NewSessionRegistry(),
		subManager: NewManager[UpdateMessage](utils.NewSequentialIDGenerator(fmt.Sprintf("room%s-sub-", id))),

		joinClientCh: make(chan Client, 100),
		commands:     make(chan ports.Command, 200),
		outgoing:     make(chan UpdateMessage, 400),

		ctx:    roomCTX,
		cancel: cancel,
	}, nil
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

	go r.responsePump()

	ticker := time.NewTicker(time.Second / ports.TargetTickRate)
	defer ticker.Stop()

	log.Printf("[Room %s] started", r.ID)

	for {
		select {
		case client := <-r.joinClientCh:
			if err := r.addPlayer(client); err != nil {
				log.Printf("Failed to add player to room %s: %v", r.ID, err)
			}
		case cmd := <-r.commands:
			entityID, ok := r.sessions.EntityID(cmd.SessionID)
			if !ok {
				log.Printf("Warning: No entity ID found for session %s", cmd.SessionID)
				continue
			}
			r.game.SetPlayerInput(entityID, cmd.Input)
		case <-ticker.C:
			r.game.Update(ports.DeltaTime)
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
	select {
	case r.joinClientCh <- client:
		return nil
	default:
		return fmt.Errorf("room %s system command channel full, cannot add player %s", r.ID, client.SessionID())
	}
}

func (r *Room) addPlayer(client Client) error {
	log.Printf("Adding player for session %s to room %s", client.SessionID(), r.ID)
	sessionID := client.SessionID()
	if _, exist := r.sessions.EntityID(sessionID); exist {
		return fmt.Errorf("session %s already registered in room %s", sessionID, r.ID)
	}

	// TODO: check reconnection logic exist ?
	entityID, err := r.game.JoinPlayer()
	if err != nil {
		return fmt.Errorf("failed to join player for session %s in room %s: %w", sessionID, r.ID, err)
	}

	r.sessions.Register(sessionID, entityID)

	log.Printf("Player created and registered - Session: %s, EntityID: %d", sessionID, entityID)
	log.Printf("Total players in room: %d", r.PlayerCount())
	return nil
}

// SendStaticData sends static map data (walls, dimensions) to specific clients
func (r *Room) SendStaticData(sessionIDs []string) {
	staticData := r.game.Statics()
	mapInfo := r.game.MapInfo()

	log.Printf("[SendStaticData] Sending %d colliders, map: %.0fx%.0f to sessions: %v",
		len(staticData), mapInfo.Width, mapInfo.Height, sessionIDs)

	colliders := make([]ports.Collider, len(staticData))
	for i, entity := range staticData {
		colliders[i] = ports.Collider{
			ID:            uint64(entity.ID),
			X:             entity.Collider.Center.X,
			Y:             entity.Collider.Center.Y,
			HalfX:         entity.Collider.HalfSize.X,
			HalfY:         entity.Collider.HalfSize.Y,
			Radius:        entity.Collider.Radius,
			ShapeType:     uint8(entity.Collider.ShapeType),
			Rotation:      0,
			Height:        entity.VerticalBody.Height,
			BaseElevation: entity.VerticalBody.BaseElevation,
		}
	}

	payloadBytes, err := json.Marshal(ports.StaticDataPayload{
		Colliders: colliders,
		MapWidth:  mapInfo.Width,
		MapHeight: mapInfo.Height,
	})
	if err != nil {
		log.Printf("Failed to marshal static data: %v", err)
		return
	}

	envelope := ports.ResponseEnvelope{
		EnvelopeType: ports.StaticDataEnvelope,
		Payload:      payloadBytes,
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
	for entityID, sessionID := range r.sessions.All() {
		snapshot, exist := r.game.PlayerSnapshotWithLocation(entityID)
		if !exist {
			log.Printf("[broadcastGameUpdate] Snapshot not found for entityID: %d, sessionID: %s", entityID, sessionID)
			continue
		}

		viewInfo := make([]ports.PlayerInfo, len(snapshot.Views))
		if len(snapshot.Views) > 0 {
			for i, view := range snapshot.Views {
				viewInfo[i] = ports.PlayerInfo{
					ID:  uint64(view.ID),
					X:   view.Position.X,
					Y:   view.Position.Y,
					Dir: float64(view.Direction),
				}
			}
		}

		bytes, err := json.Marshal(ports.GameUpdatePayload{
			Me: ports.PlayerInfo{
				ID:  uint64(entityID),
				X:   snapshot.Player.Position.X,
				Y:   snapshot.Player.Position.Y,
				Dir: float64(snapshot.Player.Direction),
			},
			Views:     viewInfo,
			Timestamp: time.Now().UnixMilli(),
		})
		if err != nil {
			// TODO: log not find
			fmt.Println("Failed to marshal game update payload:", err)
			continue
		}

		r.outgoing <- UpdateMessage{
			ToSessions: []string{sessionID},
			Envelope: ports.ResponseEnvelope{
				EnvelopeType: ports.GameUpdateEnvelope,
				Payload:      bytes,
			},
		}

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
