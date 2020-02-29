package terminal

import (
	"log"
	"strconv"
	"strings"
)

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
		t.content.SetRow(t.cursorRow, row[:t.cursorCol])
	default: // check mode (last letter) then match
		message := code[:len(code)-1]
		switch code[len(code)-1:] {
		case "H":
			parts := strings.Split(message, ";")
			row, _ := strconv.Atoi(parts[0])
			col, _ := strconv.Atoi(parts[1])

			if row < len(t.content.Buffer) {
				t.cursorRow = row
			}
			line := t.content.Row(t.cursorRow)
			if col < len(line) {
				t.cursorCol = col
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
