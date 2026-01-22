package websocket

import (
	"github.com/gorilla/websocket"

	"survival/internal/engine/ports"
)

// websocketConnection wraps gorilla/websocket.Conn to implement protocol.RawConnection
type websocketConnection struct {
	conn *websocket.Conn
}

// NewWebSocketConnection creates a new websocket connection wrapper
func NewWebSocketConnection(conn *websocket.Conn) ports.RawConnection {
	return &websocketConnection{conn: conn}
}

func (wc *websocketConnection) ReadMessage() ([]byte, error) {
	_, data, err := wc.conn.ReadMessage()
	return data, err
}

func (wc *websocketConnection) WriteMessage(data []byte) error {
	return wc.conn.WriteMessage(websocket.TextMessage, data)
}

func (wc *websocketConnection) Close() error {
	return wc.conn.Close()
}
