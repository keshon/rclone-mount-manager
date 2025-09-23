package ui

import (
	"context"
	"fmt"
	"path/filepath"
	"rclone-gui/internal/rclone"
	"sort"
	"strings"

	"github.com/AllenDang/cimgui-go/imgui"
)

var (
	searchFilter string
)

func RenderUI() {
	mainViewport := imgui.MainViewport()
	imgui.SetNextWindowPos(mainViewport.Pos())
	imgui.SetNextWindowSize(mainViewport.Size())

	flags := imgui.WindowFlagsNoTitleBar |
		imgui.WindowFlagsNoResize |
		imgui.WindowFlagsNoMove |
		imgui.WindowFlagsNoCollapse |
		imgui.WindowFlagsNoScrollbar

	imgui.BeginV("RCLONE MOUNT MANAGER", nil, flags)
	imgui.PushStyleVarVec2(imgui.StyleVarWindowPadding, imgui.Vec2{X: 20, Y: 20})

	style := imgui.CurrentStyle()
	footerHeight := imgui.FrameHeightWithSpacing() + style.ItemSpacing().Y

	contentHeight := imgui.WindowHeight() - footerHeight - style.WindowPadding().Y*2
	imgui.BeginChildStrV("MainContent", imgui.Vec2{-1, contentHeight}, imgui.ChildFlagsNone, 0)

	renderHeader()
	imgui.Spacing()
	renderConfigSection()
	imgui.Spacing()
	renderStatusSection()

	imgui.EndChild()

	imgui.Separator()
	renderFooter()

	imgui.PopStyleVar()
	imgui.End()
}

func renderHeader() {
	titleColor := imgui.Vec4{X: 0.4, Y: 0.8, Z: 1.0, W: 1.0}
	imgui.PushStyleColorVec4(imgui.ColText, titleColor)
	imgui.PushFont(mainFont)
	imgui.Text("🗂 Rclone Mount Manager")
	imgui.PopFont()
	imgui.PopStyleColor()

	imgui.Separator()
}

func renderFooter() {
	statusColor := imgui.Vec4{X: 0.2, Y: 0.8, Z: 0.2, W: 1.0}
	statusText := fmt.Sprintf("%d Active", len(ActiveMounts))
	if len(ActiveMounts) == 0 {
		statusText = "No active mounts"
		statusColor = imgui.Vec4{X: 0.7, Y: 0.7, Z: 0.7, W: 1.0}
	}

	imgui.PushStyleColorVec4(imgui.ColText, statusColor)
	imgui.Text(statusText)
	imgui.PopStyleColor()

	imgui.SameLine()

	configButtonText := "Rclone Config"
	configButtonSize := imgui.CalcTextSizeV(configButtonText, false, 0)

	style := imgui.CurrentStyle()
	buttonPadding := style.FramePadding().X * 2

	imgui.SetCursorPosX(imgui.WindowWidth() - configButtonSize.X - buttonPadding - style.WindowPadding().X)

	if imgui.Button(configButtonText) {
		// Launch Rclone config window
		err := manager.ConfigRemote()
		if err != nil {
			App.Status = err.Error()
		} else {
			App.Status = "Rclone config opened"
		}
	}
}

func renderConfigSection() {
	imgui.PushStyleColorVec4(imgui.ColText, imgui.Vec4{X: 0.9, Y: 0.9, Z: 0.9, W: 1.0})
	imgui.Text("📡 Remote Configuration")
	imgui.PopStyleColor()

	renderMountControls()

	imgui.SameLine()
	imgui.SetNextItemWidth(-1)

	filteredConfigs := getFilteredConfigs()
	sort.Slice(filteredConfigs, func(i, j int) bool {
		return filteredConfigs[i].Name < filteredConfigs[j].Name
	})

	currentDisplayName := App.SelectedConfig
	if App.SelectedConfig == "" {
		currentDisplayName = "Select a remote..."
	} else {
		currentDisplayName = fmt.Sprintf("%s %s", getRemoteStatusEmoji(App.SelectedConfig), App.SelectedConfig)
	}

	if imgui.BeginCombo("##config", currentDisplayName) {
		if len(filteredConfigs) == 0 {
			imgui.PushStyleColorVec4(imgui.ColText, imgui.Vec4{X: 0.5, Y: 0.5, Z: 0.5, W: 1.0})
			imgui.Text("No configs found")
			imgui.PopStyleColor()
		} else {
			for _, cfg := range filteredConfigs {
				isSelected := cfg.Name == App.SelectedConfig
				emoji, col := getRemoteStatusEmoji(cfg.Name), getRemoteStatusColor(cfg.Name)

				imgui.PushStyleColorVec4(imgui.ColText, col)
				if imgui.SelectableBoolV(fmt.Sprintf("%s %s", emoji, cfg.Name), isSelected, imgui.SelectableFlagsNone, imgui.Vec2{}) {
					App.SelectedConfig = cfg.Name
				}
				imgui.PopStyleColor()

				if isSelected {
					imgui.SetItemDefaultFocus()
				}
			}
		}
		imgui.EndCombo()
	}
}

