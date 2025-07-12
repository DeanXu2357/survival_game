# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

### Backend (Go Server)
```bash
cd server
go run main.go
```
Server runs on port 8030

### Frontend
Visit http://localhost:8030 (served by Go server)

### Dependencies
```bash
cd server
go mod tidy
```

## Project Architecture

### Technology Stack
- **Backend**: Go with Gorilla WebSocket
- **Frontend**: HTML5 Canvas with vanilla JavaScript
- **Communication**: WebSocket for real-time multiplayer

### Core Architecture
- **Server Authoritative**: All game logic, player positions, and combat calculations happen on the Go server
- **Client Rendering**: Frontend only handles input, rendering, and UI
- **Real-time Sync**: 60 FPS game loop on server broadcasts state to all clients

### Key Components

#### Server (`server/main.go`)
- `Player`: Represents player state (position, angle, health, connection)
- `GameState`: Global game state with all players
- `gameLoop()`: 60 FPS update loop for physics and state updates
- WebSocket handlers for client communication

#### Client (`client/game.js`)
- `Game` class: Main game controller
- Vision system: Fog of war with circular + cone visibility
- Input handling: WASD movement, mouse aiming
- Rendering: Canvas-based 2D graphics

### Game Mechanics

#### Vision System
- **Close Vision**: 1 player body-length radius (20px) around player
- **Cone Vision**: 45-degree cone extending 10 body-lengths (200px) forward
- **Fog of War**: Everything outside vision is black

#### Player Movement
- Server processes WASD input at 60 FPS
- 120 pixels/second movement speed (reduced for better control)
- Boundary checking within map limits
- Server-side rotation smoothing (4.0 radians/second)

#### Combat (To be implemented)
- Melee: 1 body-length range knife attack
- Ranged: Small pistol
- One-hit kill system initially

### Message Protocol

#### Client → Server
- `input`: Player keyboard/mouse input
- `attack`: Attack command with target coordinates

#### Server → Client  
- `init`: Initial player ID and map dimensions
- `gameState`: Complete game state with all players

### Map System
- 800x600 pixel game area
- Building floor plan with walls and doors (to be implemented)
- Triangle-shaped players

## Current Development Status

### Completed Features
- ✅ Project structure setup (client/server/shared directories)
- ✅ Go backend with WebSocket server (port 8030)
- ✅ HTML5 Canvas frontend with game loop
- ✅ Basic player movement (WASD input)
- ✅ Mouse aiming system
- ✅ Real-time multiplayer synchronization
- ✅ Player spawn system with random positions
- ✅ **Vision System**: Fog of war with circular + cone visibility
- ✅ **Movement/Rotation Speed**: Balanced and smooth controls
- ✅ **Main Menu System**: Complete game flow (Menu → Game → Results → Menu)
- ✅ **Game State Management**: Multiple screens with proper transitions
- ✅ **Result Screen**: Statistics tracking and display
- ✅ **Game Modes**: Solo, Multiplayer, Practice mode selection

### Recent Fixes
- ✅ **Vision System Fixed**: Player triangles now visible with proper fog rendering
  - Issue resolved: Used `evenodd` fill rule for fog holes instead of `destination-out`
  - Background contrast improved (#666 vs black fog)
  - Rendering order corrected: players → fog → UI
- ✅ **Speed Balancing**: Movement (120px/s) and rotation (4.0 rad/s) now coordinated
- ✅ **Architecture Consistency**: Both movement and rotation controlled server-side

### Next Steps
1. Implement basic map obstacles and collision detection
2. Add melee combat system
3. Add enemy AI and spawning
4. Implement health/damage system
5. Add sound effects and better graphics

### Technical Notes
- Server runs on port 8030 with 60 FPS game loop
- Client-server communication via WebSocket
- Server-authoritative architecture for consistency
- Game automatically ends after 30 seconds (demo mode)