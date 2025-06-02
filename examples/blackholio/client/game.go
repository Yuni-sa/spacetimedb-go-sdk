package main

import (
	"context"
	"encoding/json"
	"fmt"
	"image/color"
	"log"
	"math"
	"sync"
	"time"

	"github.com/Yuni-sa/spacetimedb-go-sdk/client"
	"github.com/google/uuid"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// Game constants
const (
	// Window constants
	DEFAULT_SCREEN_WIDTH  = 1920
	DEFAULT_SCREEN_HEIGHT = 1080
	WORLD_SIZE            = 1000

	// Database constants
	HOST    = "http://localhost:3000"
	DB_NAME = "blackholio-go"

	// Input constants
	SEND_UPDATES_PER_SEC   = 10
	SEND_UPDATES_FREQUENCY = 1.0 / SEND_UPDATES_PER_SEC

	// Interpolation constants
	LERP_DURATION_SEC = 0.1
)

// Game states
type GameState int

const (
	STATE_MENU GameState = iota
	STATE_CONNECTING
	STATE_PLAYING
	STATE_DEAD
)

// GameManager handles connection and game state
type GameManager struct {
	conn          *client.Client
	wsConn        *client.WebSocketConnection
	authToken     *client.AuthToken
	localIdentity string
	ctx           context.Context
	cancel        context.CancelFunc

	state        GameState
	playerName   string
	inputName    string
	connectError string

	// Entity management
	entities    map[uint]*Entity
	circles     map[uint]*Circle
	players     map[uint]*Player
	foods       map[uint]*Food
	localPlayer *Player

	// Thread safety
	entitiesMutex sync.RWMutex
	circlesMutex  sync.RWMutex
	playersMutex  sync.RWMutex
	foodsMutex    sync.RWMutex

	// Input handling
	lastMovementSendTime float64

	// Update throttling
	lastUpdateTime map[uint]time.Time

	// Leaderboard
	leaderboard *Leaderboard
}

// Game is the main game struct
type Game struct {
	gameManager      *GameManager
	playerController *PlayerController
	cameraController *CameraController

	screenWidth  int
	screenHeight int
	deltaTime    float64
	lastFrame    time.Time
}

func NewGame() *Game {
	game := &Game{
		screenWidth:  DEFAULT_SCREEN_WIDTH,
		screenHeight: DEFAULT_SCREEN_HEIGHT,
		lastFrame:    time.Now(),
	}

	// Initialize managers
	game.gameManager = NewGameManager()
	game.playerController = NewPlayerController(game.gameManager)
	game.cameraController = NewCameraController()

	// Start connection
	go game.gameManager.ConnectAndRun()

	return game
}

func NewGameManager() *GameManager {
	ctx, cancel := context.WithCancel(context.Background())

	authToken, err := client.NewAuthToken(client.WithAuthConfigFolder(".spacetime_blackholio"))
	if err != nil {
		log.Printf("Failed to create auth token: %v", err)
		authToken = nil
	}

	return &GameManager{
		authToken:      authToken,
		ctx:            ctx,
		cancel:         cancel,
		state:          STATE_MENU,
		entities:       make(map[uint]*Entity),
		circles:        make(map[uint]*Circle),
		players:        make(map[uint]*Player),
		foods:          make(map[uint]*Food),
		lastUpdateTime: make(map[uint]time.Time),
		leaderboard:    NewLeaderboard(),
	}
}

// Game interface methods
func (g *Game) Update() error {
	// Calculate deltaTime
	currentTime := time.Now()
	g.deltaTime = currentTime.Sub(g.lastFrame).Seconds()
	g.lastFrame = currentTime
	if g.deltaTime > 0.1 {
		g.deltaTime = 0.1
	}

	// Update based on state
	switch g.gameManager.state {
	case STATE_MENU:
		return g.updateMenu()
	case STATE_CONNECTING:
		return nil
	case STATE_PLAYING:
		return g.updatePlaying()
	case STATE_DEAD:
		return g.updateDead()
	}

	return nil
}

func (g *Game) updateMenu() error {
	// Handle text input
	inputChars := ebiten.AppendInputChars(nil)
	g.gameManager.inputName += string(inputChars)

	if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) && len(g.gameManager.inputName) > 0 {
		g.gameManager.inputName = g.gameManager.inputName[:len(g.gameManager.inputName)-1]
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) && len(g.gameManager.inputName) > 0 {
		g.gameManager.enterGame(g.gameManager.inputName)
	}

	return nil
}

