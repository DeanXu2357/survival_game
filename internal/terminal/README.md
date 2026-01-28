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

## File Structure

```
internal/terminal/
├── raycast/
│   ├── raycast.go      # Ray casting algorithm
│   └── renderer25d.go  # 2.5D half-block renderer
├── ui/
│   └── overlay.go      # UI layer (crosshair, HUD, weapon)
├── state/
│   └── singleplayer.go # Game state integration
└── README.md           # This file
```

## RaycastResult Structure

```go
type RaycastResult struct {
    Distance      float64  // Perpendicular distance to wall
    Hit           bool     // Whether ray hit anything
    WallHeight    float64  // Height of hit wall
    BaseElevation float64  // Base elevation of hit wall
    EntityID      uint64   // ID of hit entity
}
```

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
