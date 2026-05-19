# Terminal 2.5D Raycasting Renderer

This document describes the technical implementation of the 2.5D Doom-style raycasting renderer for the terminal client.

## Resolution Scaling

The renderer uses a half-block character technique to achieve doubled vertical resolution:

| Resolution | Description |
|------------|-------------|
| Physical | Terminal W × H characters |
| Logical | W × (2H) pixels |

The Unicode character `▀` (U+2580, "Upper Half Block") is used to render two vertical pixels per character cell:
- **Foreground color** = upper pixel
- **Background color** = lower pixel

## Projection Math

### Field of View
- FOV = 90° (π/2 radians)
- This simplifies projection since tan(45°) = 1

### Key Constants
- **D_proj** (Projection Distance) = CanvasWidth / 2
- **Horizon** = CanvasHeight / 2 (center of logical canvas)

### Wall Projection Formula

For each ray that hits a wall at distance `Z_depth`:

```
Y_top    = Horizon - ((BaseElevation + Height - ViewHeight) / Z_depth) × D_proj
Y_bottom = Horizon - ((BaseElevation - ViewHeight) / Z_depth) × D_proj
```

Where:
- `BaseElevation` = bottom of wall in world units
- `Height` = wall height in world units
- `ViewHeight` = player eye height in world units
- `Z_depth` = perpendicular distance to wall (with fish-eye correction applied)

## World Unit Standards

| Entity | Property | Default Value |
|--------|----------|---------------|
| Player | ViewHeight | 1.7 |
| Wall | Height | 2.0 |
| Wall | BaseElevation | 0.0 |

These values create a realistic perspective where:
- The player's eye level is slightly below the top of standard walls
- Standard walls appear at consistent heights
- Variable height walls (half-walls, raised platforms) render correctly

## Character Rendering

### Half-Block Technique

Each terminal row represents two logical pixels:

```
Terminal Row n:
┌─────────────┐
│ Upper (2n)  │  ← Foreground color
│ Lower (2n+1)│  ← Background color
└─────────────┘
```

Character: `▀` (U+2580) with ANSI 256-color codes

### ANSI Color Codes

```
Foreground: \033[38;5;{color}m
Background: \033[48;5;{color}m
Reset:      \033[0m
```

### Color Palette

| Surface | 256-Code | Description |
|---------|----------|-------------|
| Floor | 238 | Dark gray |
| Ceiling | 235 | Darker gray |
| Wall (near) | 255 | Bright white |
| Wall (mid) | 245 | Medium gray |
| Wall (far) | 240 | Dim gray |
| Dummy (near) | 229 | Light yellow |
| Dummy (mid) | 221 | Light yellow |
| Dummy (far) | 178 | Golden yellow |

Per-entity surface colors (e.g. the training dummy uses the Dummy palette) are selected by the renderer based on `ShapeType` / entity context.

## Collider Shapes

The raycaster supports two collider footprints, dispatched on `RaycastResult.ShapeType`:

- **AABB** (`rayBoxIntersect`): slab test against `HalfX` / `HalfY`. Used for walls and rectangular obstacles.
- **Circle** (`rayCircleIntersect`, `ShapeType == 1`): quadratic intersect against `Radius`. Used for player/character pillars.

Both intersection tests are purely 2D on the XY plane; vertical visibility (`WallHeight` + `BaseElevation`) is resolved later in the projection step, so tall and short objects render correctly as long as their footprint is hit by a ray.

## File Structure

```
internal/terminal/
├── construct.go        # Terminal setup (raw mode, signals)
├── locale.go           # Localization
├── manager.go          # Game manager / main loop
├── raycast/
│   ├── raycast.go      # 2D ray-vs-collider intersection (AABB + circle)
│   ├── raycast_test.go # Raycast tests
│   └── renderer25d.go  # 2.5D half-block renderer
├── ui/
│   └── overlay.go      # UI layer (crosshair, HUD, weapon)
├── state/
│   ├── mainmenu.go     # Main menu state
│   ├── setting.go      # Settings state
│   ├── singleplayer.go # Single-player game state
│   ├── multiplayer.go  # Multi-player game state
│   └── resize.go       # Terminal resize handling
├── network/            # WebSocket client
├── session/            # Session/reconnect support
├── debug/              # Debug overlays (e.g. mini-map)
└── README.md           # This file
```

## RaycastResult Structure

```go
type RaycastResult struct {
    Distance      float64 // Perpendicular distance (fish-eye corrected)
    Hit           bool    // Whether ray hit anything
    WallHeight    float64 // Height of hit collider in world units
    BaseElevation float64 // Base elevation of hit collider in world units
    EntityID      uint64  // ID of hit entity
    ShapeType     uint8   // 0 = AABB, 1 = circle (mirrors state.ColliderShape)
}
```

`CastRays` returns `[][]RaycastResult` — one slice per ray, sorted **farthest first** so the renderer can paint back-to-front (painter's algorithm).

## Renderer25D Usage

```go
// Create renderer
renderer := raycast.NewRenderer25D(termWidth, termHeight, viewHeight)

// Cast rays with view height parameter
results := raycast.CastRays(playerX, playerY, playerDir, viewHeight, colliders, numRays)

// Render to buffer
renderer.Render(results)

// Write to output
var buf bytes.Buffer
renderer.WriteToBuffer(&buf)
```

## UI Overlay System

The UI layer provides:
- **Crosshair**: Center of screen (`+`)
- **HUD**: Health (top-left), Ammo (top-right)
- **Weapon**: ASCII art sprite at bottom-center

```go
uiLayer := ui.NewUILayer(width, height)
uiLayer.SetHealth(100)
uiLayer.SetAmmo(12, 12)
uiLayer.Overlay(buffer, colors)
```
