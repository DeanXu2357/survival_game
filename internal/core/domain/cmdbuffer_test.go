package domain

import (
	"testing"
)

func TestNewCommandBuffer(t *testing.T) {
	buf := NewCommandBuffer()

	if buf == nil {
		t.Fatal("NewCommandBuffer() returned nil")
	}
	if buf.Len() != 0 {
		t.Errorf("Len() = %d, want 0", buf.Len())
	}
	if !buf.IsEmpty() {
		t.Error("IsEmpty() = false, want true")
	}
}

func TestCommandBuffer_Push(t *testing.T) {
	tests := []struct {
		name     string
		commands []WorldCommand
		wantLen  int
	}{
		{
			name:     "push single command",
			commands: []WorldCommand{{Type: CreateEntityCommand, EntityID: 1}},
			wantLen:  1,
		},
		{
			name: "push multiple commands",
			commands: []WorldCommand{
				{Type: CreateEntityCommand, EntityID: 1},
				{Type: DestroyEntityCommand, EntityID: 2},
				{Type: UpdatePlayerCommand, EntityID: 3},
			},
			wantLen: 3,
		},
		{
			name:     "push zero commands",
			commands: []WorldCommand{},
			wantLen:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := NewCommandBuffer()
			for _, cmd := range tt.commands {
				buf.Push(cmd)
			}
			if buf.Len() != tt.wantLen {
				t.Errorf("Len() = %d, want %d", buf.Len(), tt.wantLen)
			}
		})
	}
}

func TestCommandBuffer_Pop(t *testing.T) {
	tests := []struct {
		name       string
		commands   []WorldCommand
		wantCmd    WorldCommand
		wantOk     bool
		wantLenAfter int
	}{
		{
			name:       "pop from single command buffer",
			commands:   []WorldCommand{{Type: CreateEntityCommand, EntityID: 1}},
			wantCmd:    WorldCommand{Type: CreateEntityCommand, EntityID: 1},
			wantOk:     true,
			wantLenAfter: 0,
		},
		{
			name: "pop from multiple command buffer",
			commands: []WorldCommand{
				{Type: CreateEntityCommand, EntityID: 1},
				{Type: DestroyEntityCommand, EntityID: 2},
			},
			wantCmd:    WorldCommand{Type: CreateEntityCommand, EntityID: 1},
			wantOk:     true,
			wantLenAfter: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := NewCommandBuffer()
			for _, cmd := range tt.commands {
				buf.Push(cmd)
			}

			gotCmd, gotOk := buf.Pop()
			if gotOk != tt.wantOk {
				t.Errorf("Pop() ok = %v, want %v", gotOk, tt.wantOk)
			}
			if gotCmd.Type != tt.wantCmd.Type || gotCmd.EntityID != tt.wantCmd.EntityID {
				t.Errorf("Pop() cmd = %+v, want %+v", gotCmd, tt.wantCmd)
			}
			if buf.Len() != tt.wantLenAfter {
				t.Errorf("Len() after Pop() = %d, want %d", buf.Len(), tt.wantLenAfter)
			}
		})
	}
}

func TestCommandBuffer_Pop_Empty(t *testing.T) {
	buf := NewCommandBuffer()

	cmd, ok := buf.Pop()
	if ok {
		t.Error("Pop() on empty buffer returned ok = true, want false")
	}
	if cmd.Type != 0 || cmd.EntityID != 0 {
		t.Errorf("Pop() on empty buffer returned non-zero command: %+v", cmd)
	}
}

func TestCommandBuffer_Clear(t *testing.T) {
	tests := []struct {
		name     string
		commands []WorldCommand
	}{
		{
			name:     "clear empty buffer",
			commands: []WorldCommand{},
		},
		{
			name:     "clear single command buffer",
			commands: []WorldCommand{{Type: CreateEntityCommand, EntityID: 1}},
		},
		{
			name: "clear multiple command buffer",
			commands: []WorldCommand{
				{Type: CreateEntityCommand, EntityID: 1},
				{Type: DestroyEntityCommand, EntityID: 2},
				{Type: UpdatePlayerCommand, EntityID: 3},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := NewCommandBuffer()
			for _, cmd := range tt.commands {
				buf.Push(cmd)
			}

			buf.Clear()

			if buf.Len() != 0 {
				t.Errorf("Len() after Clear() = %d, want 0", buf.Len())
			}
			if !buf.IsEmpty() {
				t.Error("IsEmpty() after Clear() = false, want true")
			}
		})
	}
}

func TestCommandBuffer_FIFO_Order(t *testing.T) {
	buf := NewCommandBuffer()

	commands := []WorldCommand{
		{Type: CreateEntityCommand, EntityID: 1},
		{Type: UpdatePlayerCommand, EntityID: 2},
		{Type: UpdatePlayerCommand, EntityID: 3},
		{Type: DestroyEntityCommand, EntityID: 4},
	}

	for _, cmd := range commands {
		buf.Push(cmd)
	}

	for i, wantCmd := range commands {
		gotCmd, ok := buf.Pop()
		if !ok {
			t.Fatalf("Pop() #%d returned ok = false, expected command", i)
		}
		if gotCmd.Type != wantCmd.Type || gotCmd.EntityID != wantCmd.EntityID {
			t.Errorf("Pop() #%d = %+v, want %+v", i, gotCmd, wantCmd)
		}
	}

	if !buf.IsEmpty() {
		t.Errorf("buffer not empty after popping all commands, Len() = %d", buf.Len())
	}
}

