package ui

import (
	"fmt"
	"os"

	"github.com/AllenDang/cimgui-go/imgui"
)

var mainFont *imgui.Font

func ApplyGUIStyles(baseFontSize, dpiScale float32) *imgui.Font {
	io := imgui.CurrentIO()

	fontSize := baseFontSize * dpiScale
	io.Fonts().Clear()

	mainFontPath := "assets/Exo-2.0/fonts/ttf/Exo2-Regular.ttf"
	if fileExists(mainFontPath) {
		fmt.Println("Using custom font:", mainFontPath)
		mainFont = io.Fonts().AddFontFromFileTTF(mainFontPath, fontSize)
	} else {
		mainFont = io.Fonts().AddFontDefault()
	}

	emojiPath := "assets/segoe-ui-emoji/seguiemj-1.45-3d.ttf"
	emojiSize := fontSize * 0.8 // 80% от основного шрифта
	ranges := getEmojiGlyphRanges()
	emojiConfig := imgui.NewFontConfig()
	emojiConfig.SetMergeMode(true)
	emojiConfig.SetPixelSnapH(true)
	emojiConfig.SetGlyphMinAdvanceX(emojiSize)
	emojiConfig.SetGlyphMaxAdvanceX(emojiSize * 2.0)

	if fileExists(emojiPath) {
		fmt.Println("Merging emoji font:", emojiPath)
		io.Fonts().AddFontFromFileTTFV(emojiPath, emojiSize, emojiConfig, &ranges[0])
	}
	io.Fonts().Build()

	style := imgui.CurrentStyle()
	scaleDimensions(style, dpiScale)

	imgui.StyleColorsDark()
	fmt.Printf("Applied DPI scaling: %.2fx to UI elements with emoji support\n", dpiScale)

	return mainFont
}

func getEmojiGlyphRanges() []imgui.Wchar {
	return []imgui.Wchar{
		0x1F300, 0x1F5FF, // Miscellaneous Symbols and Pictographs
		0x1F600, 0x1F64F, // Emoticons
		0x1F650, 0x1F67F, // Ornamental Dingbats
		0x1F680, 0x1F6FF, // Transport and Map Symbols
		0x1F700, 0x1F77F, // Alchemical Symbols
		0x1F780, 0x1F7FF, // Geometric Shapes Extended
		0x1F800, 0x1F8FF, // Supplemental Arrows-C
		0x1F900, 0x1F9FF, // Supplemental Symbols and Pictographs
		0x1FA00, 0x1FA6F, // Chess Symbols
		0x1FA70, 0x1FAFF, // Symbols and Pictographs Extended-A
		0x2600, 0x26FF, // Miscellaneous Symbols
		0x2700, 0x27BF, // Dingbats
		0x1F1E6, 0x1F1FF, // Regional Indicator Symbols (flags)
		0, // Null terminator
	}
}

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

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
