package main

import (
	"encoding/json"
	"image/color"

	"github.com/Yuni-sa/spacetimedb-go-sdk/client"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// Food color palette - earthy/muted tones that complement player colors
var FoodColorPalette = []color.RGBA{
	{255, 215, 120, 255}, // Gold
	{220, 190, 100, 255}, // Dark gold
	{180, 140, 70, 255},  // Bronze
	{240, 200, 140, 255}, // Light wheat
	{200, 160, 90, 255},  // Amber
	{160, 130, 60, 255},  // Dark amber
}

// Food represents food entities
type Food struct {
	EntityID uint `json:"entity_id"`
}

// Food drawing functions
func (g *Game) drawFood(screen *ebiten.Image, entity *Entity) {
	if g.cameraController.camera == nil {
		return
	}

	// Use CurrentPosition for smooth interpolated rendering
	screenX, screenY := g.cameraController.camera.ApplyCameraTransformToPoint(entity.CurrentPosition.X, entity.CurrentPosition.Y)

	if !g.isOnScreen(screenX, screenY) {
		return
	}

	// Color based on entity ID for variety
	foodColor := FoodColorPalette[entity.EntityID%uint(len(FoodColorPalette))]

	// Draw directly with vector graphics, manually scaling by zoom
	radius := entity.Radius * float32(g.cameraController.camera.ZoomFactor)
	vector.DrawFilledCircle(screen, float32(screenX), float32(screenY), radius, foodColor, true)
}

// Food processing functions
func (gm *GameManager) processInitialFoods(updates []client.TableUpdateEntry) {
	gm.foodsMutex.Lock()
	defer gm.foodsMutex.Unlock()

	for _, update := range updates {
		for _, insertStr := range update.Inserts {
			var food Food
			if err := json.Unmarshal([]byte(insertStr), &food); err == nil {
				gm.foods[food.EntityID] = &food
			}
		}
	}
}

func (gm *GameManager) processFoodUpdates(update client.TableUpdateEntry) {
	gm.foodsMutex.Lock()
	defer gm.foodsMutex.Unlock()

	for _, deleteStr := range update.Deletes {
		if food := parseFood(deleteStr); food != nil {
			delete(gm.foods, food.EntityID)
		}
	}

	for _, insertStr := range update.Inserts {
		if food := parseFood(insertStr); food != nil {
			gm.foods[food.EntityID] = food
		}
	}
}

// Food parsing function
func parseFood(insertStr string) *Food {
	var raw []any
	if err := json.Unmarshal([]byte(insertStr), &raw); err != nil || len(raw) != 1 {
		return nil
	}

	entityID, _ := raw[0].(float64)

	return &Food{
		EntityID: uint(entityID),
	}
}