func TestCommandBuffer_WithOptionalFields(t *testing.T) {
	pos := Position{X: 100.0, Y: 200.0}
	dir := Direction(1.57)
	meta := Meta(0x01)

	tests := []struct {
		name string
		cmd  WorldCommand
	}{
		{
			name: "update player with position only",
			cmd: WorldCommand{
				Type:     UpdatePlayerCommand,
				EntityID: 1,
				Position: &pos,
			},
		},
		{
			name: "update player with direction only",
			cmd: WorldCommand{
				Type:      UpdatePlayerCommand,
				EntityID:  2,
				Direction: &dir,
			},
		},
		{
			name: "update player with all fields",
			cmd: WorldCommand{
				Type:      UpdatePlayerCommand,
				EntityID:  3,
				Position:  &pos,
				Direction: &dir,
				Meta:      &meta,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := NewCommandBuffer()
			buf.Push(tt.cmd)

			gotCmd, ok := buf.Pop()
			if !ok {
				t.Fatal("Pop() returned ok = false")
			}

			if gotCmd.Type != tt.cmd.Type {
				t.Errorf("Type = %v, want %v", gotCmd.Type, tt.cmd.Type)
			}
			if gotCmd.EntityID != tt.cmd.EntityID {
				t.Errorf("EntityID = %v, want %v", gotCmd.EntityID, tt.cmd.EntityID)
			}
			if tt.cmd.Position != nil {
				if gotCmd.Position == nil {
					t.Error("Position is nil, expected non-nil")
				} else if *gotCmd.Position != *tt.cmd.Position {
					t.Errorf("Position = %+v, want %+v", *gotCmd.Position, *tt.cmd.Position)
				}
			}
			if tt.cmd.Direction != nil {
				if gotCmd.Direction == nil {
					t.Error("Direction is nil, expected non-nil")
				} else if *gotCmd.Direction != *tt.cmd.Direction {
					t.Errorf("Direction = %v, want %v", *gotCmd.Direction, *tt.cmd.Direction)
				}
			}
			if tt.cmd.Meta != nil {
				if gotCmd.Meta == nil {
					t.Error("Meta is nil, expected non-nil")
				} else if *gotCmd.Meta != *tt.cmd.Meta {
					t.Errorf("Meta = %v, want %v", *gotCmd.Meta, *tt.cmd.Meta)
				}
			}
		})
	}
}

func TestCommandBuffer_Len(t *testing.T) {
	buf := NewCommandBuffer()

	if buf.Len() != 0 {
		t.Errorf("initial Len() = %d, want 0", buf.Len())
	}

	buf.Push(WorldCommand{Type: CreateEntityCommand, EntityID: 1})
	if buf.Len() != 1 {
		t.Errorf("Len() after 1 push = %d, want 1", buf.Len())
	}

	buf.Push(WorldCommand{Type: CreateEntityCommand, EntityID: 2})
	if buf.Len() != 2 {
		t.Errorf("Len() after 2 pushes = %d, want 2", buf.Len())
	}

	buf.Pop()
	if buf.Len() != 1 {
		t.Errorf("Len() after 1 pop = %d, want 1", buf.Len())
	}

	buf.Clear()
	if buf.Len() != 0 {
		t.Errorf("Len() after clear = %d, want 0", buf.Len())
	}
}

func TestCommandBuffer_IsEmpty(t *testing.T) {
	buf := NewCommandBuffer()

	if !buf.IsEmpty() {
		t.Error("IsEmpty() on new buffer = false, want true")
	}

	buf.Push(WorldCommand{Type: CreateEntityCommand, EntityID: 1})
	if buf.IsEmpty() {
		t.Error("IsEmpty() after push = true, want false")
	}

	buf.Pop()
	if !buf.IsEmpty() {
		t.Error("IsEmpty() after pop = false, want true")
	}
}

func TestNewPlayerUpdateCommand(t *testing.T) {
	pos := Position{X: 100.0, Y: 200.0}
	dir := Direction(1.57)
	meta := Meta(0x01)

	tests := []struct {
		name      string
		id        EntityID
		pos       *Position
		dir       *Direction
		meta      *Meta
	}{
		{
			name: "all fields set",
			id:   1,
			pos:  &pos,
			dir:  &dir,
			meta: &meta,
		},
		{
			name: "position only",
			id:   2,
			pos:  &pos,
			dir:  nil,
			meta: nil,
		},
		{
			name: "direction only",
			id:   3,
			pos:  nil,
			dir:  &dir,
			meta: nil,
		},
		{
			name: "no optional fields",
			id:   4,
			pos:  nil,
			dir:  nil,
			meta: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewPlayerUpdateCommand(tt.id, tt.pos, tt.dir, tt.meta)

			if cmd.Type != UpdatePlayerCommand {
				t.Errorf("Type = %v, want %v", cmd.Type, UpdatePlayerCommand)
			}
			if cmd.EntityID != tt.id {
				t.Errorf("EntityID = %v, want %v", cmd.EntityID, tt.id)
			}
			if cmd.Position != tt.pos {
				t.Errorf("Position = %v, want %v", cmd.Position, tt.pos)
			}
			if cmd.Direction != tt.dir {
				t.Errorf("Direction = %v, want %v", cmd.Direction, tt.dir)
			}
			if cmd.Meta != tt.meta {
				t.Errorf("Meta = %v, want %v", cmd.Meta, tt.meta)
			}
		})
	}
}
