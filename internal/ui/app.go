package ui

import (
	"context"
	"fmt"
	"rclone-gui/internal/rclone"

	"github.com/AllenDang/cimgui-go/imgui"
)

type ActiveMount struct {
	Mount    rclone.Mount
	CmdCtx   context.CancelFunc
	Progress string
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
	manager, err = rclone.NewManager("rclone.conf")
	if err != nil {
		return fmt.Errorf("failed to initialize rclone manager: %w", err)
	}
	return nil
}

func Shutdown() {

}

func AfterCreateContext() {
	// Initialize resources like icons, textures, etc.
}

func BeforeDestroyContext() {
	// Cancel all running mounts
	for _, m := range ActiveMounts {
		m.CmdCtx()
	}
	ActiveMounts = map[string]*ActiveMount{}
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
