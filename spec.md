# Survival Game Specification

## Game Overview

A top-down 2D dual-axis shooter game inspired by classic tank battle games, set in a building floor plan environment. Players control triangular characters with limited vision in a last-man-standing survival format.

## Core Game Mechanics

### Game Modes
- **Single Player vs AI**: Player fights against computer-controlled enemies
- **Multiplayer PvP**: Multiple human players compete in real-time

### Victory Condition
- Last surviving entity wins (elimination-based survival)

### Map Design
- Single floor building layout
- Includes walls, doors, and basic obstacles
- Floor plan style environment

### Player Representation
- Triangle shape for all entities (players and NPCs)
- One-hit kill system (initial development phase)
- Future consideration for health system with additional weapons

### Vision System
- **Limited Visibility**: Most of the map covered in darkness
- **Close Range Vision**: 1 player body-length radius around player
- **Directional Vision**: 45-degree cone extending 10 player body-lengths forward
- **Fog of War**: Areas outside vision range are completely black

### Weapons (Initial)
- **Melee Weapon**: Knife with 1 body-length attack range
- **Ranged Weapon**: Small pistol
- **Aiming**: Crosshair as a line in the center of vision cone

## Technical Architecture

### Client-Server Model
- Multiplayer support requires networked architecture
- Real-time synchronization for player movements and actions

### Technology Stack
- Platform: Web-based game for easy accessibility
- Frontend: HTML5 Canvas or WebGL for rendering
- Backend: Go for game logic and server
- Communication: WebSocket for real-time multiplayer

### Controls (Tentative)
- WASD: Player movement
- Mouse: Aiming and direction
- Left Click: Attack/Shoot
- Game runs at reasonable framerate for smooth gameplay

## Development Phases

### Phase 1: Core Mechanics
- Basic player movement and vision system
- Simple map with obstacles
- Melee combat implementation
- Single player vs AI mode

### Phase 2: Multiplayer
- Client-server architecture
- Real-time player synchronization
- PvP combat system

### Phase 3: Enhanced Features
- Additional weapons
- Health system implementation
- Improved AI behavior
- Enhanced map designs

## Future Considerations
- Additional weapon types and balancing
- Multiple maps and environments
- Game modes beyond elimination
- Audio and visual effects
- Spectator mode for multiplayer