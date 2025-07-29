package app

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"

	"survival/internal/protocol"
)

var (
	ErrSendFailed             = errors.New("failed to send message to client")
	ErrClientConnectionClosed = errors.New("client connection is closed")
	ErrClientNotServing       = errors.New("client is not serving requests")
)

type Client interface {
	ID() string
	Name() string
	SessionID() string
	SetSessionID(sessionID string) error
	Send(ctx context.Context, envelope protocol.ResponseEnvelope) error
	SetHubChannel(hubCh chan<- protocol.Command)
	Close(ctx context.Context) error
	Pump() error
}

func newWebsocketClient(ctx context.Context, id, name string, conn protocol.RawConnection, codec protocol.Codec) *websocketClient {
	clientCTX, cancel := context.WithCancel(ctx)

	return &websocketClient{
		id:         id,
		name:       name,
		sessionID:  "",
		conn:       conn,
		codec:      codec,
		responseCh: make(chan protocol.ResponseEnvelope, 100), // Buffered channel for responses
		hubCh:      nil,                                       // Channel to send commands to the hubz

		cancel:    cancel,
		clientCTX: clientCTX,
		wg:        sync.WaitGroup{},
		closeOnce: sync.Once{},
	}
}

type websocketClient struct {
	id         string
	name       string
	sessionID  string
	conn       protocol.RawConnection
	codec      protocol.Codec
	responseCh chan protocol.ResponseEnvelope
	hubCh      chan<- protocol.Command // Renamed from commandCh to hubCh for clarity
	serving    bool

	cancel    context.CancelFunc
	wg        sync.WaitGroup
	clientCTX context.Context
	closeOnce sync.Once // Ensures Close is called only once
}

func (c *websocketClient) ID() string {
	return c.id
}

func (c *websocketClient) Name() string {
	return c.name
}

func (c *websocketClient) SessionID() string {
	return c.sessionID
}

func (c *websocketClient) SetSessionID(sessionID string) error {
	c.sessionID = sessionID

	// todo: need to notify the client about the session ID change
	return nil
}

func (c *websocketClient) Send(ctx context.Context, envelope protocol.ResponseEnvelope) (err error) {
	defer func() {
		if r := recover(); r != nil {
			// Handle panic gracefully
			err = fmt.Errorf("%w: %v", ErrSendFailed, r)
		}
	}()

	if !c.serving {
		return ErrClientNotServing
	}

	select {
	case c.responseCh <- envelope:
		// Successfully sent response to channel
		return nil
	case <-ctx.Done():
		return fmt.Errorf("Send failed: %w", ctx.Err())
	case <-c.clientCTX.Done():
		// Client is being closed, don't attempt to send
		return ErrClientConnectionClosed
	}
}

// SetHubChannel sets the channel for sending commands to the hub.
func (c *websocketClient) SetHubChannel(hubCh chan<- protocol.Command) {
	c.hubCh = hubCh
}

func (c *websocketClient) Close(ctx context.Context) error {
	c.closeOnce.Do(func() {
		log.Println("Closing websocket client:", c.id)
		c.cancel()
	})

	return nil
}

func (c *websocketClient) Pump() (pumpErr error) {
	defer func() {
		c.wg.Wait()

		if err := c.conn.Close(); err != nil {
			if pumpErr != nil {
				pumpErr = errors.Join(pumpErr, fmt.Errorf("close connection error: %w", err))
			} else {
				pumpErr = fmt.Errorf("close connection error: %w", err)
			}
		}
	}()
	defer close(c.responseCh)

	errCh := make(chan error, 2) // Buffered channel for errors

	c.wg.Add(2)
	go c.readPump(c.clientCTX, errCh)
	go c.writePump(c.clientCTX, errCh)

	c.serving = true
	defer func() {
		c.serving = false
	}()

	select {
	case err := <-errCh:
		pumpErr = err
	case <-c.clientCTX.Done():
		pumpErr = c.clientCTX.Err()
	}

	return
}

func (c *websocketClient) readPump(ctx context.Context, errCh chan error) {
	defer c.wg.Done()

	for {
		var msg protocol.PlayerInput
		data, err := c.conn.ReadMessage()
		if err != nil {
			errCh <- fmt.Errorf("readPump error: %w", err)
			return
		}

		if errDecode := c.codec.Decode(data, &msg); errDecode != nil {
			errCh <- fmt.Errorf("readPump decoding error: %w", errDecode)
			return
		}

		command := protocol.Command{
			ClientID: c.id,
			Input:    msg,
		}

		select {
		case c.hubCh <- command: // Send command to the hub's channel
			// Successfully sent command to channel
		case <-ctx.Done():
			return
		}
	}
}

func (c *websocketClient) writePump(ctx context.Context, errCh chan error) {
	defer c.wg.Done()

	for {
		select {
		case resp := <-c.responseCh:
			data, err := c.codec.Encode(resp)
			if err != nil {
				errCh <- fmt.Errorf("writePump encoding error: %w", err)
				return
			}

			if errWrite := c.conn.WriteMessage(data); errWrite != nil {
				errCh <- fmt.Errorf("writePump write error: %w", errWrite)
				return
			}
		case <-ctx.Done():
			return
		}
	}
}
