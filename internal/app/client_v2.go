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

type CommandHandler func(cmd protocol.Command)

type clientSubscription struct {
	id      string
	handler CommandHandler
	manager *subscriptionManager
	once    sync.Once
}

// Unsubscribe terminates the subscription.
func (s *clientSubscription) Unsubscribe() error {
	s.manager.Remove(s.id)
	return nil
}

func (s *clientSubscription) DeliveryChannel(source <-chan protocol.Command) {
	s.once.Do(func() {
		go func() {
			for cmd := range source {
				if s.handler != nil {
					s.handler(cmd) // Call the handler with the command
				}
			}
		}()
	})
}

type subscriptionManager struct {
	mu            sync.RWMutex
	subscriptions map[string]*clientSubscription
	channels      map[string]chan protocol.Command
}

var ErrClientSubscriptionExists = errors.New("client subscription already exists")

func (sm *subscriptionManager) Add(subscription *clientSubscription) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if subscription == nil {
		return nil
	}

	if _, exists := sm.subscriptions[subscription.id]; exists {
		return ErrClientSubscriptionExists
	}

	sm.subscriptions[subscription.id] = subscription

	subscription.manager = sm

	var source chan protocol.Command
	subscription.DeliveryChannel(source)
	sm.channels[subscription.id] = make(chan protocol.Command, 100) // Buffered channel for commands

	return nil
}

func (sm *subscriptionManager) Remove(subscriptionID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if _, exists := sm.subscriptions[subscriptionID]; exists {
		delete(sm.subscriptions, subscriptionID)
	}
	if ch, exists := sm.channels[subscriptionID]; exists {
		delete(sm.channels, subscriptionID)
		close(ch) // Close the channel to stop receiving commands
	}

	return
}

func (sm *subscriptionManager) RemoveAll() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	for id := range sm.subscriptions {
		delete(sm.subscriptions, id)
		if ch, exists := sm.channels[id]; exists {
			close(ch) // Close the channel to stop receiving commands
			delete(sm.channels, id)
		}
	}
}

func (sm *subscriptionManager) All() []*clientSubscription {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	subscriptions := make([]*clientSubscription, 0, len(sm.subscriptions))
	for _, sub := range sm.subscriptions {
		subscriptions = append(subscriptions, sub)
	}
	return subscriptions
}

func (sm *subscriptionManager) AllChannels() []chan<- protocol.Command {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	channels := make([]chan<- protocol.Command, 0, len(sm.channels))
	for _, ch := range sm.channels {
		channels = append(channels, ch)
	}

	return channels
}

func newSubscriptionManager() *subscriptionManager {
	return &subscriptionManager{
		subscriptions: make(map[string]*clientSubscription),
		mu:            sync.RWMutex{},
	}
}

func newWebsocketClientV2(ctx context.Context, id, name string, conn protocol.RawConnection, codec protocol.Codec) *websocketClientV2 {
	subIDGen := utils.NewSequentialIDGenerator(fmt.Sprintf("c%s-sub-", id))

	clientCTX, cancel := context.WithCancel(ctx)

	client := &websocketClientV2{
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

type websocketClientV2 struct {
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

func (c *websocketClientV2) ID() string {
	return c.id
}

func (c *websocketClientV2) Name() string {
	return c.name
}

func (c *websocketClientV2) SessionID() string {
	return c.sessionID
}

func (c *websocketClientV2) SetSessionID(sessionID string) error {
	c.sessionID = sessionID

	// todo: need to notify the client about the session ID change
	return nil
}

func (c *websocketClientV2) Errors() <-chan error {
	return c.errCh
}

func (c *websocketClientV2) Subscribe(handler CommandHandler) (*clientSubscription, error) {
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

func (c *websocketClientV2) Send(ctx context.Context, envelope protocol.ResponseEnvelope) (err error) {
	defer func() {
		if r := recover(); r != nil {
			// Handle panic gracefully
			err = fmt.Errorf("%w: %v", ErrSendFailed, r)
		}
	}()

	data, err := c.codec.Encode(envelope)
	if err != nil {
		return fmt.Errorf("writePump encoding error: %w", err)
	}

	if errWrite := c.conn.WriteMessage(data); errWrite != nil {
		return fmt.Errorf("writePump write error: %w", errWrite)
	}

	return nil
}

func (c *websocketClientV2) Close() (closeErr error) {
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

func (c *websocketClientV2) readPump() {
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

func (c *websocketClientV2) writePump() {
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
