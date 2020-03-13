package terminal

import (
	"image/color"
	"log"
	"strconv"
	"strings"

	"fyne.io/fyne/theme"
)

var currentFG, currentBG color.Color

func (t *Terminal) handleEscape(code string) {
	switch code { // exact matches
	case "H", ";H":
		t.cursorCol = 0
		t.cursorRow = 0
		t.cursorMoved()
	case "2J":
		t.clearScreen()
	case "K":
		row := t.content.Row(t.cursorRow)
		if t.cursorCol > len(row) {
			return
		}
		t.content.SetRow(t.cursorRow, row[:t.cursorCol])
	default: // check mode (last letter) then match
		message := code[:len(code)-1]
		switch code[len(code)-1:] {
		case "H":
			parts := strings.Split(message, ";")
			row, _ := strconv.Atoi(parts[0])
			col, _ := strconv.Atoi(parts[1])

			if row < len(t.content.Content) {
				t.cursorRow = row
			}
			line := t.content.Row(t.cursorRow)
			if col < len(line) {
				t.cursorCol = col
			}
		case "m":
			if message == "" || message == "0" {
				currentBG = nil
				currentFG = nil
				return
			}

			modes := strings.Split(message, ";")
			bright := false

			mode := modes[0]
			if mode == "1" {
				bright = true
				if len(modes) <= 1 {
					break
				}
				mode = modes[1]
			} else if len(modes) >= 2 && modes[1] == "1" {
				bright = true
			}
			for _, mode := range modes {
				switch mode {
				case "1": // ignore, handled above
				case "7": // reverse
					currentBG, currentFG = theme.TextColor(), theme.ButtonColor() //currentFG, currentBG
				case "27": // reverse off
					currentBG, currentFG = nil, nil//currentFG, currentBG
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
		default:
			log.Println("Unrecognised Escape:", code)
		}
	}
}

func (t *Terminal) clearScreen() {
	t.content.SetText("")
	t.cursorCol = 0
	t.cursorRow = 0
}

func (t *Terminal) handleVT100(code string) {
	log.Println("Unhandled VT100:", code)
}
