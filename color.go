package terminal

import (
	"image/color"
	"log"
	"strings"

	"fyne.io/fyne/v2/theme"
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
		mode = modes[1]
	} else if len(modes) >= 2 && modes[1] == "1" {
		bright = true
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
			currentBG = theme.TextColor()
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
		if currentFG == theme.TextColor() {
			currentBG = nil
		} else {
			currentBG = currentFG
		}
		if bg == theme.DisabledButtonColor() {
			currentFG = nil
		} else {
			currentFG = bg
		}
	case "30":
		if bright {
			currentFG = &color.RGBA{85, 85, 85, 255}
		} else {
			currentFG = color.Black
		}
	case "31":
		if bright {
			currentFG = &color.RGBA{255, 85, 85, 255}
		} else {
			currentFG = &color.RGBA{170, 0, 0, 255}
		}
	case "32":
		if bright {
			currentFG = &color.RGBA{85, 255, 255, 255}
		} else {
			currentFG = &color.RGBA{0, 170, 0, 255}
		}
	case "33":
		if bright {
			currentFG = &color.RGBA{255, 255, 85, 255}
		} else {
			currentFG = &color.RGBA{170, 170, 0, 255}
		}
	case "34":
		if bright {
			currentFG = &color.RGBA{85, 85, 255, 255}
		} else {
			currentFG = &color.RGBA{0, 0, 170, 255}
		}
	case "35":
		if bright {
			currentFG = &color.RGBA{255, 85, 255, 255}
		} else {
			currentFG = &color.RGBA{170, 0, 170, 255}
		}
	case "36":
		if bright {
			currentFG = &color.RGBA{85, 255, 255, 255}
		} else {
			currentFG = &color.RGBA{0, 170, 170, 255}
		}
	case "37":
		if bright {
			currentFG = &color.RGBA{255, 255, 255, 255}
		} else {
			currentFG = &color.RGBA{170, 170, 170, 255}
		}
	case "39":
		currentFG = nil
	case "40":
		currentBG = color.Black
	case "41":
		currentBG = &color.RGBA{170, 0, 0, 255}
	case "42":
		currentBG = &color.RGBA{0, 170, 0, 255}
	case "43":
		currentBG = &color.RGBA{170, 170, 0, 255}
	case "44":
		currentBG = &color.RGBA{0, 0, 170, 255}
	case "45":
		currentBG = &color.RGBA{170, 0, 170, 255}
	case "46":
		currentBG = &color.RGBA{0, 255, 255, 255}
	case "47":
		currentBG = &color.RGBA{170, 170, 170, 255}
	case "49":
		currentBG = nil
	default:
		log.Println("Unsupported graphics mode", mode)
	}
}
