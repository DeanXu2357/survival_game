# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Requirements

### Language Requirements
- **Documentation**: Use American English for all documentation, comments, and user-facing text
- **Code**: Use English for all variable names, function names, and code comments
- **Avoid**: Do not use Chinese (Traditional or Simplified) in any project files

### Technology Stack Requirements
- **Backend**: Go with WebSocket for game logic and networking
- **Frontend**: TypeScript + PixiJS for high-performance 2D rendering  
- **Desktop Framework**: Wails v2 for cross-platform desktop application packaging
- **Communication**: WebSocket for real-time multiplayer synchronization

When implementing any features, ensure adherence to these technology choices.

## Development commands

### Development Phase (Current)
**Backend (Go WebSocket Server)**:
```bash
cd backend
go mod tidy          # Install dependencies
go run main.go       # Run WebSocket server on port 3033
go test ./...        # Run tests
```

**Frontend (TypeScript + PixiJS Web)**:
```bash
cd frontend
npm install          # Install dependencies
npm run dev          # Development server (usually port 5173)
npm run build        # Build for production
npm run test         # Run tests
```

**Current Status**:
```bash
# main.go is essentially empty - no server to run yet
# Frontend only has basic Vite template

# You can run the frontend template:
cd frontend && npm run dev
# But it's just a "Hello Vite + TypeScript" page

# No backend server exists yet to connect to
```

### Production Phase (Future)
**Wails Desktop Application**:
```bash
# Initialize Wails project (when migrating)
wails init -n survival -t vanilla-ts

# Start development with hot reload
wails dev

# Build desktop application
wails build

# Check development environment
wails doctor
```

## Project Architecture

### Technology Stack

#### Development Phase (Current)
- **Backend**: Go with WebSocket server for game logic and networking
- **Frontend**: TypeScript + PixiJS web application for high-performance 2D rendering
- **Communication**: WebSocket for real-time multiplayer (JSON serialization)
- **Development**: Separate backend/frontend with independent deployment

#### Production Phase (Future)
- **Desktop Framework**: Wails v2 for cross-platform desktop application
- **Communication**: Wails context bridge (replaces WebSocket)
- **Serialization**: Protocol Buffers for optimized performance
- **Distribution**: Single executable for Windows, macOS, and Linux

### Project Structure

#### Current Development Structure
```
survival/
├── main.go              # Go main entry point
├── go.mod               # Go module dependencies
├── go.sum               # Go dependency checksums
├── internal/            # Internal Go packages
│   ├── game/            # Game state and logic
│   │   ├── state.go     # Vector2D, Projectile, State structs
│   │   ├── logic.go     # Game mechanics and update loop
│   │   ├── player.go    # Player entity and behaviors
│   │   ├── room.go      # Room/game session management
│   │   ├── object.go    # Game objects and collision
│   │   ├── weapons.go   # Weapon systems
│   │   └── vector_test.go # Unit tests for vector math
│   ├── hub/             # WebSocket hub management
│   └── server/          # WebSocket server implementation
├── frontend/            # TypeScript + PixiJS web app (basic setup)
│   ├── index.html       # Web app entry point
│   ├── package.json     # npm dependencies (TypeScript, Vite, PixiJS)
│   ├── tsconfig.json    # TypeScript configuration
│   ├── vite.config.ts   # Vite build configuration
│   └── src/             # TypeScript source (basic Vite template)
│       ├── main.ts      # Basic Vite + TypeScript template
│       ├── counter.ts   # Template counter functionality
│       └── style.css    # Basic styling
├── spec.md              # Game specification document
├── CLAUDE.md            # This file - Claude Code instructions
├── todo.md              # Development task tracking
└── README.md            # Project overview
```

#### Future Production Structure
```
survival/            # Wails v2 desktop application
├── app.go           # Wails application entry point
├── wails.json       # Wails configuration
├── build/           # Wails build configuration
├── backend/         # Go backend (integrated with Wails)
│   ├── internal/    # Same structure as development phase
│   └── serialization/ # Protobuf serialization
├── frontend/        # TypeScript frontend (embedded in Wails)
│   ├── src/         # Same structure as development phase
│   ├── wails/       # Wails-specific integration
│   └── protobuf/    # Generated TypeScript protobuf code
└── dist/            # Final built application
```

