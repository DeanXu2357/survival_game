# Survival Game Specification

## Game Overview

A top-down 2D tactical survival shooter game centered on **vision control** and **sound recognition**. Players navigate through limited-visibility "fog of war" environments, completing the core cycle of "Search → Contact → Kill → Survive". The game features innovative visual sound cues and tactical gameplay that rewards strategic thinking over fast reflexes.

## Core Game Concepts

### Game Modes
- **PvP Mode**: Multiplayer combat with 3 lives per player, emphasizing unknown tactical encounters
- **PvE Mode**: Single-player challenge with one-hit kills and "bullet time" planning system

### Core Gameplay Loop
Players must master the tactical cycle:
1. **Search**: Use limited vision and sound cues to locate enemies
2. **Contact**: Engage enemies within visual or audio range
3. **Kill**: Execute tactical combat decisions
4. **Survive**: Maintain positioning and resources for next encounter

## Vision and Sound Systems

### Vision System

#### Display Range (Screen Area)
- **Display Area**: 15x15 square grid that follows the player
- **Player Position**: Player positioned 3 units from the bottom edge of the display square
- **Direction-Based**: Display square rotates with player's facing direction
- **Always Rendered**: This area is always displayed on screen regardless of lighting

#### Vision Range (Actual Sight)
- **Lighting-Dependent**: Players can only see objects within display range when they have light
- **Without Light**: Display range shows only darkness/fog, hiding obstacles, items, and enemies
- **With Light**: Reveals obstacles, items, enemies, and environmental features within display range
- **Dynamic Lighting**: Light sources affect what can be seen within the display range

### Sound System (Innovation Feature)
- **Three-Layer Sound Rings**: Visual representation of audio events in non-visible areas
- **Eight-Directional Audio**: Sound cues divided into 8 compass directions
- **Sound Event Types**:
  - Footsteps and movement
  - Weapon fire and reloading
  - Environmental interactions
  - Player death/elimination

### Player Representation
- Triangle-shaped entities for all players
- Server-authoritative positioning and rotation
- Smooth interpolation for movement rendering

## Game Architecture: Three-Layer Separation

### 1. Game State Layer
**Responsibility**: Store complete game world snapshot at any given moment
```go
type GameState struct {
    Players     map[string]*Player
    Projectiles []*Projectile
    Walls       []*Wall
    SoundEvents []*SoundEvent
}

type Player struct {
    ID       string
    IsAlive  bool
    Position Vector2D
    Rotation float64
    Health   int
    Lives    int
}
```

### 2. Game Logic Layer
**Responsibility**: Update and advance game state based on player input and game rules
```go
type GameLogic struct {
    PlayerMoveSpeed float64
    RotationSpeed   float64
    WeaponDamage    int
}

func (gl *GameLogic) Update(currentState *GameState, playerInputs map[string]PlayerInput, dt float64) {
    gl.processPlayerMovement(currentState, playerInputs, dt)
    gl.processShooting(currentState, playerInputs)
    gl.updateProjectiles(currentState, dt)
    gl.resolveCollisions(currentState)
    gl.processSoundEvents(currentState)
    gl.checkWinConditions(currentState)
}
```

### 3. Game Loop
**Responsibility**: Drive the entire game through continuous "Input → Logic → Render" cycles
```go
func main() {
    gameState := createInitialGameState()
    gameLogic := createGameLogic()
    renderer := createRenderer()

    for {
        playerInputs := renderer.PollInputs()
        gameLogic.Update(gameState, playerInputs, 1.0/60.0)
        renderer.Draw(gameState)
        time.Sleep(16 * time.Millisecond) // 60 FPS
    }
}
```

## Network Architecture: Host-Authoritative

### Network Model
- **Host Player**: Acts as both player and authoritative server
- **Client Players**: Send input, receive and render authoritative game state
- **Communication**: WebSocket for real-time bidirectional communication

### Network Synchronization
- **Client-Side Prediction**: Local player actions feel immediate
- **Entity Interpolation**: Smooth movement for remote players
- **Authoritative Reconciliation**: Server state resolves conflicts

### Network Optimization Strategy

#### Phase 1: JSON Full State (Development)
- Human-readable format for debugging
- Full `GameState` broadcast each frame
- Simple implementation for initial testing

#### Phase 2: Delta Updates (Optimization)
- Send only changed data between frames
- Event-driven system for one-time actions (shooting, death)
- Reduced network bandwidth usage

#### Phase 3: Binary Protocol (Production)
- Protocol Buffers (Protobuf) or MessagePack
- Maximum compression for production deployment
- Optimized for minimal latency

## Technical Stack

