package main

import (
	"encoding/json"
	"image/color"
	"log"

	"github.com/Yuni-sa/spacetimedb-go-sdk/client"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// Player color palette
var ColorPalette = []color.RGBA{
	{100, 150, 255, 255}, // Blue
	{255, 200, 100, 255}, // Orange
	{255, 50, 150, 255},  // Pink
	{100, 255, 150, 255}, // Green
	{150, 100, 255, 255}, // Purple
	{255, 100, 100, 255}, // Red
	{100, 255, 255, 255}, // Cyan
}

// Circle represents a player's circle
type Circle struct {
	EntityID      uint    `json:"entity_id"`
	PlayerID      uint    `json:"player_id"`
	Direction     Vector2 `json:"direction"`
	Speed         float32 `json:"speed"`
	LastSplitTime int64   `json:"last_split_time"`
}

func (gm *GameManager) getTotalMassForPlayer(playerID uint) uint {
	gm.entitiesMutex.RLock()
	gm.circlesMutex.RLock()
	defer gm.entitiesMutex.RUnlock()
	defer gm.circlesMutex.RUnlock()

	var totalMass uint
	for _, circle := range gm.circles {
		if circle.PlayerID == playerID {
			if entity, exists := gm.entities[circle.EntityID]; exists {
				totalMass += entity.Mass
			}
		}
	}
	return totalMass
}

// Circle drawing methods
func (g *Game) drawCircles(screen *ebiten.Image) {
	g.gameManager.circlesMutex.RLock()
	defer g.gameManager.circlesMutex.RUnlock()

	for _, circle := range g.gameManager.circles {
		if entity, exists := g.gameManager.entities[circle.EntityID]; exists {
			g.drawCircle(screen, entity, circle)
		}
	}
}

func (g *Game) drawCircle(screen *ebiten.Image, entity *Entity, circle *Circle) {
	if g.cameraController.camera == nil {
		return
	}

	// Use CurrentPosition for smooth interpolated rendering
	screenX, screenY := g.cameraController.camera.ApplyCameraTransformToPoint(entity.CurrentPosition.X, entity.CurrentPosition.Y)
	if !g.isOnScreen(screenX, screenY) {
		return
	}

	// Color based on player ID (matching Unity implementation)
	bodyColor := ColorPalette[circle.PlayerID%uint(len(ColorPalette))]

	// Draw directly with vector graphics, manually scaling by zoom (like demo)
	radius := entity.Radius * float32(g.cameraController.camera.ZoomFactor)

	// Draw body
	vector.DrawFilledCircle(screen, float32(screenX), float32(screenY), radius, bodyColor, false)

	// Draw player name above circle
	if g.gameManager.localPlayer != nil {
		g.gameManager.playersMutex.RLock()
		if player, exists := g.gameManager.players[circle.PlayerID]; exists {
			nameY := int(screenY - float64(radius) - 20)
			nameX := int(screenX) - len(player.Name)*3 // Center the text roughly
			ebitenutil.DebugPrintAt(screen, player.Name, nameX, nameY)
		}
		g.gameManager.playersMutex.RUnlock()
	}
}

// Circle management methods
func (gm *GameManager) processInitialCircles(updates []client.TableUpdateEntry) {
	for _, update := range updates {
		for _, insertStr := range update.Inserts {
			var circle Circle
			if err := json.Unmarshal([]byte(insertStr), &circle); err == nil {
				gm.circles[circle.EntityID] = &circle
			}
		}
	}
}

func (gm *GameManager) processCircleUpdates(update client.TableUpdateEntry) {
	gm.circlesMutex.Lock()
	defer gm.circlesMutex.Unlock()

	for _, deleteStr := range update.Deletes {
		if circle := parseCircle(deleteStr); circle != nil {
			delete(gm.circles, circle.EntityID)
		}
	}

	for _, insertStr := range update.Inserts {
		if circle := parseCircle(insertStr); circle != nil {
			gm.circles[circle.EntityID] = circle
		}
	}

	gm.updatePlayerCirclesUnsafe() // safe since we hold the mutex
}

func (gm *GameManager) updatePlayerCircles() {
	if gm.localPlayer == nil {
		return
	}

	gm.circlesMutex.RLock()
	defer gm.circlesMutex.RUnlock()

	var ownedCircles []uint
	for _, circle := range gm.circles {
		if circle.PlayerID == gm.localPlayer.PlayerID {
			ownedCircles = append(ownedCircles, circle.EntityID)
		}
	}

	gm.localPlayer.OwnedCircles = ownedCircles

	if len(ownedCircles) == 0 && gm.state == STATE_PLAYING {
		log.Printf("Player died - changing to DEAD state")
		gm.state = STATE_DEAD
	} else if len(ownedCircles) > 0 && gm.state == STATE_CONNECTING {
		log.Printf("Player spawned - changing to PLAYING state")
		gm.state = STATE_PLAYING
	}
}

// Unsafe version that assumes caller already holds circles mutex
func (gm *GameManager) updatePlayerCirclesUnsafe() {
	if gm.localPlayer == nil {
		return
	}

	// Don't acquire mutex here - caller already holds it
	var ownedCircles []uint
	for _, circle := range gm.circles {
		if circle.PlayerID == gm.localPlayer.PlayerID {
			ownedCircles = append(ownedCircles, circle.EntityID)
		}
	}

	previousCircleCount := len(gm.localPlayer.OwnedCircles)
	gm.localPlayer.OwnedCircles = ownedCircles

	if len(ownedCircles) == 0 && gm.state == STATE_PLAYING {
		log.Printf("PLAYER DEATH: Had %d circles, now has 0. Previous circles: %v",
			previousCircleCount, gm.localPlayer.OwnedCircles)
		log.Printf("PLAYER DEATH: Total circles in game: %d", len(gm.circles))

		// Show which circles exist in the game
		playerCircles := 0
		for _, circle := range gm.circles {
			if circle.PlayerID == gm.localPlayer.PlayerID {
				playerCircles++
			}
		}
		log.Printf("PLAYER DEATH: Circles belonging to player %d: %d", gm.localPlayer.PlayerID, playerCircles)

		gm.state = STATE_DEAD
	} else if len(ownedCircles) > 0 && gm.state == STATE_CONNECTING {
		log.Printf("Player spawned with %d circles - changing to PLAYING state", len(ownedCircles))
		gm.state = STATE_PLAYING
	}
}

// Circle parsing function
func parseCircle(insertStr string) *Circle {
	var raw []any
	if err := json.Unmarshal([]byte(insertStr), &raw); err != nil || len(raw) < 5 {
		return nil
	}

	entityID, _ := raw[0].(float64)
	playerID, _ := raw[1].(float64)

	// Parse direction vector [x, y]
	directionData, _ := raw[2].([]interface{})
	var direction Vector2
	if len(directionData) == 2 {
		direction.X, _ = directionData[0].(float64)
		direction.Y, _ = directionData[1].(float64)
	}

	speed, _ := raw[3].(float64)
	lastSplitTime, _ := raw[4].(float64)

	return &Circle{
		EntityID:      uint(entityID),
		PlayerID:      uint(playerID),
		Direction:     direction,
		Speed:         float32(speed),
		LastSplitTime: int64(lastSplitTime),
	}
}
