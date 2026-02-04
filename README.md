# Survival

A 2.5D terminal-based tactical survival shooter featuring a Doom-style raycasting renderer. Navigate through environments using a first-person perspective rendered directly in your terminal.

## Core Features

- **2.5D Raycasting Renderer**: Doom-style first-person view using Unicode half-blocks for doubled vertical resolution
- **Terminal-Based**: Runs entirely in the terminal with ANSI color support
- **Network Architecture**: Host-authoritative multiplayer with WebSocket real-time synchronization
- **Cobra CLI**: Simple commands to run server or client

## Technology Stack

- **Language**: Go (pure Go, no external frontend)
- **Rendering**: Terminal-based 2.5D raycasting using Unicode half-blocks (`▀`)
- **Networking**: WebSocket (Gorilla WebSocket) for real-time multiplayer
- **CLI Framework**: Cobra for command-line interface
- **Architecture**: Server-authoritative with ECS-style game engine

## Project Structure

```
survival/
├── main.go              # CLI entry point
├── go.mod / go.sum      # Go module dependencies
├── cmd/                 # Cobra commands
│   ├── root.go          # Root command
│   ├── backend.go       # WebSocket server command
│   └── term.go          # Terminal client command
├── internal/
│   ├── adapters/
│   │   ├── handler/websocket/    # WebSocket server (Gorilla)
│   │   └── repository/maploader/ # JSON map loader
│   ├── engine/
│   │   ├── game.go       # Game logic
│   │   ├── map.go        # Map loading
│   │   ├── ports/        # Interfaces
│   │   ├── state/        # ECS components (entities, grid, world)
│   │   ├── system/       # Game systems (movement)
│   │   ├── vector/       # Vector math
│   │   └── weapons/      # Weapon types (knife, pistol)
│   ├── services/         # Hub, Room, Client, Session management
│   ├── terminal/         # Terminal frontend
│   │   ├── manager.go    # Game manager
│   │   ├── raycast/      # 2.5D raycasting renderer
│   │   ├── state/        # UI states (menu, settings, game)
│   │   ├── network/      # WebSocket client
│   │   └── ui/           # HUD overlays
│   └── utils/            # ID generator, JSON codec, time utilities
├── maps/                 # JSON map files
│   ├── office_floor_01.json
│   ├── test_simple.json
│   └── test_complex.json
└── docs/                 # Documentation
```

## Quick Start

### Prerequisites
- Go 1.21+ installed

### Running the Game

1. **Start the WebSocket Server**
   ```bash
   go run main.go backend -p 3033
   ```

2. **Start the Terminal Client** (in another terminal)
   ```bash
   go run main.go term
   ```

### Build and Run

```bash
# Build the binary
go build -o survival

# Run server
./survival backend -p 3033

# Run terminal client
./survival term
```

### Run Tests

```bash
go test ./...
```

## Game Controls

- **W/S** - Move forward/backward
- **A/D** - Strafe left/right
- **Q/E** - Turn left/right
- **ESC** - Exit/Back to menu
- **Enter** - Select menu option

## Architecture Overview

### Game Engine (ECS-style)
- **Entities**: Players, walls, projectiles with component-based design
- **Systems**: Movement, collision detection, input processing
- **State**: World state with spatial grid for efficient collision queries

### Terminal Renderer
- **2.5D Raycasting**: Casts rays for each screen column to determine wall distances
- **Half-Block Rendering**: Uses Unicode `▀` character to achieve 2x vertical resolution
- **Distance-Based Shading**: Walls darken with distance using ANSI 256-color palette

### Network Model
- **Server-Authoritative**: All game logic runs on the server
- **WebSocket**: Real-time bidirectional communication
- **JSON Protocol**: Simple message format for game state synchronization

## Development Commands

```bash
# Run WebSocket server (default port 3033)
go run main.go backend
go run main.go backend -p 8080  # Custom port

# Run terminal client
go run main.go term

# Run all tests
go test ./...

# Build binary
go build -o survival
```

## Map Format

Maps are defined in JSON format in the `maps/` directory:

```json
{
  "name": "office_floor_01",
  "width": 800,
  "height": 600,
  "walls": [
    {"x": 0, "y": 0, "width": 800, "height": 10},
    ...
  ],
  "spawn_points": [
    {"x": 400, "y": 300}
  ]
}
```

## Contributing

See `CLAUDE.md` for development guidelines and coding standards.
