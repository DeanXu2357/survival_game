# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Requirements

### Language Requirements
- **Documentation**: Use American English for all documentation, comments, and user-facing text
- **Code**: Use English for all variable names, function names, and code comments
- **Avoid**: Do not use Chinese (Traditional or Simplified) in any project files

### Technology Stack Requirements
- **Language**: Go only (no TypeScript/JavaScript frontend)
- **Rendering**: Terminal-based 2.5D raycasting using Unicode half-blocks
- **Networking**: WebSocket (Gorilla WebSocket) for real-time multiplayer
- **CLI**: Cobra for command-line interface

When implementing any features, ensure adherence to these technology choices.

## Development Commands

```bash
# Run WebSocket server (default port 3033)
go run main.go backend
go run main.go backend -p 3033

# Run terminal client
go run main.go term

# Run tests
go test ./...

# Build binary
go build -o survival

# Run specific test package
go test ./internal/engine/...
go test ./internal/terminal/...
```

## Project Architecture

### Technology Stack
- **Backend**: Go with WebSocket server for game logic and networking
- **Frontend**: Terminal-based 2.5D raycasting renderer (pure Go)
- **Communication**: WebSocket for real-time multiplayer (JSON serialization)
- **CLI**: Cobra commands (`survival backend`, `survival term`)

### Project Structure

```
survival/
├── main.go              # CLI entry point (calls cmd.Execute())
├── go.mod / go.sum      # Go module dependencies
├── cmd/                 # Cobra commands
│   ├── root.go          # Root command definition
│   ├── backend.go       # WebSocket server command
│   └── term.go          # Terminal client command
├── internal/
│   ├── adapters/
│   │   ├── handler/websocket/    # WebSocket server implementation
│   │   │   ├── server.go         # HTTP server with WebSocket upgrade
│   │   │   └── connection.go     # Connection handling
│   │   └── repository/maploader/ # JSON map file loader
│   │       └── json_loader.go    # Loads maps from JSON files
│   ├── engine/
│   │   ├── game.go       # Game logic orchestration
│   │   ├── map.go        # Map loading and management
│   │   ├── ports/        # Interface definitions
│   │   │   ├── port.go       # Core interfaces
│   │   │   ├── protocol.go   # Network protocol types
│   │   │   └── const.go      # Game constants
│   │   ├── state/        # ECS-style game state
│   │   │   ├── component.go    # Component definitions
│   │   │   ├── entities.go     # Entity types
│   │   │   ├── entitymanager.go # Entity management
│   │   │   ├── grid.go         # Spatial grid for collisions
│   │   │   ├── world.go        # World state container
│   │   │   ├── layer.go        # Rendering layers
│   │   │   └── cmdbuffer.go    # Command buffering
│   │   ├── system/       # Game systems
│   │   │   └── movement_basic.go # Movement system
│   │   ├── vector/       # Vector math utilities
│   │   │   └── vector.go
│   │   └── weapons/      # Weapon implementations
│   │       ├── types.go      # Weapon interfaces
│   │       ├── knife.go      # Melee weapon
│   │       ├── pistol.go     # Ranged weapon
│   │       └── projectile.go # Projectile logic
│   ├── services/         # Game services
│   │   ├── hub.go        # WebSocket hub management
│   │   ├── room.go       # Game room/session
│   │   ├── client.go     # Client representation
│   │   └── session.go    # Session management
│   ├── terminal/         # Terminal frontend
│   │   ├── manager.go    # Game manager (main loop)
│   │   ├── construct.go  # Terminal construction
│   │   ├── locale.go     # Localization
│   │   ├── raycast/      # 2.5D rendering
│   │   │   ├── raycast.go      # Ray casting algorithm
│   │   │   └── renderer25d.go  # Half-block renderer
│   │   ├── state/        # UI state machine
│   │   │   ├── mainmenu.go     # Main menu state
│   │   │   ├── settings.go     # Settings state
│   │   │   ├── singleplayer.go # Single player game state
│   │   │   └── resize.go       # Terminal resize handling
│   │   ├── network/      # WebSocket client
│   │   │   └── client.go       # Network client
│   │   ├── ui/           # UI overlays
│   │   │   └── overlay.go      # HUD, crosshair, weapon display
│   │   ├── composite/    # Composite rendering
│   │   └── debug/        # Debug utilities
│   └── utils/            # Shared utilities
│       ├── id_generator.go   # Unique ID generation
│       ├── json_codec.go     # JSON encoding/decoding
│       └── time.go           # Time utilities
├── maps/                 # JSON map files
│   ├── office_floor_01.json  # Office floor map
│   ├── test_simple.json      # Simple test map
│   └── test_complex.json     # Complex test map
└── docs/                 # Documentation
```

