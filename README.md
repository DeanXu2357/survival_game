# Survival Game

A top-down 2D survival shooter game with limited vision mechanics.

## Project Structure

```
survival/
├── client/          # Frontend (HTML5 Canvas)
├── server/          # Backend (Go)
├── shared/          # Shared game logic
├── spec.md          # Game specification
└── README.md        # This file
```

## Development Setup

### Backend (Go)
```bash
cd server
go mod init survival-server
go run main.go
```

### Frontend
Open `client/index.html` in a web browser or serve via a local HTTP server.

## Game Controls
- WASD: Move player
- Mouse: Aim and turn
- Left Click: Attack (melee/ranged)