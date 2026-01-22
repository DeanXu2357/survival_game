package network

import (
	"encoding/json"
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"survival/internal/engine/ports"
)

type ConnectionState int

const (
	StateDisconnected ConnectionState = iota
	StateConnecting
	StateConnected
	StateError
)

type Client struct {
	conn      *websocket.Conn
	state     ConnectionState
	stateMu   sync.RWMutex
	clientID  string
	sessionID string
	lastError error

	gameUpdateChan  chan ports.GameUpdatePayload
	staticDataChan  chan ports.StaticDataPayload
	roomListChan    chan ports.ListRoomsResponse
	joinSuccessChan chan string
	errorChan       chan error

	closeChan chan struct{}
	closeOnce sync.Once
}

func NewClient(clientID string) *Client {
	return &Client{
		clientID:        clientID,
		state:           StateDisconnected,
		gameUpdateChan:  make(chan ports.GameUpdatePayload, 10),
		staticDataChan:  make(chan ports.StaticDataPayload, 1),
		roomListChan:    make(chan ports.ListRoomsResponse, 1),
		joinSuccessChan: make(chan string, 1),
		errorChan:       make(chan error, 10),
		closeChan:       make(chan struct{}),
	}
}

func (c *Client) Connect(serverAddr string, playerName string) error {
	c.stateMu.Lock()
	c.state = StateConnecting
	c.stateMu.Unlock()

	u := url.URL{
		Scheme:   "ws",
		Host:     serverAddr,
		Path:     "/ws",
		RawQuery: fmt.Sprintf("client_id=%s&name=%s", url.QueryEscape(c.clientID), url.QueryEscape(playerName)),
	}

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		c.stateMu.Lock()
		c.state = StateError
		c.lastError = err
		c.stateMu.Unlock()
		return err
	}

	c.conn = conn
	c.stateMu.Lock()
	c.state = StateConnected
	c.stateMu.Unlock()

	go c.readLoop()

	return nil
}

func (c *Client) readLoop() {
	defer func() {
		c.stateMu.Lock()
		if c.state == StateConnected {
			c.state = StateDisconnected
		}
		c.stateMu.Unlock()
	}()

	for {
		select {
		case <-c.closeChan:
			return
		default:
		}

		_, message, err := c.conn.ReadMessage()
		if err != nil {
			c.stateMu.Lock()
			c.state = StateError
			c.lastError = err
			c.stateMu.Unlock()
			select {
			case c.errorChan <- err:
			default:
			}
			return
		}

		var envelope ports.ResponseEnvelope
		if err := json.Unmarshal(message, &envelope); err != nil {
			continue
		}

		c.handleMessage(envelope)
	}
}

func (c *Client) handleMessage(envelope ports.ResponseEnvelope) {
	switch envelope.EnvelopeType {
	case ports.SystemSetSessionEnvelope:
		var payload ports.SystemSetSessionPayload
		if err := json.Unmarshal(envelope.Payload, &payload); err == nil {
			c.sessionID = payload.SessionID
		}

	case ports.GameUpdateEnvelope:
		var payload ports.GameUpdatePayload
		if err := json.Unmarshal(envelope.Payload, &payload); err == nil {
			select {
			case c.gameUpdateChan <- payload:
			default:
			}
		}

	case ports.StaticDataEnvelope:
		var payload ports.StaticDataPayload
		if err := json.Unmarshal(envelope.Payload, &payload); err == nil {
			select {
			case c.staticDataChan <- payload:
			default:
			}
		}

	case ports.ListRoomsResponseEnvelope:
		var payload ports.ListRoomsResponse
		if err := json.Unmarshal(envelope.Payload, &payload); err == nil {
			select {
			case c.roomListChan <- payload:
			default:
			}
		}

	case ports.JoinRoomSuccessEnvelope:
		select {
		case c.joinSuccessChan <- "success":
		default:
		}

	case ports.ErrorResponseEnvelope:
		var payload ports.ErrorPayload
		if err := json.Unmarshal(envelope.Payload, &payload); err == nil {
			select {
			case c.errorChan <- fmt.Errorf("server error: %s", payload.Message):
			default:
			}
		}
	}
}

func (c *Client) RequestRoomList() error {
	return c.sendRequest(ports.ListRoomsEnvelope, ports.ListRoomsPayload{})
}

func (c *Client) RequestJoinRoom(roomID string) error {
	return c.sendRequest(ports.RequestJoinEnvelope, ports.RequestJoinPayload{RoomID: roomID})
}

func (c *Client) SendInput(input ports.PlayerInput) error {
	input.Timestamp = time.Now().UnixMilli()
	return c.sendRequest(ports.PlayerInputEnvelope, input)
}

func (c *Client) sendRequest(envelopeType ports.RequestEnvelopeType, payload interface{}) error {
	c.stateMu.RLock()
	state := c.state
	c.stateMu.RUnlock()

	if state != StateConnected {
		return fmt.Errorf("not connected")
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	envelope := ports.RequestEnvelope{
		EnvelopeType: envelopeType,
		Payload:      payloadBytes,
	}

	data, err := json.Marshal(envelope)
	if err != nil {
		return err
	}

	return c.conn.WriteMessage(websocket.TextMessage, data)
}

func (c *Client) GameUpdateChan() <-chan ports.GameUpdatePayload {
	return c.gameUpdateChan
}

func (c *Client) StaticDataChan() <-chan ports.StaticDataPayload {
	return c.staticDataChan
}

func (c *Client) RoomListChan() <-chan ports.ListRoomsResponse {
	return c.roomListChan
}

func (c *Client) JoinSuccessChan() <-chan string {
	return c.joinSuccessChan
}

func (c *Client) ErrorChan() <-chan error {
	return c.errorChan
}

func (c *Client) State() ConnectionState {
	c.stateMu.RLock()
	defer c.stateMu.RUnlock()
	return c.state
}

func (c *Client) LastError() error {
	c.stateMu.RLock()
	defer c.stateMu.RUnlock()
	return c.lastError
}

func (c *Client) Close() {
	c.closeOnce.Do(func() {
		close(c.closeChan)
		if c.conn != nil {
			c.conn.Close()
		}
	})
}
