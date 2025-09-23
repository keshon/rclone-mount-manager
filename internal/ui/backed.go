package ui

import (
	"fmt"

	"github.com/AllenDang/cimgui-go/backend"
	"github.com/AllenDang/cimgui-go/backend/glfwbackend"
	"github.com/AllenDang/cimgui-go/imgui"
)

var currentBackend backend.Backend[glfwbackend.GLFWWindowFlags]

var baseFontSize float32 = 16.0
var dpiScale float32 = 1.0

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

func UpdateDPIScale() {
	newScale := DetectDPIScale()
	if newScale != dpiScale {
		fmt.Printf("DPI scale changed from %.2f to %.2f\n", dpiScale, newScale)
		dpiScale = newScale
		ApplyGUIStyles(baseFontSize, dpiScale)
	}
}
