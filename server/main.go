package main

import (
	"log"
	"math"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Player struct {
	ID          string  `json:"id"`
	X           float64 `json:"x"`
	Y           float64 `json:"y"`
	Angle       float64 `json:"angle"`
	TargetAngle float64 `json:"-"` // Not sent to client
	Health      int     `json:"health"`
	IsAlive     bool    `json:"isAlive"`
	conn        *websocket.Conn
	mutex       sync.Mutex
}

type GameState struct {
	Players map[string]*Player `json:"players"`
	mutex   sync.RWMutex
}

type Message struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}


var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var gameState = &GameState{
	Players: make(map[string]*Player),
}

const (
	PlayerSize     = 20
	MoveSpeed      = 120 // pixels per second (reduced from 200)
	RotationSpeed  = 4.0 // radians per second (smooth rotation)
	MapWidth       = 800
	MapHeight      = 600
)

func main() {
	http.HandleFunc("/ws", handleWebSocket)
	http.Handle("/", http.FileServer(http.Dir("../client/")))
	
	// Start game loop
	go gameLoop()
	
	log.Println("Server starting on :8030")
	log.Fatal(http.ListenAndServe(":8030", nil))
}

func gameLoop() {
	ticker := time.NewTicker(16 * time.Millisecond) // ~60 FPS
	defer ticker.Stop()

	for range ticker.C {
		updateGame()
		broadcastGameState()
	}
}

func updateGame() {
	gameState.mutex.Lock()
	defer gameState.mutex.Unlock()

	deltaTime := 1.0 / 60.0 // 60 FPS

	// Update player rotations and positions
	for _, player := range gameState.Players {
		if !player.IsAlive {
			continue
		}
		
		player.mutex.Lock()
		
		// Smooth rotation towards target angle
		if player.TargetAngle != 0 || player.Angle != 0 {
			angleDiff := player.TargetAngle - player.Angle
			
			// Normalize angle difference to [-π, π]
			for angleDiff > math.Pi {
				angleDiff -= 2 * math.Pi
			}
			for angleDiff < -math.Pi {
				angleDiff += 2 * math.Pi
			}
			
			// Apply smooth interpolation
			maxRotation := RotationSpeed * deltaTime
			if math.Abs(angleDiff) <= maxRotation {
				player.Angle = player.TargetAngle
			} else {
				if angleDiff > 0 {
					player.Angle += maxRotation
				} else {
					player.Angle -= maxRotation
				}
			}
			
			// Normalize angle to [-π, π]
			for player.Angle > math.Pi {
				player.Angle -= 2 * math.Pi
			}
			for player.Angle < -math.Pi {
				player.Angle += 2 * math.Pi
			}
		}
		
		player.mutex.Unlock()
	}
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}
	defer conn.Close()

	playerID := generatePlayerID()
	player := &Player{
		ID:      playerID,
		X:       float64(rand.Intn(MapWidth-PlayerSize) + PlayerSize/2),
		Y:       float64(rand.Intn(MapHeight-PlayerSize) + PlayerSize/2),
		Angle:   0,
		Health:  1,
		IsAlive: true,
		conn:    conn,
	}

	gameState.mutex.Lock()
	gameState.Players[playerID] = player
	gameState.mutex.Unlock()

	// Send initial game state
	msg := Message{
		Type: "init",
		Data: map[string]interface{}{
			"playerID": playerID,
			"mapWidth": MapWidth,
			"mapHeight": MapHeight,
		},
	}
	conn.WriteJSON(msg)

	defer func() {
		gameState.mutex.Lock()
		delete(gameState.Players, playerID)
		gameState.mutex.Unlock()
	}()

	for {
		var msg Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Println("Read error:", err)
			break
		}

		handleMessage(player, msg)
	}
}

func handleMessage(player *Player, msg Message) {
	switch msg.Type {
	case "input":
		data := msg.Data.(map[string]interface{})
		handlePlayerInput(player, data)
	case "attack":
		handlePlayerAttack(player, msg.Data)
	}
}

func handlePlayerInput(player *Player, data map[string]interface{}) {
	keys := data["keys"].(map[string]interface{})
	mouse := data["mouse"].(map[string]interface{})
	
	player.mutex.Lock()
	defer player.mutex.Unlock()

	// Update target angle based on mouse (server will smooth the transition)
	player.TargetAngle = mouse["angle"].(float64)

	// Update player position based on keys
	deltaTime := 1.0 / 60.0 // Assuming 60 FPS
	speed := MoveSpeed * deltaTime

	newX, newY := player.X, player.Y

	if keys["w"].(bool) {
		newY -= speed
	}
	if keys["s"].(bool) {
		newY += speed
	}
	if keys["a"].(bool) {
		newX -= speed
	}
	if keys["d"].(bool) {
		newX += speed
	}

	// Basic boundary checking
	if newX >= PlayerSize/2 && newX <= MapWidth-PlayerSize/2 {
		player.X = newX
	}
	if newY >= PlayerSize/2 && newY <= MapHeight-PlayerSize/2 {
		player.Y = newY
	}
}

func handlePlayerAttack(player *Player, data interface{}) {
	if !player.IsAlive {
		return
	}

	// Attack logic will be implemented here
	log.Printf("Player %s attacked", player.ID)
}

func broadcastGameState() {
	gameState.mutex.RLock()
	defer gameState.mutex.RUnlock()

	msg := Message{
		Type: "gameState",
		Data: gameState,
	}

	for _, player := range gameState.Players {
		if player.conn != nil {
			err := player.conn.WriteJSON(msg)
			if err != nil {
				log.Println("Write error:", err)
			}
		}
	}
}

func generatePlayerID() string {
	return "player_" + randomString(8)
}

func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}