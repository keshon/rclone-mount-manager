package main

import (
	"fmt"
	"rclone-gui/internal/ui"
	"runtime"

	"golang.org/x/sys/windows"
)

func init() {
	shcore := windows.NewLazySystemDLL("Shcore.dll")
	setDpi := shcore.NewProc("SetProcessDpiAwareness")
	if setDpi.Find() == nil {
		// 2 = PROCESS_PER_MONITOR_DPI_AWARE
		setDpi.Call(2)
	} else {
		// Fallback for older Windows
		user32 := windows.NewLazySystemDLL("user32.dll")
		setDpiAware := user32.NewProc("SetProcessDPIAware")
		if setDpiAware.Find() == nil {
			setDpiAware.Call()
		}
	}
	runtime.LockOSThread()
}

func main() {
	if err := ui.Init(); err != nil {
		fmt.Printf("Failed to initialize UI: %v\n", err)
		return
	}
	defer ui.Shutdown()

	ui.StartGUI()
}
