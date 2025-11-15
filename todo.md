# Survival Game Implementation Todo List

## Current Architecture Status

### ✅ Fully Implemented Features
- **Project Infrastructure**: Complete Go backend + TypeScript frontend setup
- **WebSocket Server**: Full server implementation with graceful shutdown (main.go:17, websocket/server.go)
- **Network Hub System**: Complete hub with client registry and room management (app/hub.go)
- **WebSocket Client Connections**: Full connection lifecycle with ReadPump/WritePump (websocket/connection.go)
- **Protocol System**: JSON codec with protocol.Command interface (protocol/json_codec.go)
- **Game Data Structures**: Vector2D, Player, State, Projectile, Wall structs (game/state.go)
- **Spatial Grid System**: Grid-based collision detection optimization (game/grid.go)
- **Map System**: JSON map loader with collision walls (infrastructure/map/json_loader.go)
- **Frontend Client**: PixiJS app with network connection and input handling (frontend/src/)
- **Client Registry**: Robust client management with ID generation (app/client_registry.go)
- **Player Registry**: Player lifecycle management (game/player_registry.go)

### ⚠️ Partially Implemented
- **Game Logic**: Basic player movement, needs combat integration (game/logic.go)
- **Room System**: Room structure exists, needs game loop integration (game/room.go)
- **Weapon System**: Interfaces defined, implementation missing (game/weapons.go)
- **Message Routing**: Protocol exists, needs game state broadcasting
- **Frontend Rendering**: PixiJS setup complete, needs game state visualization

## Phase 1: Core Game Functionality (High Priority)

### 1.1 Combat System Integration
- [ ] **Weapon Implementation** - Implement Knife and Pistol weapon logic
- [ ] **Shooting Mechanics** - Fire projectiles with proper range and collision
- [ ] **Reload System** - Normal (3s) and fast (1s) reload with magazine management
- [ ] **Ammo Management** - Magazine consumption and reload validation

### 1.2 Game State Broadcasting
- [ ] **Server Message Types** - Define GameState, PlayerUpdate, ProjectileUpdate messages
- [ ] **State Serialization** - Broadcast complete game state to all clients
- [ ] **Delta Updates** - Optimize network traffic with incremental updates
- [ ] **Client State Sync** - Frontend receives and applies server state updates

### 1.3 Game Loop Integration
- [ ] **Room Game Loop** - Integrate 60 FPS game logic with networking
- [ ] **Input Processing** - Process client input in game loop
- [ ] **Collision Detection** - Player-wall, projectile-wall, projectile-player collisions
- [ ] **Physics Updates** - Movement, projectile trajectories, and cleanup

## Phase 2: Advanced Features (Medium Priority)

### 2.1 Game Rules & Player Lifecycle
- [ ] **Health System** - One-hit kill damage system
- [ ] **Player Death** - Handle player elimination and respawn
- [ ] **Win Conditions** - Last player standing logic
- [ ] **Game Session Management** - Start, pause, reset game states

### 2.2 Vision & Rendering
- [ ] **Fog of War** - PixiJS shader-based fog rendering
- [ ] **Vision System** - Close vision (20px) + cone vision (200px, 45°)
- [ ] **Player Visibility** - Hide players outside vision range
- [ ] **Sound Visualization** - Three-layer visual sound rings

### 2.3 Frontend Game Visualization
- [ ] **Player Rendering** - Triangle sprites with rotation
- [ ] **Map Rendering** - Wall visualization from loaded map data
- [ ] **Projectile Visualization** - Bullet trails and impacts
- [ ] **UI System** - Health, ammo, and game status display

## Phase 3: Polish & Optimization (Low Priority)

### 3.1 Performance Optimization
- [ ] **Delta Updates** - Optimize network traffic with incremental state updates
- [ ] **Spatial Optimization** - Leverage existing Grid system for collision queries
- [ ] **Rendering Optimization** - PixiJS sprite batching and culling
- [ ] **Memory Management** - Projectile cleanup and object pooling

### 3.2 Error Handling & Robustness
- [ ] **Connection Recovery** - Handle client disconnection and reconnection
- [ ] **Invalid Input Handling** - Validate and sanitize client input
- [ ] **Room Cleanup** - Remove empty rooms and clean up resources
- [ ] **Multi-room Support** - Support multiple concurrent game sessions

### 3.3 Testing & Validation
- [ ] **Unit Tests** - Movement, collision, weapon systems
- [ ] **Integration Tests** - Full client-server game scenarios
- [ ] **Performance Tests** - 60 FPS validation with multiple clients
- [ ] **Network Tests** - Latency and packet loss handling

## Implementation Notes

### Current Architecture Summary
The project has evolved significantly from initial design. Key architectural components:

```
internal/
├── app/                     # Application layer
│   ├── hub.go              # Central message hub with room management
│   ├── client_registry.go  # Client connection management
│   └── client.go           # Individual client handling
├── game/                    # Game logic layer
│   ├── state.go            # Core game data structures
│   ├── logic.go            # Game mechanics and physics
│   ├── player_registry.go  # Player lifecycle management
│   ├── room.go             # Game session management
│   └── weapons.go          # Weapon system interfaces
├── infrastructure/         # Infrastructure layer
│   ├── network/websocket/  # WebSocket server implementation
│   └── map/               # Map loading and management
└── protocol/              # Communication protocol
    ├── protocol.go        # Message interfaces
    └── json_codec.go      # JSON serialization
```

### Next Implementation Steps
1. **Weapon Logic Implementation**: Complete the weapon interfaces in weapons.go
2. **Game State Broadcasting**: Implement server→client state updates
3. **Game Loop Integration**: Connect room game logic with network hub
4. **Frontend State Rendering**: Visualize server state in PixiJS client

### Technical Priorities
- **Server-Authoritative**: All game logic runs on Go backend
- **Real-time Network**: 60 FPS game loop with WebSocket synchronization  
- **Robust Architecture**: Separation of concerns between app, game, and infrastructure layers
- **JSON Protocol**: Human-readable development protocol (Protobuf for production later)
