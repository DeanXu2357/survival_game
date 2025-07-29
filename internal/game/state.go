package game

import (
	"fmt"
	"sync"
)

type State struct {
	Players          map[string]*Player
	playerMu         sync.RWMutex // Mutex to protect access to Players map
	Walls            []*Wall
	Projectiles      []*Projectile
	ObjectGrid       *Grid `json:"-"` // Exclude from JSON serialization
	allowToAddPlayer bool
}

func NewGameState() *State {
	return &State{
		Players:          make(map[string]*Player),
		Walls:            make([]*Wall, 0),
		Projectiles:      make([]*Projectile, 0),
		allowToAddPlayer: true,
		ObjectGrid: &Grid{
			CellSize: 50.0,
			Cells:    make(map[GridCoord][]MapObject),
		},
	}
}

func (s *State) ToClientState() *ClientGameState {
	return &ClientGameState{
		Players:     s.Players,
		Walls:       s.Walls,
		Projectiles: s.Projectiles,
		Timestamp:   0, // TODO: Add proper timestamp
	}
}

func (s *State) generatePlayerID() string {
	// This function should generate a unique player ID.
	// For simplicity, we can use a simple counter or UUID.
	// Here we assume a function that generates a unique ID.
	return fmt.Sprintf("player-%d", len(s.Players)+1)
}

func (s *State) NewPlayer() (*Player, error) {
	if !s.allowToAddPlayer {
		return nil, fmt.Errorf("adding new players is not allowed")
	}

	s.playerMu.Lock()
	defer s.playerMu.Unlock()

	playerID := s.generatePlayerID()
	if _, exists := s.Players[playerID]; exists {
		return nil, fmt.Errorf("player with ID %s already exists", playerID)
	}

	player := &Player{
		ID:            playerID,
		Position:      Vector2D{X: 400, Y: 300}, // Center of 800x600 map
		Direction:     0,
		Radius:        10,
		RotationSpeed: playerBaseRotationSpeed,
		MovementSpeed: playerBaseMovementSpeed * 60, // 60 pixels/second
		Health:        100,
		IsAlive:       true,
	}
	s.Players[playerID] = player
	return player, nil
}

type MapObject interface {
	ID() string
	Position() Vector2D
	IsRectangle() bool
	BoundingBox() (Vector2D, Vector2D)
}

type GridCoord struct {
	X int
	Y int
}

type ClientGameState struct {
	Players     map[string]*Player `json:"players"`
	Walls       []*Wall            `json:"walls"`
	Projectiles []*Projectile      `json:"projectiles"`
	Timestamp   int64              `json:"timestamp"`
}
