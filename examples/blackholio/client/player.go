package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/Yuni-sa/spacetimedb-go-sdk/client"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// Player represents a player
type Player struct {
	Identity struct {
		Identity string `json:"__identity__"`
	} `json:"identity"`
	PlayerID     uint   `json:"player_id"`
	Name         string `json:"name"`
	OwnedCircles []uint
}

// PlayerController handles input
type PlayerController struct {
	gameManager *GameManager
}

func NewPlayerController(gm *GameManager) *PlayerController {
	return &PlayerController{gameManager: gm}
}

// PlayerController methods
func (pc *PlayerController) HandleInput() {
	if pc.gameManager.wsConn == nil || pc.gameManager.localPlayer == nil {
		return
	}

	mouseX, mouseY := ebiten.CursorPosition()
	currentTime := float64(time.Now().UnixNano()) / 1e9

	// Throttled input
	if currentTime-pc.gameManager.lastMovementSendTime >= SEND_UPDATES_FREQUENCY {
		pc.gameManager.lastMovementSendTime = currentTime

		// Calculate direction from screen center
		windowWidth, windowHeight := ebiten.WindowSize()
		centerX := float64(windowWidth) / 2
		centerY := float64(windowHeight) / 2

		dirX := float64(mouseX) - centerX
		dirY := float64(mouseY) - centerY
		direction := Vector2{dirX, dirY}.Normalize()

		pc.gameManager.sendPlayerInput(direction)
	}

	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		pc.gameManager.playerSplit()
	}

	// Suicide command (S key)
	if inpututil.IsKeyJustPressed(ebiten.KeyS) {
		pc.gameManager.suicide()
	}
}

// Player processing functions
func (gm *GameManager) processInitialPlayers(updates []client.TableUpdateEntry) {
	gm.playersMutex.Lock()
	defer gm.playersMutex.Unlock()

	for _, update := range updates {
		for _, insertStr := range update.Inserts {
			var player Player
			if err := json.Unmarshal([]byte(insertStr), &player); err == nil {
				gm.players[player.PlayerID] = &player
			}
		}
	}
}

func (gm *GameManager) processPlayerUpdates(update client.TableUpdateEntry) {
	gm.playersMutex.Lock()
	defer gm.playersMutex.Unlock()

	for _, deleteStr := range update.Deletes {
		if player := parsePlayer(deleteStr); player != nil {
			delete(gm.players, player.PlayerID)
		}
	}

	for _, insertStr := range update.Inserts {
		if player := parsePlayer(insertStr); player != nil {
			gm.players[player.PlayerID] = player
		}
	}
}

func (gm *GameManager) findLocalPlayer() {
	gm.playersMutex.RLock()
	defer gm.playersMutex.RUnlock()

	for _, player := range gm.players {
		localIdentityWithPrefix := "0x" + gm.localIdentity

		if player.Identity.Identity == localIdentityWithPrefix {
			log.Printf("Local player found: %s", player.Name)
			gm.localPlayer = player
			gm.updatePlayerCircles()
			return
		}
	}

	log.Printf("WARNING: Local player not found")
}

// Player actions
func (gm *GameManager) sendPlayerInput(direction Vector2) {
	directionJSON, _ := json.Marshal(direction)
	args := fmt.Sprintf(`[%s]`, string(directionJSON))
	gm.wsConn.SendCallReducer("UpdatePlayerInput", args, 0)
}

func (gm *GameManager) playerSplit() {
	gm.wsConn.SendCallReducer("PlayerSplit", "[]", 0)
}

func (gm *GameManager) suicide() {
	if err := gm.wsConn.SendCallReducer("Suicide", "[]", 0); err != nil {
		log.Printf("Failed to suicide: %v", err)
	}
}

func (gm *GameManager) respawn() {
	if err := gm.wsConn.SendCallReducer("Respawn", "[]", 0); err != nil {
		log.Printf("Failed to respawn: %v", err)
	} else {
		gm.state = STATE_PLAYING
	}
}

// Player GUI functions
func (g *Game) drawPlayerUI(screen *ebiten.Image) {
	// GUI Mass Display (only for local player when connected and alive)
	if g.gameManager.localPlayer != nil && g.gameManager.state == STATE_PLAYING {
		totalMass := g.gameManager.getTotalMass()
		massText := fmt.Sprintf("Total Mass: %d", totalMass)
		ebitenutil.DebugPrintAt(screen, massText, 20, 20)
	}
}

// Player parsing function
func parsePlayer(insertStr string) *Player {
	var raw []any
	if err := json.Unmarshal([]byte(insertStr), &raw); err != nil || len(raw) != 3 {
		return nil
	}

	identityArray, _ := raw[0].([]any)
	playerID, _ := raw[1].(float64)
	name, _ := raw[2].(string)

	if len(identityArray) != 1 {
		return nil
	}

	identityStr, _ := identityArray[0].(string)

	return &Player{
		PlayerID: uint(playerID),
		Identity: struct {
			Identity string `json:"__identity__"`
		}{Identity: identityStr},
		Name: name,
	}
}
