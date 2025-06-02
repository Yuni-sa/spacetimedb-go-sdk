# Blackholio - SpacetimeDB Go Game Client

This is a Go implementation of a space arena MMORPG game client using SpacetimeDB.

## Features

- **Real-time multiplayer**: Play with other players in real-time across the cosmos
- **Space arena gameplay**: Control circles, eat food, consume other players, and grow larger
- **Player splitting**: Split your mass to move faster or escape danger
- **Live leaderboard**: See top players and your current ranking
- **Smooth graphics**: 60 FPS gameplay with camera following and zoom effects
- **WebSocket connection**: Live updates through SpacetimeDB's real-time API
- **Authentication**: Automatic identity creation and token persistence

## Prerequisites

- Go 1.24 or later
- A running SpacetimeDB instance (default: `http://localhost:3000`)
- The blackholio-go module deployed to your SpacetimeDB instance

## Setup

1. **Ensure you have the game module deployed**:
   Deploy the C# server module to your SpacetimeDB instance.

2. **Build the client**:
   ```bash
   go mod tidy
   go build -o blackholio .
   ```

3. **Run the client**:
   ```bash
   ./blackholio
   ```

## Usage

### Starting the Game

When you run the client:

1. It will automatically create a new SpacetimeDB identity
2. Connect to the game server via WebSocket
3. Show the main menu where you can enter your cosmic name
4. Press Enter to join the space arena

### Game Controls

- **Mouse**: Move your circle toward the cursor
- **Space**: Split your circle to move faster (requires minimum mass)
- **S**: Suicide (destroys all your circles)
- **Space** (when dead): Respawn in the arena

### Gameplay

- Consume food (small colored circles) to grow larger
- Consume smaller players to absorb their mass
- Avoid larger players who can consume you
- Split strategically to escape or chase other players
- Climb the cosmic leaderboard by gaining mass

### Example Session

```
BLACKHOLIO
Space Arena MMORPG
Enter your cosmic name: StarLord_

[Connecting to the cosmic arena...]
[Playing - control with mouse, split with SPACE]
Total Mass: 156
COSMIC LEADERBOARD
1. GalaxyKing - 2341
2. StarLord - 156
3. CosmicDust - 89
```

## Configuration

You can modify these constants in the code:

```go
const (
    HOST    = "http://localhost:3000"  // SpacetimeDB server URL
    DB_NAME = "blackholio-go"         // Database name
)
```