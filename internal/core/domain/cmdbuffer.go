package domain

type WorldCommandType int

const (
	CreateEntityCommand WorldCommandType = iota
	DestroyEntityCommand
)

type WorldCommand struct {
	// TODO: define properties
}

type CommandBuffer struct {
	commands []WorldCommand
}

func NewCommandBuffer() *CommandBuffer {
	return &CommandBuffer{
		commands: make([]WorldCommand, 0, 1024), // cap 1024 for no reason, just donno what to put
	}
}

func (cb *CommandBuffer) Pop() WorldCommand {
	panic("not implemented")
}

func (cb *CommandBuffer) Push(cmd WorldCommand) {
	panic("not implemented")
}

func (cb *CommandBuffer) Clear() {
	panic("not implemented")
}
