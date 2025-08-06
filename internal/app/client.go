package app

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"

	"survival/internal/protocol"
	"survival/internal/utils"
)

var (
	ErrSendFailed               = errors.New("failed to send message to client")
	ErrClientConnectionClosed   = errors.New("client connection is closed")
	ErrClientNotServing         = errors.New("client is not serving requests") // TODO: rename this
	ErrClientSubscriptionExists = errors.New("client subscription already exists")
)

type Client interface {
	ID() string
	Name() string
	SessionID() string
	SetSessionID(sessionID string) error
	Send(ctx context.Context, envelope protocol.ResponseEnvelope) error
	Subscribe(handler CommandHandler) (*clientSubscription, error)
	Errors() <-chan error
	Close() error
}

type CommandHandler func(cmd protocol.Command)

func newWebsocketClient(ctx context.Context, id, name string, conn protocol.RawConnection, codec protocol.Codec) *websocketClient {
	subIDGen := utils.NewSequentialIDGenerator(fmt.Sprintf("c%s-sub-", id))

	clientCTX, cancel := context.WithCancel(ctx)

	client := &websocketClient{
		id:          id,
		name:        name,
		sessionID:   "",
		conn:        conn,
		codec:       codec,
		subManager:  newSubscriptionManager(),
		responsePub: make(chan protocol.ResponseEnvelope, 100), // Buffered channel for responses

		closeOnce: sync.Once{},

		idGen: subIDGen,

		cancel:    cancel,
		clientCTX: clientCTX,
		errCh:     make(chan error, 10),
	}

	go client.writePump()
	go client.readPump()

	return client
}

type websocketClient struct {
	id          string
	name        string
	sessionID   string
	conn        protocol.RawConnection
	codec       protocol.Codec
	responsePub chan protocol.ResponseEnvelope
	subManager  *subscriptionManager // Manages subscriptions for this client

	closeOnce sync.Once

	idGen IDGenerator

	cancel    context.CancelFunc
	clientCTX context.Context
	errCh     chan error
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

	payload := protocol.SystemSetSessionPayload{
		ClientID:  c.id,
		SessionID: sessionID,
	}

	encoded, err := c.codec.Encode(payload)
	if err != nil {
		return fmt.Errorf("failed to encode session ID payload: %w", err)
	}

	if errSend := c.Send(c.clientCTX, protocol.ResponseEnvelope{
		Type:    protocol.SystemSetSessionEnvelope,
		Payload: encoded,
	}); errSend != nil {
		return fmt.Errorf("failed to send session ID to client %s: %w", c.id, errSend)
	}

	return nil
}

func (c *websocketClient) Errors() <-chan error {
	return c.errCh
}

func (c *websocketClient) Subscribe(handler CommandHandler) (*clientSubscription, error) {
	if handler == nil {
		return nil, errors.New("handler cannot be nil")
	}

	subscription := &clientSubscription{
		id:      c.idGen.GenerateID(),
		handler: handler,
		manager: c.subManager,
		once:    sync.Once{},
	}

	if err := c.subManager.Add(subscription); err != nil {
		return nil, fmt.Errorf("failed to add subscription: %w", err)
	}

	go c.readPump()

	return subscription, nil
}

func (c *websocketClient) Send(ctx context.Context, envelope protocol.ResponseEnvelope) (err error) {
	defer func() {
		if r := recover(); r != nil {
			// Handle panic gracefully
			err = fmt.Errorf("%w: %v", ErrSendFailed, r)
		}
	}()

	select {
	case c.responsePub <- envelope:
		// Successfully sent response to channel
		return nil
	case <-ctx.Done():
		return fmt.Errorf("Send failed: %w", ctx.Err())
	case <-c.clientCTX.Done():
		// Client is being closed, don't attempt to send
		return ErrClientConnectionClosed
	default:
		return ErrClientNotServing
	}
}

func (c *websocketClient) Close() (closeErr error) {
	c.closeOnce.Do(func() {
		c.cancel()

		if err := c.conn.Close(); err != nil {
			if closeErr != nil {
				closeErr = errors.Join(closeErr, fmt.Errorf("close connection error: %w", err))
			} else {
				closeErr = fmt.Errorf("close connection error: %w", err)
			}
		}

		c.subManager.RemoveAll()
	})

	return
}

func (c *websocketClient) readPump() {
	for {
		var msg protocol.PlayerInput
		data, err := c.conn.ReadMessage()
		if err != nil {
			select {
			case c.errCh <- fmt.Errorf("readPump error: %w", err):
			default:
			}
			return
		}

		if errDecode := c.codec.Decode(data, &msg); errDecode != nil {
			select {
			case c.errCh <- fmt.Errorf("readPump decoding error: %w", errDecode):
			default:
			}
			return
		}

		command := protocol.Command{
			ClientID: c.id,
			Input:    msg,
		}

		for _, sub := range c.subManager.AllChannels() {
			select {
			case sub <- command: // Send command to the subscription channel
			default:
				// If the channel is full, skip sending to avoid blocking
				log.Printf("Warning: Subscription channel for client %s is full, skipping command delivery", c.id)
			}
		}

		select {
		case <-c.clientCTX.Done():
			return
		default:
		}
	}
}

func (c *websocketClient) writePump() {
	for {
		select {
		case resp := <-c.responsePub:
			data, err := c.codec.Encode(resp)
			if err != nil {
				select {
				case c.errCh <- fmt.Errorf("writePump encoding error: %w", err):
				default:
					log.Printf("Error encoding message for client %s: %v", c.id, err)
				}
				return
			}

			if errWrite := c.conn.WriteMessage(data); errWrite != nil {
				select {
				case c.errCh <- fmt.Errorf("writePump write error: %w", errWrite):
				default:
					// If the error channel is full, log the error instead of blocking
					log.Printf("Error sending message client client %s: %v", c.id, errWrite)
				}
				return
			}
		case <-c.clientCTX.Done():
			return
		}
	}
}