### Backend
- **Language**: Go with WebSocket for game logic and networking
- **Architecture**: Server-authoritative with 60 FPS game loop
- **Communication**: WebSocket for real-time multiplayer synchronization

### Frontend
- **Language**: TypeScript for type-safe development
- **Rendering**: PixiJS for high-performance 2D graphics
- **Input**: WASD movement, mouse aiming and shooting
- **UI**: Modern TypeScript UI components with game state management

### Desktop Application
- **Framework**: Wails v2 for cross-platform desktop deployment
- **Integration**: Go backend + TypeScript frontend in native desktop app
- **Distribution**: Single executable for Windows, macOS, and Linux

### Controls
- **Movement**: WASD keys (120 pixels/second)
- **Aiming**: Mouse cursor controls player rotation (4.0 radians/second)
- **Combat**: Left click for attacks
- **UI Navigation**: Mouse for menu interactions

## Development Roadmap

### Phase 1: Core Gameplay & Single-Player Validation
**Objective**: Complete core gameplay implementation in single-player environment
- Implement `GameState` and `GameLogic` structure
- Vision system with fog of war
- Sound event system with visual representation
- Basic combat and movement mechanics

### Phase 2: Basic Networking & Functional Synchronization
**Objective**: Enable multiplayer "connection and play" capability
- WebSocket implementation with Host-Authoritative model
- JSON format for complete `GameState` broadcasting
- Basic client-server communication

### Phase 3: Smooth Experience & Optimization
**Objective**: Eliminate lag and improve game feel
- Client-side entity interpolation
- Client-side prediction for local player
- Network latency handling and reconciliation

### Phase 4: Network Performance Optimization
**Objective**: Reduce network load for scalability
1. **Content Optimization**: Delta updates and event-driven messaging
2. **Format Optimization**: Binary protocol implementation (Protobuf)

## Map and Combat Design

### Map Design
- Building floor plan layout (800x600 pixels)
- Walls and obstacles for tactical positioning
- Doors and chokepoints for strategic control
- Spawn points distributed for balanced encounters

### Combat System

#### Initial Weapon Implementation (Development Phase)
- **Melee Weapon**: Knife with 1 body-length attack range
- **Ranged Weapon**: Pistol with projectile physics and advanced magazine system
- **Damage Model**: One-hit kills initially, expandable to health system

#### Pistol-Specific Mechanics
**Magazine System:**
- **Magazine Capacity**: 7 rounds per magazine
- **Magazine as Items**: Magazines are inventory items that can be found and collected
- **Ammunition Acquisition**: When finding bullets, only collect up to available magazine capacity
- **Last Magazine Rule**: Cannot reload when only one magazine remains (must preserve final magazine)

**Reload System (Pistol-Only):**
- **Normal Reload**: 3 seconds duration, preserves the partially-used magazine
- **Fast Reload**: 1 second duration, but discards the current magazine entirely
- **Range**: 7 units effective shooting distance
- **Strategic Decision**: Players must choose between speed and resource conservation

**Resource Management:**
- **Mid-Combat Decisions**: Fast reload for immediate advantage vs preserving ammunition
- **Post-Combat Planning**: Whether to replace partially-used magazines with full ones
- **Inventory Management**: Balancing magazine quantity against other essential items

#### Future Weapon Expansion (Post-Launch)
- **Melee Weapons**: One close-combat weapon (knife, axe, sword, etc.)
- **Ranged Weapons**: Two firearms with different characteristics (pistol, rifle, shotgun, etc.)
- **Equipment Items**: Multiple utility items (grenades, healing items, tools, etc.)
- **Inventory System**: Equipment management and weapon switching mechanics

#### Tactical Elements
- Sound masking and audio positioning
- Ambush tactics and positioning
- Resource and ammunition management

## Current Implementation Status

### Completed Features ✅
- Project structure (client/server/shared directories)
- Go backend with WebSocket server (port 8030)
- HTML5 Canvas frontend with game loop
- Basic player movement (WASD input)
- Mouse aiming system
- Real-time multiplayer synchronization
- Player spawn system with random positions
- Vision system: Fog of war with circular + cone visibility
- Movement/Rotation speed balancing (120px/s, 4.0 rad/s)
- Main menu system: Complete game flow (Menu → Game → Results → Menu)
- Game state management with proper transitions
- Result screen with statistics tracking
- Game mode selection (Solo, Multiplayer, Practice)

### Next Development Priorities
1. Sound event system implementation
2. Basic map obstacles and collision detection
3. Combat system (melee and ranged)
4. Enemy AI and spawning
5. Health/damage system
6. Enhanced graphics and sound effects