package domain

type WorldCommandType int

const (
	CreateEntityCommand WorldCommandType = iota
	DestroyEntityCommand
	UpdatePlayerCommand
)

type WorldCommand struct {
	Type     WorldCommandType
	EntityID EntityID

	Position  *Position
	Direction *Direction
	Meta      *Meta
}

func NewPlayerUpdateCommand(id EntityID, pos *Position, dir *Direction, meta *Meta) WorldCommand {
	return WorldCommand{
		Type:      UpdatePlayerCommand,
		EntityID:  id,
		Position:  pos,
		Direction: dir,
		Meta:      meta,
	}
}

type CommandBuffer struct {
	commands []WorldCommand
}

func NewCommandBuffer() *CommandBuffer {
	return &CommandBuffer{
		commands: make([]WorldCommand, 0, 1024), // cap 1024 for no reason, just donno what to put
	}
}

func (cb *CommandBuffer) Push(cmd WorldCommand) {
	cb.commands = append(cb.commands, cmd)
}

func (cb *CommandBuffer) Pop() (WorldCommand, bool) {
	if len(cb.commands) == 0 {
		return WorldCommand{}, false
	}
	cmd := cb.commands[0]
	cb.commands = cb.commands[1:]
	return cmd, true
}

func (cb *CommandBuffer) Clear() {
	cb.commands = cb.commands[:0]
}

func (cb *CommandBuffer) Len() int {
	return len(cb.commands)
}

func (cb *CommandBuffer) IsEmpty() bool {
	return len(cb.commands) == 0
}
