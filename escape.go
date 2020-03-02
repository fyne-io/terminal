package terminal

import (
	"image/color"
	"log"
	"strconv"
	"strings"

	"fyne.io/fyne/theme"
)

var currentFG color.Color

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
				currentFG = nil
				return
			}

			modes := strings.Split(message, ";")
			for _, mode := range modes {
				switch mode {
				case "7": // reverse
					currentFG = theme.BackgroundColor()
				case "27": // reverse off
					currentFG = nil
				case "30":
					currentFG = color.Black
				case "31":
					currentFG = &color.RGBA{255, 0, 0, 255}
				case "32":
					currentFG = &color.RGBA{0, 255, 0, 255}
				case "33":
					currentFG = &color.RGBA{255, 255, 0, 255}
				case "34":
					currentFG = &color.RGBA{0, 0, 255, 255}
				case "35":
					currentFG = &color.RGBA{255, 0, 255, 255}
				case "36":
					currentFG = &color.RGBA{0, 255, 255, 255}
				case "37":
					currentFG = color.White
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
