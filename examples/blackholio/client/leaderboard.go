package main

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// LeaderboardEntry represents a player entry in the leaderboard
type LeaderboardEntry struct {
	PlayerID uint
	Name     string
	Mass     uint
	IsLocal  bool
}

// Leaderboard manages the player ranking system
type Leaderboard struct {
	entries []LeaderboardEntry
}

func NewLeaderboard() *Leaderboard {
	return &Leaderboard{
		entries: make([]LeaderboardEntry, 0),
	}
}

func (lb *Leaderboard) Update(gm *GameManager) {
	gm.playersMutex.RLock()
	defer gm.playersMutex.RUnlock()

	// Clear previous entries
	lb.entries = lb.entries[:0]

	// Calculate mass for all players with circles
	for _, player := range gm.players {
		mass := gm.getTotalMassForPlayer(player.PlayerID)
		if mass > 0 { // Only include players with mass
			isLocal := gm.localPlayer != nil && player.PlayerID == gm.localPlayer.PlayerID
			lb.entries = append(lb.entries, LeaderboardEntry{
				PlayerID: player.PlayerID,
				Name:     player.Name,
				Mass:     mass,
				IsLocal:  isLocal,
			})
		}
	}

	// Sort by mass (descending)
	for i := 0; i < len(lb.entries); i++ {
		for j := i + 1; j < len(lb.entries); j++ {
			if lb.entries[j].Mass > lb.entries[i].Mass {
				lb.entries[i], lb.entries[j] = lb.entries[j], lb.entries[i]
			}
		}
	}
}

func (lb *Leaderboard) GetTop10PlusLocal() []LeaderboardEntry {
	// Get top 10
	top10 := make([]LeaderboardEntry, 0, 11)

	// Add top 10 players
	for i := 0; i < len(lb.entries) && i < 10; i++ {
		top10 = append(top10, lb.entries[i])
	}

	// Check if local player is in top 10
	localInTop10 := false
	for _, entry := range top10 {
		if entry.IsLocal {
			localInTop10 = true
			break
		}
	}

	// If local player not in top 10, add them
	if !localInTop10 {
		for _, entry := range lb.entries {
			if entry.IsLocal {
				top10 = append(top10, entry)
				break
			}
		}
	}

	return top10
}

func (g *Game) drawLeaderboard(screen *ebiten.Image) {
	// Get leaderboard data
	entries := g.gameManager.leaderboard.GetTop10PlusLocal()
	if len(entries) == 0 {
		return
	}

	// Leaderboard position and styling
	startX := g.screenWidth - 250
	startY := 50
	rowHeight := 25
	headerHeight := 30

	// Draw leaderboard background
	leaderboardHeight := headerHeight + len(entries)*rowHeight + 20
	vector.DrawFilledRect(screen, float32(startX-10), float32(startY-10),
		240, float32(leaderboardHeight), color.RGBA{0, 0, 0, 180}, false)

	// Draw border
	vector.StrokeRect(screen, float32(startX-10), float32(startY-10),
		240, float32(leaderboardHeight), 2, color.RGBA{100, 100, 100, 255}, false)

	// Draw header
	ebitenutil.DebugPrintAt(screen, "COSMIC LEADERBOARD", startX, startY)

	// Draw entries
	for i, entry := range entries {
		y := startY + headerHeight + i*rowHeight

		// Highlight local player
		if entry.IsLocal {
			vector.DrawFilledRect(screen, float32(startX-5), float32(y-2),
				230, float32(rowHeight-2), color.RGBA{100, 100, 0, 100}, false)
		}

		// Rank number
		rank := i + 1
		if i >= 10 && entry.IsLocal {
			// Find actual rank for local player
			for j, e := range g.gameManager.leaderboard.entries {
				if e.PlayerID == entry.PlayerID {
					rank = j + 1
					break
				}
			}
		}

		// Format: "1. PlayerName - 1234"
		name := entry.Name
		if len(name) > 12 {
			name = name[:12] + "..."
		}

		rankText := fmt.Sprintf("%d. %s - %d", rank, name, entry.Mass)
		ebitenutil.DebugPrintAt(screen, rankText, startX, y)
	}
}

func (gm *GameManager) getTotalMass() uint {
	if gm.localPlayer == nil {
		return 0
	}

	gm.entitiesMutex.RLock()
	defer gm.entitiesMutex.RUnlock()

	var totalMass uint
	for _, entityID := range gm.localPlayer.OwnedCircles {
		if entity, exists := gm.entities[entityID]; exists {
			totalMass += entity.Mass
		}
	}
	return totalMass
}
