// Package ui provides the graphical user interface for the rclone mount manager
// application using ImGui. It handles window management, DPI detection and scaling,
// font loading, remote configuration display, mount controls and status information.
package ui

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"rclone-gui/internal/rclone"
	"rclone-gui/internal/version"
	"runtime"
	"sort"
	"strings"

	"github.com/AllenDang/cimgui-go/imgui"
)

// searchFilter is the current text filter for remote configurations.
var searchFilter string

// RenderUI draws the main window with header, configuration, status, and footer.
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
	numFooterLines := 1
	footerHeight := float32(numFooterLines)*imgui.FrameHeightWithSpacing() + style.ItemSpacing().Y
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

// renderHeader shows the app title, description, and repository link.
func renderHeader() {
	// Title
	imgui.PushStyleVarVec2(imgui.StyleVarItemSpacing, imgui.Vec2{X: 0, Y: baseFontSize})
	titleColor := imgui.Vec4{X: 1.0, Y: 1.0, Z: 1.0, W: 1.0}
	imgui.PushStyleColorVec4(imgui.ColText, titleColor)
	imgui.PushFont(boldFont, dpifontSize*1.2)
	imgui.Text(strings.ToUpper(version.AppName))
	imgui.PopFont()
	imgui.PopStyleColor()
	imgui.PopStyleVar()

	imgui.PushStyleVarVec2(imgui.StyleVarItemSpacing, imgui.Vec2{X: 0, Y: 2})

	// Description
	imgui.PushStyleColorVec4(imgui.ColText, imgui.Vec4{X: 0.5, Y: 0.5, Z: 0.5, W: 1.0})
	imgui.PushFont(mainFont, dpifontSize)
	imgui.Text(version.AppDescription)
	imgui.PopFont()
	imgui.PopStyleColor()

	// Repo link
	imgui.PushStyleColorVec4(imgui.ColText, imgui.Vec4{X: 0.4, Y: 0.8, Z: 1.0, W: 1.0})
	imgui.PushFont(mainFont, dpifontSize)
	toggleLink := false
	if imgui.SelectableBoolPtr(version.AppRepo, &toggleLink) {
		openURL("https://" + version.AppRepo)
		toggleLink = false
	}
	imgui.PopFont()
	imgui.PopStyleColor()

	imgui.PopStyleVar()

	// Spacer
	imgui.PushStyleVarVec2(imgui.StyleVarItemSpacing, imgui.Vec2{X: 0, Y: -6})
	imgui.Text("")
	imgui.PopStyleVar()

	imgui.Separator()
}

// renderConfigSection shows the remote selector and mount controls.
func renderConfigSection() {
	imgui.PushStyleColorVec4(imgui.ColText, imgui.Vec4{X: 0.9, Y: 0.9, Z: 0.9, W: 1.0})
	imgui.Text("📂 Remote Configurations")
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

// renderMountControls shows mount/unmount buttons for the selected remote.
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

// renderStatusSection shows mount status and transfer progress.
func renderStatusSection() {
	imgui.PushStyleColorVec4(imgui.ColText, imgui.Vec4{X: 0.9, Y: 0.9, Z: 0.9, W: 1.0})
	imgui.Text("📊 Status")
	imgui.PopStyleColor()

	if App.SelectedConfig == "" {
		return
	}
	active, ok := ActiveMounts[App.SelectedConfig]
	if !ok {
		return
	}

	// Mount status
	statusColor := imgui.Vec4{X: 0.7, Y: 0.7, Z: 0.7, W: 1.0}
	statusText := "Ready"
	if active.Mounted {
		statusColor = imgui.Vec4{X: 0.2, Y: 0.8, Z: 0.2, W: 1.0}
		statusText = "\\\\cloud\\" + App.SelectedConfig
	} else {
		statusColor = imgui.Vec4{X: 0.9, Y: 0.7, Z: 0.3, W: 1.0}
		statusText = "Mounting..."
	}
	imgui.PushStyleColorVec4(imgui.ColText, statusColor)
	imgui.TextWrapped(statusText)
	imgui.PopStyleColor()

	// Active file transfers
	if len(active.Files) == 0 {
		return
	}
	for _, f := range active.Files {
		imgui.PushStyleColorVec4(imgui.ColText, imgui.Vec4{X: 0.6, Y: 0.8, Z: 1.0, W: 1.0})
		progressText := fmt.Sprintf("%s %s - %.1f%%, %.2f MB/s, ETA %ds",
			f.Icon,
			filepath.Base(f.Name),
			f.Percentage,
			f.Speed/1024/1024,
			f.ETA,
		)
		imgui.TextWrapped(progressText)
		imgui.PopStyleColor()
	}
}

// renderFooter shows the number of active mounts and a button to open config.
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

	if imgui.SmallButton(configButtonText) {
		if err := manager.ConfigRemote(); err != nil {
			App.Status = err.Error()
		} else {
			App.Status = "Rclone config opened"
		}
	}
}

