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
	brightColors = []color.Color {
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

func (t *Terminal) handleColorMode(mode string, bright bool) {
	switch mode {
	case "0", "00":
		currentBG, currentFG = nil, nil
	case "1", "01": // ignore, handled above
	case "4", "24": // italic
	case "7": // reverse
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
	case "27": // reverse off
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
	case "30", "90":
		if bright || mode[0] == '9' {
			currentFG = brightColors[0]
		} else {
			currentFG = basicColors[0]
		}
	case "31", "91":
		if bright || mode[0] == '9' {
			currentFG = brightColors[1]
		} else {
			currentFG = basicColors[1]
		}
	case "32", "92":
		if bright || mode[0] == '9' {
			currentFG = brightColors[2]
		} else {
			currentFG = basicColors[2]
		}
	case "33", "93":
		if bright || mode[0] == '9' {
			currentFG = brightColors[3]
		} else {
			currentFG = basicColors[3]
		}
	case "34", "94":
		if bright || mode[0] == '9' {
			currentFG = brightColors[4]
		} else {
			currentFG = basicColors[4]
		}
	case "35", "95":
		if bright || mode[0] == '9' {
			currentFG = brightColors[5]
		} else {
			currentFG = basicColors[5]
		}
	case "36", "96":
		if bright || mode[0] == '9' {
			currentFG = brightColors[6]
		} else {
			currentFG = basicColors[6]
		}
	case "37", "97":
		if bright || mode[0] == '9' {
			currentFG = brightColors[7]
		} else {
			currentFG = basicColors[7]
		}
	case "39":
		currentFG = nil
	case "40", "100":
		if bright || mode[0] == '1' {
			currentBG = brightColors[0]
		} else {
			currentBG = basicColors[0]
		}
	case "41", "101":
		if bright || mode[0] == '1' {
			currentBG = brightColors[1]
		} else {
			currentBG = basicColors[1]
		}
	case "42", "102":
		if bright || mode[0] == '1' {
			currentBG = brightColors[2]
		} else {
			currentBG = basicColors[2]
		}
	case "43", "103":
		if bright || mode[0] == '1' {
			currentBG = brightColors[3]
		} else {
			currentBG = basicColors[3]
		}
	case "44", "104":
		if bright || mode[0] == '1' {
			currentBG = brightColors[4]
		} else {
			currentBG = basicColors[4]
		}
	case "45", "105":
		if bright || mode[0] == '1' {
			currentBG = brightColors[5]
		} else {
			currentBG = basicColors[5]
		}
	case "46", "106":
		if bright || mode[0] == '1' {
			currentBG = brightColors[6]
		} else {
			currentBG = basicColors[6]
		}
	case "47", "107":
		if bright || mode[0] == '1' {
			currentBG = brightColors[7]
		} else {
			currentBG = basicColors[7]
		}
	case "49":
		currentBG = nil
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
		inc := 256/5
		id -= 16
		b := id % 6
		id = (id-b) / 6
		g := id % 6
		r := (id-g) / 6
		c = &color.RGBA{uint8(r*inc), uint8(g*inc), uint8(b*inc), 255}
	} else if id <= 255 {
		id -= 232
		inc := 256/24
		y := id*inc
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