### Core Architecture
- **Server Authoritative**: All game logic, player positions, and combat calculations in Go backend
- **Client Rendering**: TypeScript frontend handles input, PixiJS rendering, and UI
- **Real-time Sync**: 60 FPS game loop on server broadcasts state to all clients
- **Development**: WebSocket communication with JSON for easy debugging
- **Production**: Wails context bridge with Protobuf for optimal performance

### Key Components

#### Backend (`internal/game/`)
- `State`: Game state structure with players, walls, projectiles (state.go)
- `Logic`: Basic collision detection and player movement (logic.go)
- `Player`: Player struct with movement logic (player.go)
- `Wall`: Wall collision system (object.go)
- `Room`: Game session framework (room.go) - no networking yet
- Various weapon interfaces (weapons.go) - no implementation yet

#### Frontend (`frontend/src/`) - To Be Implemented
- `GameRenderer`: PixiJS-based rendering system (planned)
- `InputManager`: WASD movement and mouse input handling (planned)
- `NetworkClient`: WebSocket client for server communication (planned)
- `UIManager`: Game menus and interface components (planned)

#### Wails Integration - Future Phase
- `app.go`: Application lifecycle management (not yet implemented)
- Frontend-backend communication bridge (planned for production phase)
- Native desktop features (planned for production phase)

### Game Mechanics

#### Vision System
- **Close Vision**: 1 player body-length radius (20px) around player
- **Cone Vision**: 45-degree cone extending 10 body-lengths (200px) forward
- **Fog of War**: PixiJS shaders for efficient fog rendering

#### Sound System (Planned)
- **Three-Layer Sound Rings**: Visual representation of audio events
- **Eight-Directional Audio**: Sound cues divided into 8 compass directions
- **Event Types**: Footsteps, weapon fire, environmental interactions

#### Player Movement
- Server processes WASD input at 60 FPS
- 120 pixels/second movement speed
- Boundary checking within map limits
- Server-side rotation smoothing (4.0 radians/second)

#### Combat System (To be implemented)
- Melee: 1 body-length range knife attack
- Ranged: Small pistol with projectile physics
- One-hit kill system initially, expandable to health system

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
Server broadcasts game state updates through WebSocket using JSON protocol. Connection uses request body with game name, client ID, and session information for reconnection support.

### Map System
- 800x600 pixel game area
- Building floor plan with walls and doors
- Triangle-shaped players rendered with PixiJS sprites
- Collision detection and pathfinding

## Current Development Status

### Current Implementation Status

#### What Actually Exists
- ✅ **Basic Go Project**: go.mod, go.sum, basic module setup
- ✅ **Data Structures**: Vector2D, State, Projectile structs (state.go)
- ✅ **Spatial Grid**: Grid system for collision detection (state.go)
- ✅ **Player Structure**: Player struct with basic movement logic (player.go)
- ✅ **Wall System**: Wall struct with collision detection (object.go)
- ✅ **Game Logic Framework**: Basic collision detection and player movement (logic.go)
- ✅ **Room Structure**: Room struct definition with game loop framework (room.go)
- ✅ **Weapon Interfaces**: Weapon type definitions and interfaces (weapons.go)
- ✅ **Frontend Setup**: Vite + TypeScript + PixiJS environment configured

#### What Doesn't Exist Yet
- ❌ **Main Application**: main.go is essentially empty
- ❌ **WebSocket Server**: No server implementation
- ❌ **Network Communication**: No client-server protocol
- ❌ **Frontend Game Code**: Only basic Vite template exists
- ❌ **Weapon Implementation**: Only interfaces/types, no actual weapons
- ❌ **Game Integration**: No way to actually run or play the game

#### Production Phase (Future)
- ⏸️ **Wails Integration**: Desktop application packaging
- ⏸️ **Protocol Buffers**: Optimized serialization
- ⏸️ **Desktop Polish**: Native features and distribution

### What Needs to Be Built Next
1. **WebSocket Server**: main.go needs actual server implementation
2. **Frontend Game Client**: Replace Vite template with actual game
3. **Network Protocol**: Connect frontend and backend
4. **Basic Gameplay**: Make the existing structures actually work together

### Later Development
- Vision system, combat system, UI system, etc.

### Performance Targets
- 60 FPS rendering with PixiJS
- Sub-50ms network latency for multiplayer
- Smooth client-side prediction and reconciliation
- Efficient fog of war rendering with shaders

### Technical Notes
- These are future considerations once basic functionality exists
- Current focus should be on getting a minimal working game first
