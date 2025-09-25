package ui

import (
	"context"
	"fmt"
	"rclone-gui/internal/rclone"
	"time"

	"github.com/AllenDang/cimgui-go/imgui"
)

type ActiveFileProgress struct {
	Name       string
	Icon       string
	Percentage float64
	Speed      float64 // bytes/sec
	ETA        int64   // seconds
}

type ActiveMount struct {
	Mount   rclone.Mount
	CmdCtx  context.CancelFunc
	Mounted bool
	Files   []ActiveFileProgress
}

var ActiveMounts = map[string]*ActiveMount{}

var (
	manager *rclone.Manager
	App     = &AppState{}
)

type AppState struct {
	SelectedConfig string
	Status         string
	Progress       string
	LastAction     string
}

func Init() error {
	var err error
	manager, err = rclone.NewManager("rclone.conf", "rclone.log")
	if err != nil {
		return fmt.Errorf("failed to initialize rclone manager: %w", err)
	}

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
				active.Files = nil // очистить список файлов, если прогресса нет
				continue
			}

			files := make([]ActiveFileProgress, 0, len(snap.Transferring))
			for _, t := range snap.Transferring {
				files = append(files, ActiveFileProgress{
					Name:       t.Name,
					Icon:       "📄",
					Percentage: t.Percentage,
					Speed:      t.Speed,
					// ETA:        *t.ETA, // TODO: fix panic here
				})
			}

			active.Files = files
		}
	}()

	return nil
}

func Shutdown() {
	// Размонтируем всё через manager
	for name := range ActiveMounts {
		_ = manager.Unmount(name)
		delete(ActiveMounts, name)
	}

	// Останавливаем rclone rcd процесс
	if manager != nil {
		time.Sleep(200 * time.Millisecond)
		manager.StopRclone()
	}
}

func BeforeDestroyContext() {
	Shutdown()
}

func AfterCreateContext() {
	// Initialize resources like icons, textures, etc.
}

func OnFileDrop(paths []string) {
	fmt.Printf("Files dropped: %v\n", paths)
}

func OnClose() {
	fmt.Println("Window closed")
}

func StartGUI() {
	InitBackend("Rclone Mount Manager", 600, 800)

	currentBackend.SetAfterCreateContextHook(AfterCreateContext)
	currentBackend.SetBeforeDestroyContextHook(BeforeDestroyContext)

	currentBackend.Run(Loop)
}

func Loop() {
	RenderUI()
	imgui.Render()
}
