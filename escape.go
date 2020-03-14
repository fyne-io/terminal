package terminal

import (
	"image/color"
	"log"
	"strconv"
	"strings"

	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
)

var currentFG, currentBG color.Color

func (t *Terminal) handleEscape(code string) {
	switch code { // exact matches
	case "H", "f":
		t.moveCursor(0, 0)
	case "J":
		t.clearScreenFromCursor()
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
		case "A":
			rows, _ := strconv.Atoi(message)
			t.moveCursor(t.cursorRow-rows, t.cursorCol)
		case "B":
			rows, _ := strconv.Atoi(message)
			t.moveCursor(t.cursorRow+rows, t.cursorCol)
		case "C":
			cols, _ := strconv.Atoi(message)
			t.moveCursor(t.cursorRow, t.cursorCol-cols)
		case "D":
			cols, _ := strconv.Atoi(message)
			t.moveCursor(t.cursorRow, t.cursorCol+cols)
		case "H", "f":
			parts := strings.Split(message, ";")
			row, _ := strconv.Atoi(parts[0])
			col, _ := strconv.Atoi(parts[1])

			t.moveCursor(row, col)
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
					currentBG, currentFG = nil, nil //currentFG, currentBG
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
	t.moveCursor(0, 0)
	t.clearScreenFromCursor()
}

func (t *Terminal) clearScreenFromCursor() {
	row := t.content.Row(t.cursorRow)
	t.content.SetRow(t.cursorRow, row[:t.cursorCol])

	for i := t.cursorRow; i < len(t.content.Content); i++ {
		t.content.SetRow(i, []widget.TextGridCell{})
	}
}

func (t *Terminal) handleVT100(code string) {
	log.Println("Unhandled VT100:", code)
}

func (t *Terminal) moveCursor(row, col int) {
	if col < 0 {
		col = 0
	} else if col >= int(t.config.Columns) {
		col = int(t.config.Columns) - 1
	}

	if row < 0 {
		row = 0
	} else if row >= int(t.config.Rows) {
		row = int(t.config.Rows) - 1
	}

	t.cursorCol = col
	t.cursorRow = row

	if t.cursorMoved != nil {
		t.cursorMoved()
	}
}