### Core Architecture
- **Server Authoritative**: All game logic, player positions, and combat calculations in Go backend
- **Terminal Rendering**: 2.5D raycasting with Unicode half-blocks for first-person view
- **Real-time Sync**: 60 FPS game loop on server broadcasts state to all clients
- **ECS-Style Engine**: Components, entities, and systems for game state management

### Key Components

#### CLI Commands (`cmd/`)
- `survival backend -p PORT`: Start WebSocket game server
- `survival term`: Start terminal-based game client

#### Engine (`internal/engine/`)
- **state/**: ECS components (Transform, Collider, Health, etc.)
- **system/**: Game systems (movement, collision)
- **vector/**: Vector2D math operations
- **weapons/**: Knife, Pistol with projectile physics
- **ports/**: Interface definitions for dependency injection

#### Terminal Frontend (`internal/terminal/`)
- **raycast/**: 2.5D Doom-style rendering using ray casting
- **state/**: State machine for UI (menu, settings, game)
- **network/**: WebSocket client for multiplayer
- **ui/**: HUD overlays (health, ammo, crosshair)

#### Services (`internal/services/`)
- **Hub**: Manages all connected clients
- **Room**: Game session with players
- **Client**: Individual player connection
- **Session**: Reconnection support

### Game Mechanics

#### Terminal Rendering
- **Half-Block Technique**: Uses `▀` (U+2580) to render 2 pixels per character
- **FOV**: 90 degrees for natural perspective
- **Distance Shading**: Walls darken with distance (ANSI 256-color)

#### Player Movement
- Server processes WASD input at 60 FPS
- 120 pixels/second movement speed
- Boundary checking within map limits
- Server-side rotation smoothing

#### Combat System
- Melee: Knife attack
- Ranged: Pistol with projectile physics
- One-hit kill system (expandable to health system)

### Network Protocol

#### Client → Server
```go
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
```

#### Server → Client
Server broadcasts game state updates through WebSocket using JSON protocol.

### Map System
- JSON format map files in `maps/` directory
- Walls with position, dimensions, and optional height/elevation
- Spawn points for player placement
- Spatial grid for efficient collision detection

## Current Development Status

### What Exists
- ✅ **CLI Application**: Cobra commands (`backend`, `term`)
- ✅ **WebSocket Server**: Full server implementation with Gorilla WebSocket
- ✅ **Terminal Client**: 2.5D raycasting renderer with state machine
- ✅ **Game Engine**: ECS-style with entities, components, systems
- ✅ **Spatial Grid**: Efficient collision detection
- ✅ **Map System**: JSON map loader with multiple test maps
- ✅ **Weapon System**: Knife and pistol implementations
- ✅ **Network Protocol**: JSON-based client-server communication
- ✅ **UI System**: Main menu, settings, HUD overlays

### Areas for Future Development
- Vision/fog of war system
- Sound visualization system
- Additional weapons and items
- AI enemies for PvE mode
- More maps and level design
- Client-side prediction and reconciliation

### Performance Targets
- 60 FPS terminal rendering
- Sub-50ms network latency for multiplayer
- Efficient spatial queries for collision detection