func getRemoteStatusEmoji(name string) string {
	if active, ok := ActiveMounts[name]; ok {
		if active.Progress == "" {
			return "🔄" // в процессе
		}
		return "✅" // подключен
	}
	return "💤" // не подключен
}

func getRemoteStatusColor(name string) imgui.Vec4 {
	if active, ok := ActiveMounts[name]; ok {
		if active.Progress == "" {
			return imgui.Vec4{X: 0.95, Y: 0.85, Z: 0.3, W: 1.0} // жёлтый
		}
		return imgui.Vec4{X: 0.2, Y: 0.9, Z: 0.2, W: 1.0} // зелёный
	}
	return imgui.Vec4{X: 0.8, Y: 0.8, Z: 0.8, W: 1.0} // серый
}

func renderMountControls() {
	isMounted := false
	if App.SelectedConfig != "" {
		if _, ok := ActiveMounts[App.SelectedConfig]; ok {
			isMounted = true
		}
	}

	imgui.BeginDisabledV(App.SelectedConfig == "")

	if isMounted {
		if imgui.Button("Unmount Remote") {
			unmountSelected(App.SelectedConfig)
		}
	} else {
		if imgui.Button("Mount Remote") {
			go mountSelected(App.SelectedConfig)
		}
	}

	imgui.EndDisabled()
}

func renderStatusSection() {
	imgui.PushStyleColorVec4(imgui.ColText, imgui.Vec4{X: 0.9, Y: 0.9, Z: 0.9, W: 1.0})
	imgui.Text("📋 Status")
	imgui.PopStyleColor()

	statusColor := imgui.Vec4{X: 0.7, Y: 0.7, Z: 0.7, W: 1.0}
	statusText := "Ready"

	if App.SelectedConfig != "" {
		if active, ok := ActiveMounts[App.SelectedConfig]; ok {
			if active.Progress == "" {
				statusColor = imgui.Vec4{X: 0.9, Y: 0.7, Z: 0.3, W: 1.0}
				statusText = "Mounting..."
			} else {
				statusColor = imgui.Vec4{X: 0.2, Y: 0.8, Z: 0.2, W: 1.0}
				statusText = "\\\\cloud\\" + App.SelectedConfig
			}
		} else {
			statusText = "Ready"
			statusColor = imgui.Vec4{X: 0.7, Y: 0.7, Z: 0.7, W: 1.0}
		}
	}

	imgui.PushStyleColorVec4(imgui.ColText, statusColor)
	imgui.TextWrapped(statusText)
	imgui.PopStyleColor()

	if App.SelectedConfig != "" {
		if active, ok := ActiveMounts[App.SelectedConfig]; ok {
			if active.Progress != "" {
				imgui.PushStyleColorVec4(imgui.ColText, imgui.Vec4{X: 0.6, Y: 0.8, Z: 1.0, W: 1.0})
				imgui.TextWrapped(active.Progress)
				imgui.PopStyleColor()
			}
		}
	}
}

func getFilteredConfigs() []rclone.Remote {
	if searchFilter == "" {
		return manager.Remotes
	}

	var filtered []rclone.Remote
	searchLower := strings.ToLower(searchFilter)
	for _, cfg := range manager.Remotes {
		if strings.Contains(strings.ToLower(cfg.Name), searchLower) {
			filtered = append(filtered, cfg)
		}
	}
	return filtered
}

func unmountSelected(configName string) {
	if _, ok := ActiveMounts[configName]; ok {
		err := manager.Unmount(configName)
		if err != nil {
			App.Status = "❎ Error: " + err.Error()
		} else {
			delete(ActiveMounts, configName)
			App.Status = "✅ Unmounted successfully"
		}
	}
}

func mountSelected(cfgName string) {
	var cfg *rclone.Remote
	for _, r := range manager.Remotes {
		if r.Name == cfgName {
			cfg = &r
			break
		}
	}
	if cfg == nil {
		App.Status = "❎ Configuration not found"
		return
	}

	mount := rclone.Mount{
		Name:       cfg.Name,
		MountPoint: "\\\\cloud\\" + cfg.Name,
		CacheMode:  "writes",
		CacheSize:  "500M",
		CacheAge:   "1m",
		LogPath:    manager.ConfigPath,
		CacheDir:   filepath.Join("C:/temp/rclone", cfg.Name),
	}

	ctx, cancel := context.WithCancel(context.Background())
	progressChan := make(chan string, 100)

	active := &ActiveMount{
		Mount:    mount,
		CmdCtx:   cancel,
		Progress: "",
	}
	ActiveMounts[cfg.Name] = active

	go func() {
		for msg := range progressChan {
			active.Progress = msg
		}
	}()

	go func() {
		err := manager.Mount(ctx, mount, progressChan)
		if err != nil {
			App.Status = err.Error()
			delete(ActiveMounts, cfg.Name)
		} else {
			App.Status = fmt.Sprintf("✅ %s mounted successfully", cfg.Name)
			active.Progress = ""
		}
	}()
}
