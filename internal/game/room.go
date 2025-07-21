package game

import (
	"context"
	"time"
)

type Room struct {
	ID       string
	state    *State
	logic    *Logic
	commands chan Command
}

func NewRoom(id string) *Room {
	return &Room{
		ID:       id,
		state:    NewGameState(),
		logic:    NewGameLogic(),
		commands: make(chan Command, 100),
	}
}

func (r *Room) Run(ctx context.Context) {
	ticker := time.NewTicker(time.Second / targetTickRate)
	defer ticker.Stop()

	currentInputs := make(map[string]PlayerInput)

	for {
		select {
		case cmd := <-r.commands:
			currentInputs[cmd.FromPlayer] = cmd.Input
		case <-ticker.C:
			r.logic.Update(r.state, currentInputs, deltaTime)

		case <-ctx.Done():
			// todo: log message ?
			return
		}
	}
}
func (r *Room) AddCommand(cmd Command) bool {
	select {
	case r.commands <- cmd:
		return true
	default:
		return false // Channel is full, command not added
	}
}

type Command struct {
	FromPlayer string
	Input      PlayerInput
}
