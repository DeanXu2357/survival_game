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

### Wails Development
```bash
# Start development with hot reload
wails dev

# Build desktop application
wails build

# Check development environment
wails doctor
```

### Backend (Go)
```bash
cd backend
go mod tidy          # Install dependencies
go run main.go       # Run backend standalone
go test ./...        # Run tests
```

### Frontend (TypeScript + PixiJS)
```bash
cd frontend
npm install          # Install dependencies
npm run dev          # Development server
npm run build        # Build for production
npm run test         # Run tests
```

## Project Architecture

### Technology Stack
- **Backend**: Go with WebSocket for game logic and networking
- **Frontend**: TypeScript + PixiJS for high-performance 2D rendering
- **Desktop Framework**: Wails v2 for cross-platform desktop application
- **Communication**: WebSocket for real-time multiplayer

### Project Structure
```
survival/
├── frontend/        # TypeScript + PixiJS frontend
│   ├── src/
│   │   ├── game/    # Game logic and rendering
│   │   ├── ui/      # User interface components
│   │   └── types/   # TypeScript type definitions
│   ├── assets/      # Game assets (sprites, sounds)
│   └── package.json
├── backend/         # Go backend with game logic
│   ├── internal/    # Internal packages
│   │   ├── game/    # Game state and logic
│   │   ├── network/ # WebSocket handling
│   │   └── types/   # Go type definitions
│   ├── pkg/         # Public packages
│   └── main.go
├── shared/          # Shared types and constants
├── build/           # Wails build configuration
├── wails.json       # Wails configuration
└── app.go           # Wails application entry point
```

### Core Architecture
- **Server Authoritative**: All game logic, player positions, and combat calculations in Go backend
- **Client Rendering**: TypeScript frontend handles input, PixiJS rendering, and UI
- **Real-time Sync**: 60 FPS game loop on server broadcasts state to all clients
- **Desktop Integration**: Wails provides native desktop features and packaging

### Key Components

#### Backend (`backend/`)
- `GameState`: Global game state with all players, projectiles, and events
- `GameLogic`: Core game mechanics and update loop
- `NetworkManager`: WebSocket communication handling
- `Player`: Player state (position, angle, health, connection)

#### Frontend (`frontend/src/`)
- `GameRenderer`: PixiJS-based rendering system
- `InputManager`: WASD movement and mouse input handling
- `NetworkClient`: WebSocket client for server communication
- `UIManager`: Game menus and interface components

#### Wails Integration (`app.go`)
- Application lifecycle management
- Frontend-backend communication bridge
- Native desktop features (file dialogs, system notifications)

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
```typescript
interface PlayerInput {
  isPressingW: boolean;
  isPressingA: boolean;
  isPressingS: boolean;
  isPressingD: boolean;
  isShooting: boolean;
  mousePosition: { x: number; y: number };
}
```

#### Server → Client
```go
type GameState struct {
    Players     map[string]*Player
    Projectiles []*Projectile
    Walls       []*Wall
    SoundEvents []*SoundEvent
    Timestamp   int64
}
```

### Map System
- 800x600 pixel game area
- Building floor plan with walls and doors
- Triangle-shaped players rendered with PixiJS sprites
- Collision detection and pathfinding

## Current Development Status

### Completed Features (Legacy Canvas Implementation)
- ✅ Basic Go backend with WebSocket server
- ✅ HTML5 Canvas frontend with game loop (to be migrated)
- ✅ Basic player movement (WASD input)
- ✅ Mouse aiming system
- ✅ Real-time multiplayer synchronization
- ✅ Vision system: Fog of war with circular + cone visibility
- ✅ Main menu system and game state management
- ✅ Game mode selection (Solo, Multiplayer, Practice)

### Migration to New Tech Stack
- 🚧 **Wails Integration**: Convert to desktop application
- 🚧 **TypeScript Migration**: Rewrite frontend in TypeScript
- 🚧 **PixiJS Rendering**: High-performance 2D graphics
- 🚧 **Modern UI Framework**: Component-based UI system

### Next Development Priorities
1. **Wails Setup**: Initialize Wails project structure
2. **TypeScript Frontend**: Migrate from vanilla JS to TypeScript
3. **PixiJS Integration**: Replace Canvas with PixiJS rendering
4. **Sound Event System**: Visual audio cue implementation
5. **Map System**: Obstacles, walls, and collision detection
6. **Combat System**: Melee and ranged weapon mechanics
7. **AI System**: Enemy spawning and behavior

### Performance Targets
- 60 FPS rendering with PixiJS
- Sub-50ms network latency for multiplayer
- Smooth client-side prediction and reconciliation
- Efficient fog of war rendering with shaders

### Technical Notes
- Use Wails context for frontend-backend communication
- Implement TypeScript interfaces matching Go structs
- Leverage PixiJS texture atlases for sprite optimization
- Consider WebAssembly for shared game logic between frontend/backend
