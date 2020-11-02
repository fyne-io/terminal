package terminal

import (
	"log"
	"strconv"
	"strings"

	"fyne.io/fyne/widget"
)

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
		if t.cursorCol >= len(row.Cells) {
			return
		}
		t.content.SetRow(t.cursorRow, widget.TextGridRow{Cells: row.Cells[:t.cursorCol]})
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
			t.moveCursor(t.cursorRow, t.cursorCol+cols)
		case "D":
			cols, _ := strconv.Atoi(message)
			t.moveCursor(t.cursorRow, t.cursorCol-cols)
		case "H", "f":
			parts := strings.Split(message, ";")
			row, _ := strconv.Atoi(parts[0])
			col, _ := strconv.Atoi(parts[1])

			t.moveCursor(row, col)
		case "m":
			t.handleColorEscape(message)
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
	t.content.SetRow(t.cursorRow, widget.TextGridRow{Cells: row.Cells[:t.cursorCol]})

	for i := t.cursorRow; i < len(t.content.Rows); i++ {
		t.content.SetRow(i, widget.TextGridRow{})
	}
}

func (t *Terminal) handleVT100(code string) {
	log.Println("Unhandled VT100:", code)
}

func (t *Terminal) moveCursor(row, col int) {
	if t.config.Columns == 0 || t.config.Rows == 0 {
		return
	}
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
