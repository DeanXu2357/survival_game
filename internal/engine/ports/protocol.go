package ports

import (
	"encoding/json"
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

type MovementType uint8

const (
	MovementTypeAbsolute MovementType = 0
	MovementTypeRelative MovementType = 1
)

type RequestEnvelope struct {
	EnvelopeType RequestEnvelopeType `json:"envelope_type"`
	Payload      json.RawMessage     `json:"payload"`
}

type ResponseEnvelope struct {
	EnvelopeType ResponseEnvelopeType `json:"envelope_type"`
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
	SessionID string
	Input     PlayerInput
}

type RequestCommand struct {
	ClientID      string
	EnvelopeType  RequestEnvelopeType
	Payload       json.RawMessage
	ParsedPayload any
	ReceivedTime  time.Time
}

type PlayerInput struct {
	MoveVertical   float64      `json:"MoveVertical"`
	MoveHorizontal float64      `json:"MoveHorizontal"`
	LookHorizontal float64      `json:"LookHorizontal"`
	MovementType   MovementType `json:"MovementType"`
	SwitchWeapon   bool         `json:"SwitchWeapon"`
	Reload         bool         `json:"Reload"`
	FastReload     bool         `json:"FastReload"`
	Fire           bool         `json:"Fire"`
	Timestamp      int64        `json:"Timestamp"`
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

type ErrorPayload struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type GameUpdatePayload struct {
	Me        PlayerInfo   `json:"me"`
	Views     []PlayerInfo `json:"views"`
	Timestamp int64        `json:"timestamp"` // timestamp unix milli
}

type PlayerInfo struct {
	ID  uint64  `json:"id"`
	X   float64 `json:"x"`
	Y   float64 `json:"y"`
	Dir float64 `json:"dir"`
}

type StaticDataPayload struct {
	Colliders []Collider `json:"colliders"`
	MapWidth  float64    `json:"map_width"`
	MapHeight float64    `json:"map_height"`
}

type Collider struct {
	ID            uint64  `json:"id"`
	X             float64 `json:"x"`
	Y             float64 `json:"y"`
	HalfX         float64 `json:"half_x"`
	HalfY         float64 `json:"half_y"`
	Radius        float64 `json:"radius"`
	ShapeType     uint8   `json:"shapeType"`
	Rotation      float64 `json:"rotation"`
	Height        float64 `json:"height"`
	BaseElevation float64 `json:"base_elevation"`
}
