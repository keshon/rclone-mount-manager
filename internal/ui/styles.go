// Package ui provides the graphical user interface for the rclone mount manager
// application using ImGui. It handles window management, DPI detection and scaling,
// font loading, remote configuration display, mount controls and status information.
package ui

import (
	"fmt"
	"log"
	"os"

	"github.com/AllenDang/cimgui-go/imgui"
)

// Fonts used in the interface.
var (
	mainFont    *imgui.Font // Primary font
	boldFont    *imgui.Font // Bold font for headers
	dpifontSize float32     // DPI-scaled font size
)

// ApplyGUIStyles sets fonts, emoji support, and DPI-aware scaling.
func ApplyGUIStyles(baseFontSize, dpiScale float32) *imgui.Font {
	io := imgui.CurrentIO()
	dpifontSize = baseFontSize * dpiScale
	io.Fonts().Clear()

	// Main font
	mainFontPath := "assets/Noto_Sans/static/NotoSans-Medium.ttf"
	fontConfig := imgui.NewFontConfig()
	if fileExists(mainFontPath) {
		log.Println("Using custom font:", mainFontPath)
		mainFont = io.Fonts().AddFontFromFileTTFV(mainFontPath, dpifontSize, fontConfig, nil)
	} else {
		mainFont = io.Fonts().AddFontDefault()
	}

	// Emoji font (merged)
	emojiPath := "assets/segoe-ui-emoji/seguiemj-1.45-3d.ttf"
	emojiSize := dpifontSize * 0.78
	ranges := getEmojiGlyphRanges()
	emojiConfig := imgui.NewFontConfig()
	emojiConfig.SetMergeMode(true)
	emojiConfig.SetPixelSnapH(true)
	emojiConfig.SetGlyphMinAdvanceX(emojiSize)
	emojiConfig.SetGlyphMaxAdvanceX(emojiSize * 2.0)

	if fileExists(emojiPath) {
		log.Println("Merging emoji font:", emojiPath)
		io.Fonts().AddFontFromFileTTFV(emojiPath, emojiSize, emojiConfig, &ranges[0])
	}

	// Bold font
	boldFontPath := "assets/Noto_Sans/static/NotoSans-Bold.ttf"
	if fileExists(boldFontPath) {
		boldFont = io.Fonts().AddFontFromFileTTFV(boldFontPath, dpifontSize, imgui.NewFontConfig(), nil)
	}

	// Apply style and scaling
	style := imgui.CurrentStyle()
	scaleDimensions(style, dpiScale)
	imgui.StyleColorsDark()

	fmt.Printf("Applied DPI scaling: %.2fx with emoji support\n", dpiScale)

	return mainFont
}

// getEmojiGlyphRanges defines Unicode ranges for common emoji and symbols.
func getEmojiGlyphRanges() []imgui.Wchar {
	return []imgui.Wchar{
		0x1F300, 0x1F5FF, // Symbols & Pictographs
		0x1F600, 0x1F64F, // Emoticons
		0x1F650, 0x1F67F, // Ornamental Dingbats
		0x1F680, 0x1F6FF, // Transport & Map Symbols
		0x1F700, 0x1F77F, // Alchemical Symbols
		0x1F780, 0x1F7FF, // Geometric Shapes Extended
		0x1F800, 0x1F8FF, // Supplemental Arrows-C
		0x1F900, 0x1F9FF, // Supplemental Symbols & Pictographs
		0x1FA00, 0x1FA6F, // Chess Symbols
		0x1FA70, 0x1FAFF, // Symbols & Pictographs Extended-A
		0x2600, 0x26FF, // Misc Symbols
		0x2700, 0x27BF, // Dingbats
		0x1F1E6, 0x1F1FF, // Regional Indicator Symbols (flags)
		0, // Terminator
	}
}

// scaleDimensions applies DPI scaling to ImGui style values.
func scaleDimensions(style *imgui.Style, scale float32) {
	style.SetWindowPadding(imgui.Vec2{X: 15 * scale, Y: 15 * scale})
	style.SetFramePadding(imgui.Vec2{X: 10 * scale, Y: 4 * scale})
	style.SetCellPadding(imgui.Vec2{X: 8 * scale, Y: 4 * scale})
	style.SetItemSpacing(imgui.Vec2{X: 6 * scale, Y: 10 * scale})
	style.SetItemInnerSpacing(imgui.Vec2{X: 6 * scale, Y: 6 * scale})
	style.SetIndentSpacing(20 * scale)

	style.SetScrollbarSize(20 * scale)
	style.SetGrabMinSize(20 * scale)

	style.SetWindowRounding(3 * scale)
	style.SetChildRounding(3 * scale)
	style.SetFrameRounding(3 * scale)
	style.SetPopupRounding(3 * scale)
	style.SetScrollbarRounding(3 * scale)
	style.SetGrabRounding(3 * scale)
	style.SetTabRounding(3 * scale)

	style.SetWindowBorderSize(1 * scale)
	style.SetChildBorderSize(1 * scale)
	style.SetPopupBorderSize(1 * scale)

	style.SetWindowMinSize(imgui.Vec2{X: 100 * scale, Y: 100 * scale})
	style.SetWindowTitleAlign(imgui.Vec2{X: 0.0, Y: 0.5})
	style.SetWindowMenuButtonPosition(imgui.DirLeft)
	style.SetColorButtonPosition(imgui.DirRight)
	style.SetButtonTextAlign(imgui.Vec2{X: 0.5, Y: 0.5})
	style.SetSelectableTextAlign(imgui.Vec2{X: 0.0, Y: 0.0})
}

// fileExists reports whether the given path exists and is not a directory.
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
