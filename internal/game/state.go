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

func NewGameStateFromMap(mapConfig *MapConfig) *State {
	state := &State{
		Players:          make(map[string]*Player),
		Walls:            make([]*Wall, 0),
		Projectiles:      make([]*Projectile, 0),
		allowToAddPlayer: true,
		ObjectGrid: &Grid{
			CellSize: mapConfig.GridSize,
			Cells:    make(map[GridCoord][]MapObject),
		},
	}

	// Create walls from map configuration
	for _, wallConfig := range mapConfig.Walls {
		wall := NewWall(wallConfig.ID, wallConfig.Center, wallConfig.HalfSize, wallConfig.Rotation)
		state.Walls = append(state.Walls, wall)
		state.ObjectGrid.AddObject(wall)
	}

	return state
}

func (s *State) ToClientState() *ClientGameState {
	wallDTOs := make([]*WallDTO, len(s.Walls))
	for i, wall := range s.Walls {
		wallDTOs[i] = &WallDTO{
			ID:       wall.ID(),
			Center:   wall.Center,
			HalfSize: wall.HalfSize,
			Rotation: wall.Rotation,
		}
	}

	return &ClientGameState{
		Players:     s.Players,
		Walls:       wallDTOs,
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
	return s.NewPlayerAtPosition(Vector2D{X: 400, Y: 300}) // Default center position
}

func (s *State) NewPlayerAtPosition(position Vector2D) (*Player, error) {
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
		Position:      position,
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

// WallDTO is a data transfer object for Wall data sent to the client.
type WallDTO struct {
	ID       string   `json:"id"`
	Center   Vector2D `json:"center"`
	HalfSize Vector2D `json:"half_size"`
	Rotation float64  `json:"rotation"`
}

type ClientGameState struct {
	Players     map[string]*Player `json:"players"`
	Walls       []*WallDTO         `json:"walls"`
	Projectiles []*Projectile      `json:"projectiles"`
	Timestamp   int64              `json:"timestamp"`
}
