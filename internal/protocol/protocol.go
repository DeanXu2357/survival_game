package protocol

import (
	"encoding/json"
	"fmt"
	"time"
)

const (
	PlayerInputEnvelope RequestEnvelopeType = "player_input"
	ListRoomsEnvelope   RequestEnvelopeType = "list_rooms"
	RequestJoinEnvelope RequestEnvelopeType = "request_join"

	GameUpdateEnvelope        ResponseEnvelopeType = "game_update"
	StaticDataEnvelope        ResponseEnvelopeType = "static_data"
	SystemNotifyEnvelop       ResponseEnvelopeType = "system_notify"
	SystemSetSessionEnvelope  ResponseEnvelopeType = "system_set_session"
	ErrInvalidSession         ResponseEnvelopeType = "error_invalid_session"
	ListRoomsResponseEnvelope ResponseEnvelopeType = "list_rooms_response"
	ErrorResponseEnvelope     ResponseEnvelopeType = "error"
	JoinRoomSuccessEnvelope   ResponseEnvelopeType = "join_room_success"
)

type ResponseEnvelopeType string

type RequestEnvelopeType string

type RequestEnvelope struct {
	Type    RequestEnvelopeType `json:"type"`
	Payload json.RawMessage
}

type ResponseEnvelope struct {
	EnvelopeType ResponseEnvelopeType `json:"envelopetype"`
	Payload      json.RawMessage      `json:"payload"`
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

type RequestCommand struct {
	ClientID      string
	EnvelopeType  RequestEnvelopeType
	Payload       json.RawMessage
	ParsedPayload any
	ReceivedTime  time.Time
}

type PlayerInput struct {
	MoveUp       bool  `json:"MoveUp"`
	MoveDown     bool  `json:"MoveDown"`
	MoveLeft     bool  `json:"MoveLeft"`
	MoveRight    bool  `json:"MoveRight"`
	RotateLeft   bool  `json:"RotateLeft"`
	RotateRight  bool  `json:"RotateRight"`
	SwitchWeapon bool  `json:"SwitchWeapon"`
	Reload       bool  `json:"Reload"`
	FastReload   bool  `json:"FastReload"`
	Fire         bool  `json:"Fire"`
	Timestamp    int64 `json:"Timestamp"`
}

type RequestJoinPayload struct {
	RoomID string `json:"room_id"`
}

type ListRoomsPayload struct {
	// No fields needed for listing rooms
}

type ListRoomsResponse struct {
	Rooms []RoomInfo `json:"rooms"`
}

type RoomInfo struct {
	RoomID      string `json:"room_id"`
	Name        string `json:"name"`
	PlayerCount int    `json:"player_count"`
	MaxPlayers  int    `json:"max_players"`
}

type SystemNotify struct {
	Message string `json:"message"`
}

type SystemSetSessionPayload struct {
	ClientID  string `json:"client_id"`
	SessionID string `json:"session_id"`
}

func GetPayloadStruct(envelopeType RequestEnvelopeType) (any, error) {
	switch envelopeType {
	case PlayerInputEnvelope:
		return PlayerInput{}, nil
	default:
		return nil, fmt.Errorf("unknown envelope type: %s", envelopeType)
	}
}

type ErrorPayload struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
