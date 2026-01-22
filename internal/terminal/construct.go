package terminal

import (
	"bytes"
	"time"
)

const (
	HideCursor  = "\033[?25l"
	ShowCursor  = "\033[?25h"
	ResetCursor = "\033[H"
	ClearScreen = "\033[2J"
	ClearToEnd  = "\033[J"
)

type GameState interface {
	Init()

	Update(input InputEvent, dt time.Duration) Command

	Draw(buf *bytes.Buffer, width, height int)
}

type KeyEvent []byte

type InputEvent int

const (
	InputNone InputEvent = iota
	InputMoveForward
	InputMoveBackward
	InputMoveLeft
	InputMoveRight
	InputTurnLeft
	InputTurnRight
	InputAction
	InputCancel
)
