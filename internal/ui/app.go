// Package ui provides the graphical user interface for the rclone mount manager
// application using ImGui. It handles window management, DPI detection and scaling,
// font loading, remote configuration display, mount controls and status information.
package ui

import (
	"context"
	"fmt"
	"rclone-gui/internal/rclone"
	"rclone-gui/internal/version"
	"time"

	"github.com/AllenDang/cimgui-go/imgui"
)

// ActiveFileProgress holds transfer progress for a single file.
type ActiveFileProgress struct {
	Name       string  // File path
	Icon       string  // File type icon
	Percentage float64 // Completion percentage [0–100]
	Speed      float64 // Bytes per second
	ETA        int64   // Seconds remaining
}

// ActiveMount represents an active rclone mount.
type ActiveMount struct {
	Mount   rclone.Mount         // Mount configuration
	CmdCtx  context.CancelFunc   // Cancel function
	Mounted bool                 // Whether the mount is active
	Files   []ActiveFileProgress // Current file transfers
}

// ActiveMounts tracks active mounts by remote config name.
var ActiveMounts = map[string]*ActiveMount{}

// Global state.
var (
	manager *rclone.Manager // rclone manager
	App     = &AppState{}   // application state
)

// AppState holds the current application state.
type AppState struct {
	SelectedConfig string // Selected remote config
	Status         string // Current status message
	Progress       string // Progress text
	LastAction     string // Most recent action
}

// Init sets up the rclone manager and starts background progress updates.
func Init() error {
	var err error
	manager, err = rclone.NewManager("rclone.conf", "rclone.log")
	if err != nil {
		return fmt.Errorf("failed to initialize rclone manager: %w", err)
	}

	// Background progress monitor.
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			if err := manager.UpdateProgress(); err != nil {
				App.Status = err.Error()
				continue
			}
			if App.SelectedConfig == "" {
				continue
			}
			active, ok := ActiveMounts[App.SelectedConfig]
			if !ok {
				continue
			}
			snap, ok := manager.GetRemoteProgress(App.SelectedConfig)
			if !ok {
				active.Files = nil
				continue
			}
			files := make([]ActiveFileProgress, 0, len(snap.Transferring))
			for _, t := range snap.Transferring {
				files = append(files, ActiveFileProgress{
					Name:       t.Name,
					Icon:       "📄",
					Percentage: t.Percentage,
					Speed:      t.Speed,
					ETA:        t.ETA,
				})
			}
			active.Files = files
		}
	}()

	return nil
}

// Shutdown unmounts all active mounts and stops the rclone daemon.
func Shutdown() {
	for name := range ActiveMounts {
		_ = manager.Unmount(name)
		delete(ActiveMounts, name)
	}
	if manager != nil {
		time.Sleep(200 * time.Millisecond)
		manager.StopRclone()
	}
}

// BeforeDestroyContext is called before the ImGui context is destroyed.
func BeforeDestroyContext() {
	Shutdown()
}

// AfterCreateContext is called after the ImGui context is created.
func AfterCreateContext() {
	// Initialize resources here.
}

// OnFileDrop handles file drag-and-drop events.
func OnFileDrop(paths []string) {
	fmt.Printf("Files dropped: %v\n", paths)
}

// OnClose handles window close events.
func OnClose() {
	fmt.Println("Window closed")
}

// StartGUI initializes the backend and runs the main loop.
func StartGUI() {
	InitBackend(version.AppName+" - "+version.GoVersion, 600, 800)
	currentBackend.SetAfterCreateContextHook(AfterCreateContext)
	currentBackend.SetBeforeDestroyContextHook(BeforeDestroyContext)
	currentBackend.Run(Loop)
}

// Loop renders the UI each frame.
func Loop() {
	RenderUI()
	imgui.Render()
}
