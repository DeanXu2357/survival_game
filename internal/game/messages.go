package game

import "time"

// PlayerSnapshot represents essential player data for client updates
type PlayerSnapshot struct {
	ID        string   `json:"id"`
	Position  Vector2D `json:"position"`
	Direction float64  `json:"direction"`
	Health    int      `json:"health"`
	IsAlive   bool     `json:"isAlive"`
}

// ProjectileSnapshot represents essential projectile data for client updates
type ProjectileSnapshot struct {
	ID        string   `json:"id"`
	Position  Vector2D `json:"position"`
	Direction Vector2D `json:"direction"`
	Speed     float64  `json:"speed"`
}

// GameUpdate is the lightweight message sent to clients during gameplay
type GameUpdate struct {
	Type        string                    `json:"type"`
	Players     map[string]PlayerSnapshot `json:"players"`
	Projectiles []ProjectileSnapshot      `json:"projectiles"`
	Timestamp   int64                     `json:"timestamp"`
}

// StaticGameData contains data sent once on connection (walls, map layout, etc.)
type StaticGameData struct {
	Type      string     `json:"type"`
	Walls     []*WallDTO `json:"walls"`
	MapWidth  int        `json:"mapWidth"`
	MapHeight int        `json:"mapHeight"`
}

// Convert game state to lightweight update message
func (s *State) ToGameUpdate() *GameUpdate {
	playerSnapshots := make(map[string]PlayerSnapshot)
	for id, player := range s.Players {
		playerSnapshots[id] = PlayerSnapshot{
			ID:        player.ID,
			Position:  player.Position,
			Direction: player.Direction,
			Health:    player.Health,
			IsAlive:   player.IsAlive,
		}
	}

	projectileSnapshots := make([]ProjectileSnapshot, len(s.Projectiles))
	for i, proj := range s.Projectiles {
		projectileSnapshots[i] = ProjectileSnapshot{
			ID:        proj.ID,
			Position:  proj.Position,
			Direction: proj.Direction,
			Speed:     proj.Speed,
		}
	}

	return &GameUpdate{
		Type:        "gameUpdate",
		Players:     playerSnapshots,
		Projectiles: projectileSnapshots,
		Timestamp:   time.Now().UnixMilli(),
	}
}

// ToStaticData Convert walls to static data message (sent once on connect)
func (s *State) ToStaticData() *StaticGameData {
	// Convert walls to DTOs for client transmission
	wallDTOs := make([]*WallDTO, len(s.Walls))
	for i, wall := range s.Walls {
		wallDTOs[i] = &WallDTO{
			ID:       wall.GetID(),
			Center:   wall.Center,
			HalfSize: wall.HalfSize,
			Rotation: wall.Rotation,
		}
	}

	return &StaticGameData{
		Type:      "staticData",
		Walls:     wallDTOs,
		MapWidth:  800, // TODO: Make configurable
		MapHeight: 600, // TODO: Make configurable
	}
}
