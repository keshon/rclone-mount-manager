package ui

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"syscall"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/go-gl/glfw/v3.3/glfw"
)

func DetectDPIScale() float32 {
	var scale float32 = 1.0

	if runtime.GOOS == "windows" {
		if winScale := getWindowsDPIScale(); winScale > 1.0 {
			scale = winScale
			fmt.Printf("Windows DPI scale: %.2f\n", scale)
		}
	}

	if window := getCurrentGLFWWindow(); window != nil {
		contentScaleX, contentScaleY := window.GetContentScale()
		glfwScale := (contentScaleX + contentScaleY) / 2.0
		fmt.Printf("GLFW content scale: X=%.2f, Y=%.2f, avg=%.2f\n", contentScaleX, contentScaleY, glfwScale)
		if glfwScale > scale && glfwScale <= 4.0 {
			scale = glfwScale
		}
	}

	io := imgui.CurrentIO()
	if io != nil {
		fbScale := io.DisplayFramebufferScale()
		if fbScale.X > 0 && fbScale.Y > 0 {
			imguiScale := (fbScale.X + fbScale.Y) / 2.0
			fmt.Printf("ImGui framebuffer scale: avg=%.2f\n", imguiScale)
			if imguiScale > scale {
				scale = imguiScale
			}
		}
	}

	switch runtime.GOOS {
	case "windows":
		if scale <= 1.0 {
			if envScale := getWindowsEnvironmentDPIScale(); envScale > 1.0 {
				scale = envScale
			}
		}
	case "darwin":
		if scale < 1.0 {
			scale = 1.0
		}
	case "linux":
		if scale < 1.0 || scale > 3.0 {
			scale = 1.0
		}
	}

	if scale < 0.5 {
		scale = 0.5
	} else if scale > 4.0 {
		scale = 4.0
	}

	return scale
}

func getCurrentGLFWWindow() *glfw.Window {
	// TODO: expose actual window from backend if needed
	return nil
}

func getWindowsDPIScale() float32 {
	user32 := syscall.NewLazyDLL("user32.dll")
	shcore := syscall.NewLazyDLL("shcore.dll")
	gdi32 := syscall.NewLazyDLL("gdi32.dll")

	getDpiForSystem := user32.NewProc("GetDpiForSystem")
	if getDpiForSystem.Find() == nil {
		if ret, _, _ := getDpiForSystem.Call(); ret != 0 {
			return float32(ret) / 96.0
		}
	}

	getDC := user32.NewProc("GetDC")
	getDeviceCaps := gdi32.NewProc("GetDeviceCaps")
	releaseDC := user32.NewProc("ReleaseDC")
	if getDC.Find() == nil && getDeviceCaps.Find() == nil && releaseDC.Find() == nil {
		if hdc, _, _ := getDC.Call(0); hdc != 0 {
			defer releaseDC.Call(0, hdc)
			const LOGPIXELSX = 88
			if dpi, _, _ := getDeviceCaps.Call(hdc, LOGPIXELSX); dpi != 0 {
				return float32(dpi) / 96.0
			}
		}
	}

	_ = shcore // reserved for GetDpiForMonitor
	return 1.0
}

func getWindowsEnvironmentDPIScale() float32 {
	if scaleEnv := os.Getenv("QT_SCALE_FACTOR"); scaleEnv != "" {
		if scale, err := strconv.ParseFloat(scaleEnv, 32); err == nil {
			return float32(scale)
		}
	}
	if scaleEnv := os.Getenv("GDK_SCALE"); scaleEnv != "" {
		if scale, err := strconv.ParseFloat(scaleEnv, 32); err == nil {
			return float32(scale)
		}
	}
	return 1.0
}
