package terminal

import (
	"log"
	"strconv"
	"strings"

	"fyne.io/fyne/v2/widget"
)

func (t *Terminal) handleEscape(code string) {
	switch code { // exact matches
	case "A":
		t.moveCursor(t.cursorRow-1, t.cursorCol)
	case "B":
		t.moveCursor(t.cursorRow+1, t.cursorCol)
	case "C":
		t.moveCursor(t.cursorRow, t.cursorCol+1)
	case "D":
		t.moveCursor(t.cursorRow, t.cursorCol-1)
	case "H", "f":
		t.moveCursor(0, 0)
	case "?25h":
		t.cursorHidden = false
		t.refreshCursor()
	case "?25l":
		t.cursorHidden = true
		t.refreshCursor()
	case "?1049h":
		t.bufferMode = true
	case "?1049l":
		t.bufferMode = false
	case "J", "0J":
		t.clearScreenFromCursor()
	case "1J":
		t.clearScreenToCursor()
	case "2J":
		t.clearScreen()
	case "K", "0K":
		row := t.content.Row(t.cursorRow)
		if t.cursorCol >= len(row.Cells) {
			return
		}
		t.content.SetRow(t.cursorRow, widget.TextGridRow{Cells: row.Cells[:t.cursorCol]})
	case "1K":
		row := t.content.Row(t.cursorRow)
		if t.cursorCol >= len(row.Cells) {
			return
		}
		cells := make([]widget.TextGridCell, t.cursorCol)
		t.content.SetRow(t.cursorRow, widget.TextGridRow{Cells: append(cells, row.Cells[t.cursorCol:]...)})
	case "2K":
		row := t.content.Row(t.cursorRow)
		if t.cursorCol >= len(row.Cells) {
			return
		}
		cells := make([]widget.TextGridCell, len(row.Cells))
		t.content.SetRow(t.cursorRow, widget.TextGridRow{Cells: cells})
	case "s":
		t.savedRow = t.cursorRow
		t.savedCol = t.cursorCol
	case "u":
		t.moveCursor(t.savedRow, t.savedCol)
	default: // check mode (last letter) then match
		message := code[:len(code)-1]
		part := code[len(code)-1:]
		switch part {
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
		case "d":
			row, _ := strconv.Atoi(message)
			t.moveCursor(row-1, t.cursorCol)
		case "G":
			col, _ := strconv.Atoi(message)
			t.moveCursor(t.cursorRow, col-1)
		case "H", "f":
			parts := strings.Split(message, ";")
			row, _ := strconv.Atoi(parts[0])
			col := 1
			if len(parts) == 2 {
				col, _ = strconv.Atoi(parts[1])
			}

			t.moveCursor(row-1, col-1)
		case "m":
			t.handleColorEscape(message)
		case "P":
			dels, _ := strconv.Atoi(message)
			for i := 0; i < dels-1; i++ {
				_, _ = t.pty.Write([]byte{asciiBackspace})
			}
		case "r":
			parts := strings.Split(message, ";")
			start := 0
			end := int(t.config.Rows)
			if len(parts) == 2 {
				if parts[0] != "" {
					start, _ = strconv.Atoi(parts[0])
					start--
				}
				if parts[1] != "" {
					end, _ = strconv.Atoi(parts[1])
					end--
				}
			}

			t.scrollTop = start
			t.scrollBottom = end
		case "L":
			rows, _ := strconv.Atoi(message)
			if rows == 0 {
				rows = 1
			}
			i := t.scrollBottom
			for ; i > t.cursorRow-rows; i-- {
				t.content.SetRow(i, t.content.Row(i-rows))
			}
			for ; i >= t.cursorRow; i-- {
				t.content.SetRow(i, widget.TextGridRow{})
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
	from := t.cursorCol
	if t.cursorCol >= len(row.Cells) {
		from = len(row.Cells) - 1
	}
	if from > 0 {
		t.content.SetRow(t.cursorRow, widget.TextGridRow{Cells: row.Cells[:from]})
	} else {
		t.content.SetRow(t.cursorRow, widget.TextGridRow{})
	}

	for i := t.cursorRow + 1; i < len(t.content.Rows); i++ {
		t.content.SetRow(i, widget.TextGridRow{})
	}
}

func (t *Terminal) clearScreenToCursor() {
	row := t.content.Row(t.cursorRow)
	cells := make([]widget.TextGridCell, t.cursorCol)
	if t.cursorCol < len(row.Cells) {
		cells = append(cells, row.Cells[t.cursorCol:]...)
	}
	t.content.SetRow(t.cursorRow, widget.TextGridRow{Cells: cells})

	for i := 0; i < t.cursorRow-1; i++ {
		t.content.SetRow(i, widget.TextGridRow{})
	}
}

func (t *Terminal) handleVT100(code string) {
	if code == "(A" || code == ")A" || code == "(B" || code == ")B" {
		return // keycode handling A = en_GB, B = en_US
	}
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
