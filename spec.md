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

#### Camera System
- **Display Area**: 25x25 square grid centered on the player
- **Camera Position**: Player positioned at screen center
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
- **Visual**: Triangle-shaped entities for all players
- **Collision**: Circular hitbox with diameter 1.0 meter (radius 0.5m)
- **Unit System**: Backend uses 1 float = 1 meter for all calculations
- Server-authoritative positioning and rotation
- Smooth interpolation for movement rendering

## Game Architecture: Three-Layer Separation

### 1. Game State Layer
**Responsibility**: Store complete game world snapshot at any given moment

The game state contains all players, projectiles, walls, and other game objects. The actual implementation includes additional fields for thread safety, spatial optimization, and game management. See `internal/game/state.go` for the complete structure.

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

## Network Architecture: Server-Authoritative

### Network Model
- **Dedicated Server**: Go server handles all game logic and state
- **Client Players**: Send input, receive and render authoritative game state  
- **Communication**: WebSocket for real-time bidirectional communication

### Client Connection Flow
1. **Client ID Generation**: Client generates unique ID on installation (MAC hash, etc.)
2. **WebSocket Connection**: Client connects with connection data in request body (JSON format)
3. **Hub Registration**: Server assigns Player ID and maps to Client ID
4. **Session Management**: Client stores Player ID for reconnection
5. **Game Participation**: Client sends input, receives game state updates

### Connection Management
- **Client-Provided IDs**: Clients generate and provide their own unique identifiers
- **Hub Architecture**: Central hub manages multiple game rooms and client connections
- **Player Mapping**: Hub maintains Client ID ↔ Player ID relationships
- **Reconnection Support**: Clients can reconnect using stored Player ID
- **Invite System**: Future feature for secure client access (bypassed in development)

### Network Synchronization
- **Server Authority**: All game logic, collision, and state updates on server
- **Client Rendering**: Clients handle input, rendering, and UI only
- **Real-time Updates**: 60 FPS server tick rate with state broadcasting
- **Message Protocol**: JSON messages for development, binary for production

### Communication Requirements

#### Development Phase
- WebSocket + JSON for easy debugging and development
- Client-provided connection identifiers for session management
- Simple message routing between Hub, Room, and Client layers
- Basic error handling and connection recovery

#### Production Phase  
- Desktop application integration with Wails framework
- Protocol Buffers for optimized binary serialization
- Advanced reconnection and state synchronization
- Native desktop features and performance optimizations

## Technical Requirements

### Core Technologies
- **Backend**: Go for game logic and server
- **Frontend**: TypeScript + PixiJS for 2D graphics
- **Architecture**: Server-authoritative multiplayer
- **Target Platform**: Cross-platform desktop application

### Development Approach
- Start with web-based frontend and standalone backend for rapid iteration
- Migrate to integrated desktop application for production release
- Maintain consistent game logic throughout development phases

### Camera System Requirements
- Fixed upward-facing camera with player always at screen center
- World rotates around player instead of player rotating on screen
- Smooth camera movement and rotation
- Efficient coordinate transformation for all game objects


### Controls
- **Movement**: WASD keys at 1.0 meters/second
- **Aiming**: Mouse cursor controls player rotation at 2.0 radians/second 
- **Combat**: Left click for attacks
- **UI Navigation**: Mouse for menu interactions

## Acceptance Criteria

### Core Gameplay Requirements
- [ ] Real-time multiplayer with 2-8 players
- [ ] Vision system with fog of war (circular + cone visibility)
- [ ] Sound event system with visual 3-ring representation
- [ ] Combat system with melee (knife) and ranged (pistol) weapons
- [ ] Player collision and map boundaries
- [ ] 60 FPS smooth gameplay with minimal network latency

### Game Modes
- [ ] Solo mode for single-player practice
- [ ] Multiplayer mode with player vs player combat
- [ ] Practice mode for testing and experimentation

### User Interface
- [ ] Main menu with mode selection
- [ ] In-game HUD with player status
- [ ] Game state transitions (Menu → Game → Results → Menu)
- [ ] Settings and configuration options

### Technical Performance
- [ ] Sub-50ms network latency for multiplayer
- [ ] Smooth 60 FPS rendering
- [ ] Cross-platform desktop application
- [ ] Stable multiplayer sessions without disconnections

### Final Deliverable
- Cross-platform desktop application (Windows, macOS, Linux)
- Complete game with all specified features
- Optimized performance and user experience

## Map and Combat Design

### Map Design
- Building floor plan layout (200x200 meters)
- Walls and obstacles for tactical positioning
- Doors and chokepoints for strategic control
- Spawn points distributed for balanced encounters

### Combat System

#### Pistol Weapon Specifications
- **Magazine Capacity**: 9 rounds per magazine
- **Range**: 70 meters effective distance
- **Fire Rate**: 0.5 second cooldown between shots
- **Reload Time**: 3 seconds (normal reload)
- **Key Bindings**: Spacebar to shoot, R key to reload

#### Bullet Specifications
- **Flight Speed**: 50 backend units per second
- **Visual Appearance**: Yellow 3x6 pixel rectangle
- **Trail Effect**: Display bullet path for past 1 second (frontend implementation location must be marked)
- **Direction**: Fired in player's facing direction

#### Player Initial Equipment
- **Starting Loadout**: 1 pistol + 1 magazine (loaded with 9 bullets)
- **Other Weapons**: Knife and other weapons not implemented in this phase

#### User Interface Requirements
- **Weapon Ammo Display**: Bottom-right corner showing current weapon and remaining ammo count
- **Crosshair**: Green crosshair UI visible when ready to shoot, hidden during reload or cooldown
- **Visual Feedback**: Crosshair serves as visual indicator for shooting availability

#### Implementation Requirements

**Backend Implementation:**
- Modify Player struct to include weapon and magazine systems
- Implement shooting logic, reload logic, and cooldown mechanics
- Bullet collision detection not implemented initially - focus on display

**Frontend Implementation:**
- Key bindings: Spacebar for shooting, R key for reload
- Bullet rendering and trail effects
- Weapon UI and crosshair UI
- **Important**: Mark implementation location for bullet trail effect

**Network Synchronization:**
- Synchronize shooting and reload events between client and server
- Real-time bullet state updates

#### Technical Constraints
- Use existing Go backend + TypeScript PixiJS frontend architecture
- Follow existing WebSocket communication protocol
- Reserve space for future fast reload expansion (not implemented yet)

#### Future Weapon Expansion (Post-Launch)
- **Melee Weapons**: One close-combat weapon (knife, axe, sword, etc.)
- **Ranged Weapons**: Two firearms with different characteristics (pistol, rifle, shotgun, etc.)
- **Equipment Items**: Multiple utility items (grenades, healing items, tools, etc.)
- **Inventory System**: Equipment management and weapon switching mechanics

#### Tactical Elements
- Sound masking and audio positioning
- Ambush tactics and positioning
- Resource and ammunition management

