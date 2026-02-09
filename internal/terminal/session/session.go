package session

import (
	"survival/internal/engine"
	"survival/internal/engine/ports"
	"survival/internal/engine/state"
)

type GameSession struct {
	game     *engine.Game
	entityID state.EntityID
}

func NewGameSession(game *engine.Game, entityID state.EntityID) *GameSession {
	return &GameSession{
		game:     game,
		entityID: entityID,
	}
}

func (s *GameSession) SetInput(input ports.PlayerInput) {
	s.game.SetPlayerInput(s.entityID, input)
}

func (s *GameSession) Update(dt float64) {
	s.game.Update(dt)
}

func (s *GameSession) PlayerState() (x, y, dir float64, ok bool) {
	snapshot, exists := s.game.PlayerSnapshotWithLocation(s.entityID)
	if !exists {
		return 0, 0, 0, false
	}
	return snapshot.Player.Position.X, snapshot.Player.Position.Y, float64(snapshot.Player.Direction), true
}

func (s *GameSession) Statics() []state.StaticEntity {
	return s.game.Statics()
}
