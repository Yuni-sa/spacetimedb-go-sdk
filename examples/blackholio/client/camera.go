package main

import (
	"math"

	"github.com/setanarut/kamera/v2"
)

// Camera constants
const (
	CAMERA_BASE_SIZE      = 1.1
	CAMERA_MASS_ZOOM_MULT = 0.005
	CAMERA_SPLIT_ZOOM_ADD = 1.0
	CAMERA_LERP_SPEED     = 10
)

// CameraController handles camera
type CameraController struct {
	camera     *kamera.Camera
	targetSize float64
}

func NewCameraController() *CameraController {
	camera := kamera.NewCamera(WORLD_SIZE/2, WORLD_SIZE/2, DEFAULT_SCREEN_WIDTH, DEFAULT_SCREEN_HEIGHT)
	camera.SmoothType = kamera.Lerp
	camera.SmoothOptions.SmoothDampTimeX = 0.3
	camera.SmoothOptions.SmoothDampTimeY = 0.3

	return &CameraController{
		camera:     camera,
		targetSize: CAMERA_BASE_SIZE,
	}
}

// CameraController methods
func (cc *CameraController) Update(gm *GameManager, deltaTime float64) {
	if cc.camera == nil {
		return
	}

	// Follow player
	centerOfMass := gm.getCenterOfMass()
	if centerOfMass != nil {
		cc.camera.LookAt(centerOfMass.X, centerOfMass.Y)
	} else {
		// No player - stay at world center
		cc.camera.LookAt(WORLD_SIZE/2, WORLD_SIZE/2)
	}

	// Update zoom based on mass
	if gm.localPlayer != nil {
		totalMass := gm.getTotalMass()
		numCircles := len(gm.localPlayer.OwnedCircles)

		targetSize := CAMERA_BASE_SIZE +
			math.Min(5, float64(totalMass)*CAMERA_MASS_ZOOM_MULT) +
			math.Min(float64(numCircles-1), 1.0)*CAMERA_SPLIT_ZOOM_ADD

		// Smooth camera size transition
		sizeDiff := targetSize - cc.targetSize
		cc.targetSize += sizeDiff * deltaTime * CAMERA_LERP_SPEED

		// Apply camera size - higher zoom factor means more zoomed in
		baseZoom := 20.0
		cc.camera.ZoomFactor = baseZoom / cc.targetSize
	} else {
		// Default zoom when no player
		cc.camera.ZoomFactor = 15.0
	}
}
