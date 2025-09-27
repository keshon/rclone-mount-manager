// Package ui provides the graphical user interface for the rclone mount manager
// application using ImGui. It handles window management, DPI detection and scaling,
// font loading, remote configuration display, mount controls and status information.
package ui

import (
	"fmt"

	"github.com/AllenDang/cimgui-go/backend"
	"github.com/AllenDang/cimgui-go/backend/glfwbackend"
	"github.com/AllenDang/cimgui-go/imgui"
)

// Backend and scaling configuration.
var (
	currentBackend backend.Backend[glfwbackend.GLFWWindowFlags]        // active GLFW backend
	baseFontSize   float32                                      = 16.0 // base font size
	dpiScale       float32                                      = 1.0  // current DPI scale
)

// InitBackend initializes the GLFW backend and creates the main window.
// It applies DPI-aware styling, sets window size limits, and installs
// file drop and close callbacks.
func InitBackend(windowTitle string, width, height int) {
	currentBackend, _ = backend.CreateBackend(glfwbackend.NewGLFWBackend())
	currentBackend.SetBgColor(imgui.NewVec4(0.1, 0.12, 0.15, 1.0))
	currentBackend.CreateWindow(windowTitle, width, height)
	currentBackend.SetWindowSizeLimits(600, 800, -1, -1)
	currentBackend.SetDropCallback(OnFileDrop)
	currentBackend.SetCloseCallback(OnClose)

	dpiScale = DetectDPIScale()
	fmt.Printf("Final DPI scale: %.2f\n", dpiScale)

	ApplyGUIStyles(baseFontSize, dpiScale)
}

// UpdateDPIScale detects DPI changes and reapplies styling if needed.
func UpdateDPIScale() {
	newScale := DetectDPIScale()
	if newScale != dpiScale {
		fmt.Printf("DPI scale changed from %.2f to %.2f\n", dpiScale, newScale)
		dpiScale = newScale
		ApplyGUIStyles(baseFontSize, dpiScale)
	}
}