// getRemoteStatusEmoji returns an emoji for the remote's current status.
func getRemoteStatusEmoji(name string) string {
	if active, ok := ActiveMounts[name]; ok {
		if len(active.Files) > 0 {
			return "🔄"
		}
		if active.Mounted {
			return "✅"
		}
	}
	return "💤"
}

// getRemoteStatusColor returns a color for the remote's current status.
func getRemoteStatusColor(name string) imgui.Vec4 {
	if active, ok := ActiveMounts[name]; ok {
		if len(active.Files) > 0 {
			return imgui.Vec4{X: 0.95, Y: 0.85, Z: 0.3, W: 1.0}
		}
		if active.Mounted {
			return imgui.Vec4{X: 0.2, Y: 0.9, Z: 0.2, W: 1.0}
		}
	}
	return imgui.Vec4{X: 0.8, Y: 0.8, Z: 0.8, W: 1.0}
}

// getFilteredConfigs returns remotes filtered by searchFilter.
func getFilteredConfigs() []*rclone.Remote {
	if searchFilter == "" {
		remotes := make([]*rclone.Remote, len(manager.Remotes))
		for i := range manager.Remotes {
			remotes[i] = &manager.Remotes[i]
		}
		return remotes
	}
	var filtered []*rclone.Remote
	searchLower := strings.ToLower(searchFilter)
	for i := range manager.Remotes {
		cfg := &manager.Remotes[i]
		if strings.Contains(strings.ToLower(cfg.Name), searchLower) {
			filtered = append(filtered, cfg)
		}
	}
	return filtered
}

// unmountSelected unmounts the given remote and updates state.
func unmountSelected(configName string) {
	if _, ok := ActiveMounts[configName]; ok {
		err := manager.Unmount(configName)
		if err != nil {
			App.Status = "Error: " + err.Error()
		} else {
			delete(ActiveMounts, configName)
			App.Status = "Unmounted successfully"
		}
	}
}

// mountSelected mounts the given remote in a background goroutine.
func mountSelected(cfgName string) {
	var cfg *rclone.Remote
	for i := range manager.Remotes {
		if manager.Remotes[i].Name == cfgName {
			cfg = &manager.Remotes[i]
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
		CacheDir:   filepath.Join("C:/temp/rclone", cfg.Name),
	}
	ctx, cancel := context.WithCancel(context.Background())
	active := &ActiveMount{
		Mount:   mount,
		CmdCtx:  cancel,
		Files:   []ActiveFileProgress{},
		Mounted: false,
	}
	ActiveMounts[cfg.Name] = active

	go func() {
		err := manager.Mount(ctx, mount)
		if err != nil {
			App.Status = err.Error()
			delete(ActiveMounts, cfg.Name)
		} else {
			if active, ok := ActiveMounts[cfg.Name]; ok {
				active.Mounted = true
			}
			App.Status = fmt.Sprintf("✅ %s mounted successfully", cfg.Name)
		}
	}()
}

// openURL opens url in the system default browser.
func openURL(url string) {
	switch runtime.GOOS {
	case "windows":
		exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		exec.Command("open", url).Start()
	default:
		exec.Command("xdg-open", url).Start()
	}
}
