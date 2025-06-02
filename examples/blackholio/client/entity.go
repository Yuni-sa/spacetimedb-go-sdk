package main

import (
	"encoding/json"
	"log"
	"math"

	"github.com/Yuni-sa/spacetimedb-go-sdk/client"
	"github.com/hajimehoshi/ebiten/v2"
)

// Entity represents a game entity
type Entity struct {
	EntityID uint    `json:"entity_id"`
	Position Vector2 `json:"position"`
	Mass     uint    `json:"mass"`
	Radius   float32 // Calculated from mass

	// Interpolation fields
	LerpTime           float64
	LerpStartPosition  Vector2
	LerpTargetPosition Vector2
	CurrentPosition    Vector2 // The interpolated position for rendering
}

// Entity methods
func (e *Entity) spawn() {
	e.Radius = massToRadius(e.Mass)
	// Initialize interpolation
	e.LerpStartPosition = e.Position
	e.LerpTargetPosition = e.Position
	e.CurrentPosition = e.Position
	e.LerpTime = LERP_DURATION_SEC // Start fully lerped
}

func (e *Entity) onEntityUpdated(newEntity Entity) {
	// Set up interpolation
	e.LerpTime = 0.0
	e.LerpStartPosition = e.CurrentPosition // Start from current rendered position
	e.LerpTargetPosition = newEntity.Position

	// Update other properties
	e.Position = newEntity.Position
	e.Mass = newEntity.Mass
	e.Radius = massToRadius(e.Mass)
}

// Update interpolation
func (e *Entity) updateInterpolation(deltaTime float64) {
	// Interpolate position
	e.LerpTime = math.Min(e.LerpTime+deltaTime, LERP_DURATION_SEC)
	lerpProgress := e.LerpTime / LERP_DURATION_SEC

	// Linear interpolation between start and target
	e.CurrentPosition.X = e.LerpStartPosition.X + (e.LerpTargetPosition.X-e.LerpStartPosition.X)*lerpProgress
	e.CurrentPosition.Y = e.LerpStartPosition.Y + (e.LerpTargetPosition.Y-e.LerpStartPosition.Y)*lerpProgress
}

// Utility functions
func massToRadius(mass uint) float32 {
	return float32(math.Sqrt(float64(mass)))
}

// Drawing functions
func (g *Game) drawEntities(screen *ebiten.Image) {
	if g.cameraController.camera == nil {
		return
	}

	g.gameManager.entitiesMutex.RLock()
	g.gameManager.foodsMutex.RLock()
	defer g.gameManager.entitiesMutex.RUnlock()
	defer g.gameManager.foodsMutex.RUnlock()

	// Draw food
	for _, food := range g.gameManager.foods {
		if entity, exists := g.gameManager.entities[food.EntityID]; exists {
			g.drawFood(screen, entity)
		}
	}
}

// Entity processing functions
func (gm *GameManager) processInitialEntities(updates []client.TableUpdateEntry) {
	gm.entitiesMutex.Lock()
	defer gm.entitiesMutex.Unlock()

	for _, update := range updates {
		for _, insertStr := range update.Inserts {
			var entity Entity
			if err := json.Unmarshal([]byte(insertStr), &entity); err == nil {
				entity.spawn()
				gm.entities[entity.EntityID] = &entity
			}
		}
	}
}

func (gm *GameManager) processEntityUpdates(update client.TableUpdateEntry) {
	gm.entitiesMutex.Lock()
	defer gm.entitiesMutex.Unlock()

	// First pass: collect all inserts to see if we're getting delete+insert pairs
	insertMap := make(map[uint]*Entity)
	for _, insertStr := range update.Inserts {
		if entity := parseEntity(insertStr); entity != nil {
			insertMap[entity.EntityID] = entity
		}
	}

	// Process deletes, but skip if we have a corresponding insert (delete+insert = update)
	for _, deleteStr := range update.Deletes {
		if entity := parseEntity(deleteStr); entity != nil {
			// If we have a delete+insert pair, treat as update instead of delete+insert
			if newEntity, hasInsert := insertMap[entity.EntityID]; hasInsert {
				// This is really an update, not a delete+insert
				if existing, exists := gm.entities[entity.EntityID]; exists {
					existing.onEntityUpdated(*newEntity)
				} else {
					newEntity.spawn()
					gm.entities[entity.EntityID] = newEntity
				}
				// Remove from insert map so we don't process it again
				delete(insertMap, entity.EntityID)
				continue
			}

			// This is a real delete
			delete(gm.entities, entity.EntityID)

			// Clean up orphaned circle reference
			gm.circlesMutex.Lock()
			delete(gm.circles, entity.EntityID)
			gm.circlesMutex.Unlock()
		}
	}

	// Process remaining inserts (those that weren't part of delete+insert pairs)
	for _, entity := range insertMap {
		entity.spawn()
		gm.entities[entity.EntityID] = entity
	}

	// Update player circles after entity changes
	gm.updatePlayerCircles()
}

// Parsing functions
func parseEntity(insertStr string) *Entity {
	var raw []any
	if err := json.Unmarshal([]byte(insertStr), &raw); err != nil || len(raw) != 3 {
		return nil
	}

	entityID, ok := raw[0].(float64)
	if !ok {
		return nil
	}

	posData, ok := raw[1].([]interface{})
	if !ok || len(posData) != 2 {
		return nil
	}

	// Check for null positions from server
	if posData[0] == nil || posData[1] == nil {
		log.Printf("CRITICAL: Server sent null position for EntityID:%d - IGNORING", uint(entityID))
		return nil
	}

	posX, ok := posData[0].(float64)
	if !ok {
		return nil
	}

	posY, ok := posData[1].(float64)
	if !ok {
		return nil
	}

	mass, ok := raw[2].(float64)
	if !ok {
		return nil
	}

	// Check for suspicious (0,0) positions - this indicates a teleportation bug
	if posX == 0.0 && posY == 0.0 {
		log.Printf("WARNING: Entity at (0,0) detected! EntityID:%d Mass:%d", uint(entityID), uint(mass))
	}

	return &Entity{
		EntityID: uint(entityID),
		Position: Vector2{X: posX, Y: posY},
		Mass:     uint(mass),
	}
}
