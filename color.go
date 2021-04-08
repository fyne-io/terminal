package terminal

import (
	"image/color"
	"log"
	"strconv"
	"strings"

	"fyne.io/fyne/v2/theme"
)

var (
	basicColors = []color.Color{
		color.Black,
		&color.RGBA{170, 0, 0, 255},
		&color.RGBA{0, 170, 0, 255},
		&color.RGBA{170, 170, 0, 255},
		&color.RGBA{0, 0, 170, 255},
		&color.RGBA{170, 0, 170, 255},
		&color.RGBA{0, 255, 255, 255},
		&color.RGBA{170, 170, 170, 255},
	}
	brightColors = []color.Color{
		&color.RGBA{85, 85, 85, 255},
		&color.RGBA{255, 85, 85, 255},
		&color.RGBA{85, 255, 255, 255},
		&color.RGBA{255, 255, 85, 255},
		&color.RGBA{85, 85, 255, 255},
		&color.RGBA{255, 85, 255, 255},
		&color.RGBA{85, 255, 255, 255},
		&color.RGBA{255, 255, 255, 255},
	}
)

var currentFG, currentBG color.Color

func (t *Terminal) handleColorEscape(message string) {
	if message == "" || message == "0" {
		currentBG = nil
		currentFG = nil
		return
	}

	modes := strings.Split(message, ";")
	bright := false

	mode := modes[0]
	if mode == "1" || mode == "01" {
		bright = true
		if len(modes) <= 1 {
			return
		}
	} else if len(modes) >= 2 && modes[1] == "1" {
		bright = true
	}
	if (mode == "38" || mode == "48") && len(modes) >= 2 {
		if modes[1] == "5" && len(modes) >= 3 {
			t.handleColorModeMap(mode, modes[2])
			modes = modes[3:]
		} else if modes[1] == "2" && len(modes) >= 5 {
			t.handleColorModeRGB(mode, modes[2], modes[3], modes[4])
			modes = modes[5:]
		}
	}
	for _, mode := range modes {
		t.handleColorMode(mode, bright)
	}
}

func (t *Terminal) handleColorMode(modeStr string, bright bool) {
	mode, err := strconv.Atoi(modeStr)
	if err != nil {
		log.Println("Failed to parse color mode", modeStr)
	}
	switch mode {
	case 0:
		currentBG, currentFG = nil, nil
	case 1: // ignore, handled above
	case 4, 24: // italic
	case 7: // reverse
		bg := currentBG
		if currentFG == nil {
			currentBG = theme.ForegroundColor()
		} else {
			currentBG = currentFG
		}
		if bg == nil {
			currentFG = theme.DisabledButtonColor()
		} else {
			currentFG = bg
		}
	case 27: // reverse off
		bg := currentBG
		if currentFG == theme.ForegroundColor() {
			currentBG = nil
		} else {
			currentBG = currentFG
		}
		if bg == theme.DisabledButtonColor() {
			currentFG = nil
		} else {
			currentFG = bg
		}
	case 30, 31, 32, 33, 34, 35, 36, 37:
		if bright {
			currentFG = brightColors[mode-30]
		} else {
			currentFG = basicColors[mode-30]
		}
	case 39:
		currentFG = nil
	case 40, 41, 42, 43, 44, 45, 46, 47:
		if bright {
			currentBG = brightColors[mode-40]
		} else {
			currentBG = basicColors[mode-40]
		}
	case 49:
		currentBG = nil
	case 90, 91, 92, 93, 94, 95, 96, 97:
		currentFG = brightColors[mode-90]
	case 100, 101, 102, 103, 104, 105, 106, 107:
		currentBG = brightColors[mode-100]
	default:
		log.Println("Unsupported graphics mode", mode)
	}
}

func (t *Terminal) handleColorModeMap(mode, ids string) {
	var c color.Color
	id, err := strconv.Atoi(ids)
	if err != nil {
		log.Println("Invalid color map ID", ids)
		return
	}
	if id <= 7 {
		c = basicColors[id]
	} else if id <= 15 {
		c = brightColors[id-8]
	} else if id <= 231 {
		inc := 256 / 5
		id -= 16
		b := id % 6
		id = (id - b) / 6
		g := id % 6
		r := (id - g) / 6
		c = &color.RGBA{uint8(r * inc), uint8(g * inc), uint8(b * inc), 255}
	} else if id <= 255 {
		id -= 232
		inc := 256 / 24
		y := id * inc
		c = &color.Gray{uint8(y)}
	} else {
		log.Println("Invalid colour map id", id)
	}

	if mode == "38" {
		currentFG = c
	} else if mode == "48" {
		currentBG = c
	}
}

func (t *Terminal) handleColorModeRGB(mode, rs, gs, bs string) {
	r, _ := strconv.Atoi(rs)
	g, _ := strconv.Atoi(gs)
	b, _ := strconv.Atoi(bs)
	c := &color.RGBA{uint8(r), uint8(g), uint8(b), 255}

	if mode == "38" {
		currentFG = c
	} else if mode == "48" {
		currentBG = c
	}
}
