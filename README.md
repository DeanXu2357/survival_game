# Tactical Survival Shooter

A top-down 2D tactical survival shooter featuring innovative vision control and sound recognition mechanics. Players navigate through fog of war environments using limited vision and visual sound cues to complete the tactical cycle of Search → Contact → Kill → Survive.

## Core Features

- **Vision System**: Circular close-range vision + directional cone vision with fog of war
- **Sound System**: Three-layer visual sound rings with eight-directional audio cues
- **Network Architecture**: Host-authoritative multiplayer with real-time synchronization
- **Game Modes**: PvP (3 lives) and PvE (one-hit kill with bullet time)
- **Desktop Application**: Cross-platform desktop app built with Wails

## Technology Stack

- **Backend**: Go with WebSocket for game logic and networking
- **Frontend**: TypeScript + PixiJS for high-performance 2D rendering
- **Desktop Framework**: Wails for cross-platform desktop application
- **Communication**: WebSocket for real-time bidirectional data
- **Architecture**: Server-authoritative with client-side prediction

## Project Structure

```
survival/
├── frontend/        # TypeScript + PixiJS frontend
│   ├── src/
│   ├── assets/
│   └── package.json
├── backend/         # Go backend with game logic
│   ├── internal/
│   ├── pkg/
│   └── main.go
├── shared/          # Shared types and constants
├── build/           # Wails build configuration
├── spec.md          # Detailed game specification
├── CLAUDE.md        # Development instructions
├── wails.json       # Wails configuration
└── README.md        # This file
```

## Quick Start

### Prerequisites
- Go 1.19+ installed
- Node.js 16+ and npm/yarn
- Wails v2 installed: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`

### Development Setup

1. **Install Dependencies**
   ```bash
   # Install Go dependencies
   go mod tidy
   
   # Install frontend dependencies
   cd frontend
   npm install
   cd ..
   ```

2. **Development Mode**
   ```bash
   # Start development server with hot reload
   wails dev
   ```

3. **Build Desktop Application**
   ```bash
   # Build for current platform
   wails build
   
   # Build for specific platforms
   wails build -platform windows/amd64
   wails build -platform darwin/amd64
   wails build -platform linux/amd64
   ```

### Game Controls
- **Movement**: WASD keys (120 pixels/second)
- **Aiming**: Mouse cursor controls player rotation (4.0 radians/second)
- **Combat**: Left click for attacks
- **UI Navigation**: Mouse for menu interactions

## Architecture Overview

### Three-Layer Game Architecture
1. **Game State Layer**: Stores complete world snapshot (players, projectiles, walls, sound events)
2. **Game Logic Layer**: Updates game state based on input and rules at 60 FPS
3. **Rendering Layer**: PixiJS handles efficient 2D graphics rendering

### Network Model
- **Host-Authoritative**: Host player acts as authoritative server
- **Real-Time Sync**: WebSocket communication for multiplayer
- **Client Prediction**: Local actions feel immediate with server reconciliation

### Wails Integration
- **Go Backend**: Handles game logic, AI, and network communication
- **TypeScript Frontend**: PixiJS rendering and user interface
- **Native Desktop**: Cross-platform deployment without browser dependencies

## Development Status

Current development progress and task tracking is maintained in `todo.md`. The project includes a complete WebSocket server infrastructure, game logic foundation, and PixiJS frontend client ready for game state integration.

## Game Modes

### PvP Mode
- Multiplayer combat with 3 lives per player
- Emphasis on tactical encounters and positioning
- Real-time competition between human players

### PvE Mode (Planned)
- Single-player challenge mode
- One-hit kill system with "bullet time" planning
- AI enemies with varying difficulty levels

## Development commands

### Wails Development
```bash
wails dev            # Start development with hot reload
wails build          # Build production desktop app
wails doctor         # Check development environment
```

### Backend Development
```bash
cd backend
go mod tidy          # Install Go dependencies
go test ./...        # Run backend tests
```

### Frontend Development
```bash
cd frontend
npm install          # Install TypeScript/PixiJS dependencies
npm run dev          # Start frontend development server
npm run build        # Build frontend for production
npm run test         # Run frontend tests
```

## Network Optimization Roadmap

1. **Phase 1 (Current)**: JSON full state broadcasting for development
2. **Phase 2**: Delta updates and event-driven messaging
3. **Phase 3**: Binary protocol (Protobuf) for production optimization

## PixiJS Rendering Features

- **High Performance**: Hardware-accelerated 2D rendering
- **Sprite Management**: Efficient texture atlas and sprite batching
- **Visual Effects**: Shaders for fog of war and lighting effects
- **UI Framework**: Rich UI components for game interface
- **Animation System**: Smooth interpolation and particle effects

## Contributing

This project follows a tactical game development approach with emphasis on:
- Clean architecture separation (State-Logic-Rendering)
- Server-authoritative networking for consistency
- Modern desktop application development with Wails
- High-performance rendering with PixiJS
- Innovative gameplay mechanics (vision + sound systems)

See `spec.md` for detailed technical specifications and `CLAUDE.md` for development guidelines.
