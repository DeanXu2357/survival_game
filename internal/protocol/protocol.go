package protocol

import (
	"encoding/json"
)

const (
	PlayerInputEnvelope RequestEnvelopeType = "player_input"

	GameUpdateEnvelope       ResponseEnvelopeType = "game_update"
	SystemNotifyEnvelop      ResponseEnvelopeType = "system_notify"
	SystemSetSessionEnvelope ResponseEnvelopeType = "system_set_session"
)

type ResponseEnvelopeType string

type RequestEnvelopeType string

type RequestEnvelope struct {
	Type    RequestEnvelopeType `json:"type"`
	Payload json.RawMessage
}

type ResponseEnvelope struct {
	Type    ResponseEnvelopeType `json:"type"`
	Payload json.RawMessage      `json:"payload"`
}

// Codec is an interface for encoding and decoding messages.
type Codec interface {
	Encode(data interface{}) ([]byte, error)
	Decode(data []byte, v interface{}) error
}

// RawConnection is an interface for raw connection operations.
type RawConnection interface {
	ReadMessage() ([]byte, error)
	WriteMessage(data []byte) error
	Close() error
}

type Command struct {
	ClientID string
	Input    PlayerInput
}

type PlayerInput struct {
	MoveUp       bool
	MoveDown     bool
	MoveLeft     bool
	MoveRight    bool
	RotateLeft   bool
	RotateRight  bool
	SwitchWeapon bool
	Reload       bool
	FastReload   bool
	Fire         bool
}

type SystemNotify struct {
	Message string `json:"message"`
}

type SystemSetSessionPayload struct {
	ClientID  string `json:"client_id"`
	SessionID string `json:"session_id"`
}
