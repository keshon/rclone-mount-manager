package main

import (
	"fmt"
	"rclone-gui/internal/ui"
	"runtime"

	"golang.org/x/sys/windows"
)

// init configures Windows DPI awareness and locks the OS thread.
func init() {
	shcore := windows.NewLazySystemDLL("Shcore.dll")
	setDpi := shcore.NewProc("SetProcessDpiAwareness")
	if setDpi.Find() == nil {
		// PROCESS_PER_MONITOR_DPI_AWARE = 2
		setDpi.Call(2)
	} else {
		// Fallback for older Windows versions
		user32 := windows.NewLazySystemDLL("user32.dll")
		setDpiAware := user32.NewProc("SetProcessDPIAware")
		if setDpiAware.Find() == nil {
			setDpiAware.Call()
		}
	}
	runtime.LockOSThread()
}

// main initializes the UI and starts the GUI loop.
func main() {
	if err := ui.Init(); err != nil {
		fmt.Printf("Failed to initialize UI: %v\n", err)
		return
	}
	defer ui.Shutdown()

	ui.StartGUI()
}
