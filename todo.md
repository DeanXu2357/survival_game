# Survival Game Implementation Todo List

## Completed Features ✅
- [x] Project structure setup (Go backend, client frontend)
- [x] Basic Player and State structures
- [x] Weapon system design (Knife, Pistol with magazine system)
- [x] PlayerInput structure definition
- [x] Game logic foundation (60 FPS game loop)
- [x] Room-based session management
- [x] WebSocket communication foundation

## Phase 1: Core Game Objects & Basic Controls (High Priority)

### 1.1 Game Objects Definition
- [ ] **Wall Structure Definition** - Define Wall struct with position, dimensions, and collision bounds
- [ ] **Projectile Structure Definition** - Bullet physics, trajectory, and lifetime
- [ ] **Update State Structure** - Add Projectiles and Walls arrays to game state

### 1.2 Player Movement System
- [ ] **Basic Player Movement** - WASD movement processing (120 pixels/second)
- [ ] **Player Rotation** - Left/Right rotation controls (4.0 radians/second)
- [ ] **Movement-Wall Collision** - Prevent players from walking through walls
- [ ] **Boundary Checking** - Keep players within map limits (800x600)

### 1.3 Weapon System
- [ ] **Weapon Switching Logic** - Toggle between knife and pistol
- [ ] **Player Initialization** - Default weapon setup and magazine inventory
- [ ] **Weapon State Management** - Track current weapon and availability

## Phase 2: Combat System (High Priority)

### 2.1 Shooting Mechanics
- [ ] **Pistol Shooting Logic** - Fire projectiles with 7-unit range
- [ ] **Knife Attack Logic** - Melee attack with 1-unit range
- [ ] **Ammo Consumption** - Decrease magazine ammo on pistol fire
- [ ] **Weapon Cooldowns** - Prevent spam firing

### 2.2 Reload System
- [ ] **Normal Reload** - 3-second duration, preserve magazine
- [ ] **Fast Reload** - 1-second duration, discard magazine
- [ ] **Reload Validation** - Check available magazines and weapon type
- [ ] **Reload Timing** - Async reload with proper state management

## Phase 3: Collision & Physics (Medium Priority)

### 3.1 Projectile System
- [ ] **Projectile Physics** - Movement, trajectory, and lifetime
- [ ] **Projectile-Wall Collision** - Bullets stop on wall impact
- [ ] **Projectile-Player Collision** - Damage calculation and hit detection
- [ ] **Projectile Cleanup** - Remove expired or collided projectiles

### 3.2 Collision Detection Optimization
- [ ] **Spatial Partitioning** - Use existing Grid system for collision optimization
- [ ] **AABB Collision** - Basic rectangular collision detection
- [ ] **Circle-Rectangle Collision** - Player (circle) vs Wall (rectangle)

## Phase 4: Game State Management (Medium Priority)

### 4.1 Game Loop Integration
- [ ] **Input Processing** - Process PlayerInput in game logic
- [ ] **State Updates** - Update all game objects per tick
- [ ] **Network Synchronization** - Broadcast state changes to clients
- [ ] **Delta Time Handling** - Smooth movement with variable frame rates

### 4.2 Game Rules
- [ ] **Health System** - One-hit kill implementation
- [ ] **Player Death** - Handle player elimination
- [ ] **Win Conditions** - Last player standing logic
- [ ] **Respawn System** - Player revival mechanics

## Phase 5: Testing & Validation (Low Priority)

### 5.1 Unit Testing
- [ ] **Movement Tests** - Validate player movement and rotation
- [ ] **Collision Tests** - Test all collision detection scenarios
- [ ] **Weapon Tests** - Verify shooting, reloading, and switching
- [ ] **Game State Tests** - Ensure proper state management

### 5.2 Integration Testing
- [ ] **Multiplayer Testing** - Multi-player game scenarios
- [ ] **Network Testing** - Client-server communication validation
- [ ] **Performance Testing** - 60 FPS performance validation

## Implementation Notes

### Collision System Implementation Order:
1. **Wall-Player Movement Collision** (Phase 1) - Prevent walking through walls
2. **Projectile-Wall Collision** (Phase 3) - Bullets stop on walls
3. **Projectile-Player Collision** (Phase 3) - Damage detection
4. **Spatial Optimization** (Phase 3) - Use Grid system for performance

### Key Design Decisions:
- Server-authoritative architecture (all logic on backend)
- 60 FPS game loop with delta time support
- Magazine-based ammunition system for pistol
- One-hit kill damage model initially
- Grid-based spatial partitioning for collision optimization

### Current Architecture:
```
internal/game/
├── state.go    - Game objects and state management
├── logic.go    - Core game logic and input processing
└── room.go     - Session and network management
```