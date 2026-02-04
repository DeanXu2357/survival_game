package state

import "sync"

type WorldCommand struct {
	EntityID   EntityID
	UpdateMeta Meta

	Position      Position
	Direction     Direction
	Meta          Meta
	RotationSpeed RotationSpeed
	MovementSpeed MovementSpeed
	PlayerShape   PlayerHitbox
	Health        Health
	Collider      Collider
	VerticalBody  VerticalBody
	Input         Input
	PrePosition   PrePosition
}

// CommandBuffer is a thread-safe buffer for WorldCommands.
// It allows multiple goroutines to push and pop commands concurrently.
type CommandBuffer struct {
	mu       sync.RWMutex
	commands []WorldCommand
}

func NewCommandBuffer() *CommandBuffer {
	return &CommandBuffer{
		commands: make([]WorldCommand, 0, 1024), // cap 1024 for no reason, just donno what to put
	}
}

func (cb *CommandBuffer) Push(cmd WorldCommand) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.commands = append(cb.commands, cmd)
}

func (cb *CommandBuffer) Pop() (WorldCommand, bool) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if len(cb.commands) == 0 {
		return WorldCommand{}, false
	}
	cmd := cb.commands[0]
	cb.commands = cb.commands[1:]
	return cmd, true
}

func (cb *CommandBuffer) Clear() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.commands = cb.commands[:0]
}

func (cb *CommandBuffer) Len() int {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	return len(cb.commands)
}

func (cb *CommandBuffer) IsEmpty() bool {
	return cb.Len() == 0
}