func (g *Game) updatePlaying() error {
	// Handle input
	g.playerController.HandleInput()

	// Update entity interpolation
	g.updateEntityInterpolation()

	// Update camera
	g.cameraController.Update(g.gameManager, g.deltaTime)

	// Update leaderboard
	g.gameManager.leaderboard.Update(g.gameManager)

	return nil
}

func (g *Game) updateEntityInterpolation() {
	g.gameManager.entitiesMutex.RLock()
	defer g.gameManager.entitiesMutex.RUnlock()

	// Update interpolation for all entities
	for _, entity := range g.gameManager.entities {
		entity.updateInterpolation(g.deltaTime)
	}
}

func (g *Game) updateDead() error {
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.gameManager.respawn()
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	switch g.gameManager.state {
	case STATE_MENU:
		g.drawMenu(screen)
	case STATE_CONNECTING:
		g.drawConnecting(screen)
	case STATE_PLAYING:
		g.drawGame(screen)
	case STATE_DEAD:
		g.drawDead(screen)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	g.screenWidth = outsideWidth
	g.screenHeight = outsideHeight

	if g.cameraController.camera != nil {
		g.cameraController.camera.SetSize(float64(outsideWidth), float64(outsideHeight))
	}

	return outsideWidth, outsideHeight
}

// Drawing methods
func (g *Game) drawMenu(screen *ebiten.Image) {
	screen.Fill(color.RGBA{10, 10, 30, 255})

	centerX := g.screenWidth / 2
	centerY := g.screenHeight / 2

	ebitenutil.DebugPrintAt(screen, "BLACKHOLIO", centerX-50, centerY-150)
	ebitenutil.DebugPrintAt(screen, "Space Arena MMORPG", centerX-70, centerY-120)
	ebitenutil.DebugPrintAt(screen, "Enter your cosmic name:", centerX-90, centerY-50)

	nameDisplay := g.gameManager.inputName + "_"
	ebitenutil.DebugPrintAt(screen, nameDisplay, centerX-len(nameDisplay)*4, centerY-20)

	ebitenutil.DebugPrintAt(screen, "Press ENTER to begin", centerX-75, centerY+50)

	if g.gameManager.connectError != "" {
		ebitenutil.DebugPrintAt(screen, g.gameManager.connectError, 50, centerY+150)
	}
}

func (g *Game) drawConnecting(screen *ebiten.Image) {
	screen.Fill(color.RGBA{10, 10, 30, 255})
	centerX := g.screenWidth / 2
	centerY := g.screenHeight / 2
	ebitenutil.DebugPrintAt(screen, "Connecting to the cosmic arena...", centerX-120, centerY)
}

func (g *Game) drawGame(screen *ebiten.Image) {
	screen.Fill(color.RGBA{5, 5, 15, 255})

	g.drawGrid(screen)
	g.drawEntities(screen)
	g.drawCircles(screen)
	g.drawUI(screen)
	g.drawLeaderboard(screen)
}

func (g *Game) drawDead(screen *ebiten.Image) {
	screen.Fill(color.RGBA{20, 5, 5, 255})
	centerX := g.screenWidth / 2
	centerY := g.screenHeight / 2
	ebitenutil.DebugPrintAt(screen, "Your cosmic journey has ended...", centerX-120, centerY-20)
	ebitenutil.DebugPrintAt(screen, "Press SPACE to be reborn", centerX-100, centerY+10)
}

func (g *Game) drawGrid(screen *ebiten.Image) {
	if g.cameraController.camera == nil {
		return
	}

	gridSize := 100.0
	camera := g.cameraController.camera

	// Get camera bounds
	camX, camY := camera.Center()
	halfWidth := float64(g.screenWidth) / 2
	halfHeight := float64(g.screenHeight) / 2

	// Calculate visible range
	startX := math.Floor((camX-halfWidth)/gridSize) * gridSize
	endX := math.Ceil((camX+halfWidth)/gridSize) * gridSize
	startY := math.Floor((camY-halfHeight)/gridSize) * gridSize
	endY := math.Ceil((camY+halfHeight)/gridSize) * gridSize

	// Draw lines
	for x := startX; x <= endX; x += gridSize {
		sx1, sy1 := camera.ApplyCameraTransformToPoint(x, startY)
		sx2, sy2 := camera.ApplyCameraTransformToPoint(x, endY)
		vector.StrokeLine(screen, float32(sx1), float32(sy1), float32(sx2), float32(sy2), 1, color.RGBA{20, 20, 40, 100}, false)
	}

	for y := startY; y <= endY; y += gridSize {
		sx1, sy1 := camera.ApplyCameraTransformToPoint(startX, y)
		sx2, sy2 := camera.ApplyCameraTransformToPoint(endX, y)
		vector.StrokeLine(screen, float32(sx1), float32(sy1), float32(sx2), float32(sy2), 1, color.RGBA{20, 20, 40, 100}, false)
	}
}

func (g *Game) drawUI(screen *ebiten.Image) {
	// Game controls
	ebitenutil.DebugPrintAt(screen, "Move: Mouse | Split: SPACE | Suicide: S", 20, g.screenHeight-60)

	// Draw player UI (mass display)
	g.drawPlayerUI(screen)
}

func (g *Game) isOnScreen(x, y float64) bool {
	margin := 100.0
	return x >= -margin && x <= float64(g.screenWidth)+margin &&
		y >= -margin && y <= float64(g.screenHeight)+margin
}

// GameManager methods (connection and entity management)
func (gm *GameManager) ConnectAndRun() {
	defer func() {
		if gm.wsConn != nil {
			gm.wsConn.GracefulClose()
		}
	}()

	log.Printf("Starting connection process...")

	var err error
	if gm.conn, err = gm.connectToDB(); err != nil {
		gm.connectError = fmt.Sprintf("Failed to connect: %v", err)
		gm.state = STATE_MENU
		return
	}
	defer gm.conn.Close()

	if err := gm.setupIdentity(); err != nil {
		gm.connectError = fmt.Sprintf("Identity setup failed: %v", err)
		gm.state = STATE_MENU
		return
	}

	if gm.wsConn, err = gm.conn.Database.ConnectWebSocket(DB_NAME, client.SatsProtocol); err != nil {
		gm.connectError = fmt.Sprintf("Failed to connect WebSocket: %v", err)
		gm.state = STATE_MENU
		return
	}

	subscriptionID := uuid.New().ID()
	if err := gm.wsConn.SendSubscribeAll(subscriptionID); err != nil {
		gm.connectError = fmt.Sprintf("Failed to subscribe: %v", err)
		gm.state = STATE_MENU
		return
	}

	go gm.handleWebSocketMessages()
	log.Printf("Connection established successfully")
	<-gm.ctx.Done()
}

func (gm *GameManager) connectToDB() (*client.Client, error) {
	log.Printf("Creating database client for host: %s", HOST)
	builder := client.NewClientBuilder().
		WithBaseURL(HOST).
		WithTimeout(10 * time.Second)

	return builder.Build()
}

func (gm *GameManager) setupIdentity() error {

	//log.Printf("Using existing identity: %s", gm.localIdentity)
	return gm.createIdentity()
}

func (gm *GameManager) createIdentity() error {
	log.Printf("Creating new identity...")
	resp, err := gm.conn.Identity.Create()
	if err != nil {
		log.Printf("ERROR: Failed to create identity: %v", err)
		return err
	}

	gm.localIdentity = resp.Identity
	if gm.authToken != nil {
		gm.authToken.SaveToken(resp.Token)
	}
	gm.conn.SetToken(resp.Token)
	gm.conn.SetIdentity(resp.Identity)

	log.Printf("Created new identity: %s", resp.Identity)
	return nil
}

func (gm *GameManager) enterGame(name string) {
	log.Printf("=== ENTER GAME START ===")
	log.Printf("Entering game with name: %s", name)
	gm.playerName = name
	gm.state = STATE_CONNECTING
	log.Printf("State changed to CONNECTING (%d)", int(gm.state))

	args := fmt.Sprintf(`[%q]`, name)
	if err := gm.wsConn.SendCallReducer("EnterGame", args, 0); err != nil {
		log.Printf("ERROR: Failed to send EnterGame: %v", err)
		gm.connectError = fmt.Sprintf("Failed to enter game: %v", err)
		gm.state = STATE_MENU
	} else {
		log.Printf("EnterGame reducer sent successfully, waiting for response...")
	}
}

func (gm *GameManager) getCenterOfMass() *Vector2 {
	if gm.localPlayer == nil || len(gm.localPlayer.OwnedCircles) == 0 {
		return nil
	}

	gm.entitiesMutex.RLock()
	defer gm.entitiesMutex.RUnlock()

	var totalPos Vector2
	var totalMass uint
	var validCircles int

	for _, entityID := range gm.localPlayer.OwnedCircles {
		if entity, exists := gm.entities[entityID]; exists {
			// Use CurrentPosition for smooth camera following
			totalPos.X += entity.CurrentPosition.X * float64(entity.Mass)
			totalPos.Y += entity.CurrentPosition.Y * float64(entity.Mass)
			totalMass += entity.Mass
			validCircles++

			// Log if any circle is at (0,0) - indicates teleportation bug
			if entity.CurrentPosition.X == 0.0 && entity.CurrentPosition.Y == 0.0 {
				log.Printf("CRITICAL: Circle %d is at (0,0)! Mass:%d", entityID, entity.Mass)
			}
		}
	}

	if totalMass == 0 {
		return nil
	}

	centerOfMass := &Vector2{
		X: totalPos.X / float64(totalMass),
		Y: totalPos.Y / float64(totalMass),
	}

	return centerOfMass
}

// WebSocket message handling (simplified)
func (gm *GameManager) handleWebSocketMessages() {
	log.Printf("WebSocket message handler started")
	for {
		select {
		case <-gm.ctx.Done():
			log.Printf("WebSocket message handler stopping")
			return
		default:
			msg, err := gm.wsConn.ReceiveMessage()
			if err != nil {
				log.Printf("WebSocket receive error: %v", err)
				return
			}
			gm.processMessage(msg)
		}
	}
}

func (gm *GameManager) processMessage(msg any) {
	msgBytes, _ := json.Marshal(msg)
	serverMsg, err := client.ParseServerMessage(msgBytes)
	if err != nil {
		log.Printf("ERROR: Failed to parse server message: %v", err)
		return
	}

	if txUpdate, ok := serverMsg.AsTransactionUpdate(); ok {
		gm.handleTransactionUpdate(txUpdate)
	} else if initSub, ok := serverMsg.AsInitialSubscription(); ok {
		log.Printf("Received initial subscription")
		gm.handleInitialSubscription(initSub)
	}
}

func (gm *GameManager) handleInitialSubscription(initSub *client.InitialSubscription) {
	for _, table := range initSub.DatabaseUpdate.Tables {
		switch table.TableName {
		case "entity":
			gm.processInitialEntities(table.Updates)
		case "circle":
			gm.processInitialCircles(table.Updates)
		case "player":
			gm.processInitialPlayers(table.Updates)
		case "food":
			gm.processInitialFoods(table.Updates)
		}
	}
	gm.findLocalPlayer()

	// Force state to PLAYING after initial subscription
	if gm.state == STATE_CONNECTING {
		log.Printf("Initial subscription complete - changing to PLAYING state")
		gm.state = STATE_PLAYING
	}
}

func (gm *GameManager) handleTransactionUpdate(txUpdate *client.TransactionUpdate) {
	if txUpdate.Status.Committed != nil {
		for _, tableUpdate := range txUpdate.Status.Committed.Tables {
			for _, update := range tableUpdate.Updates {
				gm.processTableUpdate(tableUpdate.TableName, update)
			}
		}

		switch txUpdate.ReducerCall.ReducerName {
		case "EnterGame":
			log.Printf("EnterGame completed - changing to PLAYING state")
			if gm.state == STATE_CONNECTING {
				gm.state = STATE_PLAYING
				gm.findLocalPlayer()
			}
		}
	} else if txUpdate.Status.Failed != nil {
		log.Printf("ERROR: Transaction FAILED for %s: %s", txUpdate.ReducerCall.ReducerName, *txUpdate.Status.Failed)

		// If EnterGame fails, still try to continue
		if txUpdate.ReducerCall.ReducerName == "EnterGame" && gm.state == STATE_CONNECTING {
			log.Printf("EnterGame failed but forcing state to PLAYING anyway")
			gm.state = STATE_PLAYING
		}
	}
}

func (gm *GameManager) processTableUpdate(tableName string, update client.TableUpdateEntry) {
	switch tableName {
	case "entity":
		gm.processEntityUpdates(update)
	case "circle":
		gm.processCircleUpdates(update)
	case "player":
		gm.processPlayerUpdates(update)
	case "food":
		gm.processFoodUpdates(update)
	}
}
