package terminal

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

var escapes = map[rune]func(*Terminal, string){
	'@': escapeInsertChars,
	'A': escapeMoveCursorUp,
	'B': escapeMoveCursorDown,
	'C': escapeMoveCursorRight,
	'D': escapeMoveCursorLeft,
	'd': escapeMoveCursorRow,
	'H': escapeMoveCursor,
	'f': escapeMoveCursor,
	'G': escapeMoveCursorCol,
	'h': escapePrivateModeOn,
	'L': escapeInsertLines,
	'l': escapePrivateModeOff,
	'm': escapeColorMode,
	'J': escapeEraseInScreen,
	'K': escapeEraseInLine,
	'P': escapeDeleteChars,
	'r': escapeSetScrollArea,
	's': escapeSaveCursor,
	'S': escapeScrollUp,
	'u': escapeRestoreCursor,
	'i': escapePrinterMode,
	'c': escapeDeviceAttribute,
}

func (t *Terminal) handleEscape(code string) {
	code = trimLeftZeros(code)
	if code == "" {
		return
	}

	runes := []rune(code)
	if esc, ok := escapes[runes[len(code)-1]]; ok {
		esc(t, code[:len(code)-1])
	} else if t.debug {
		log.Println("Unrecognised Escape:", code)
	}
}

func (t *Terminal) clearScreen() {
	t.moveCursor(0, 0)
	t.clearScreenFromCursor()
}

func (t *Terminal) clearScreenFromCursor() {
	row := t.content.Row(t.cursorRow)
	from := t.cursorCol
	if t.cursorCol > len(row.Cells) {
		from = len(row.Cells)
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

	fyne.Do(func() {
		t.content.SetRow(t.cursorRow, widget.TextGridRow{Cells: cells})

		for i := 0; i < t.cursorRow-1; i++ {
			t.content.SetRow(i, widget.TextGridRow{})
		}
	})
}

func (t *Terminal) handleVT100(code string) {
	switch code {
	case "(A":
		t.g0Charset = charSetAlternate
	case ")A":
		t.g1Charset = charSetAlternate
	case "(B":
		t.g0Charset = charSetANSII
	case ")B":
		t.g1Charset = charSetANSII
	case "(0":
		t.g0Charset = charSetDECSpecialGraphics
	case ")0":
		t.g1Charset = charSetDECSpecialGraphics
	default:
		if t.debug {
			log.Println("Unhandled VT100:", code)
		}
	}
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
		fyne.Do(t.cursorMoved)
	}
}

func escapeColorMode(t *Terminal, msg string) {
	t.handleColorEscape(msg)
}

func escapeDeleteChars(t *Terminal, msg string) {
	i, _ := strconv.Atoi(msg)
	if i == 0 {
		i = 1
	}
	right := t.cursorCol + i

	row := t.content.Row(t.cursorRow)
	cells := row.Cells[:t.cursorCol]
	if right < len(row.Cells) {
		cells = append(cells, row.Cells[right:]...)
	}

	t.content.SetRow(t.cursorRow, widget.TextGridRow{Cells: cells})
}

func escapeEraseInLine(t *Terminal, msg string) {
	mode, _ := strconv.Atoi(msg)
	switch mode {
	case 0:
		row := t.content.Row(t.cursorRow)
		if t.cursorCol >= len(row.Cells) {
			return
		}
		t.content.SetRow(t.cursorRow, widget.TextGridRow{Cells: row.Cells[:t.cursorCol]})
	case 1:
		row := t.content.Row(t.cursorRow)
		if t.cursorCol >= len(row.Cells) {
			return
		}
		cells := make([]widget.TextGridCell, t.cursorCol)
		t.content.SetRow(t.cursorRow, widget.TextGridRow{Cells: append(cells, row.Cells[t.cursorCol:]...)})
	case 2:
		row := t.content.Row(t.cursorRow)
		if t.cursorCol >= len(row.Cells) {
			return
		}
		cells := make([]widget.TextGridCell, len(row.Cells))
		t.content.SetRow(t.cursorRow, widget.TextGridRow{Cells: cells})
	}
}

func escapeEraseInScreen(t *Terminal, msg string) {
	mode, _ := strconv.Atoi(msg)
	switch mode {
	case 0:
		t.clearScreenFromCursor()
	case 1:
		t.clearScreenToCursor()
	case 2:
		t.clearScreen()
	}
}

func escapeInsertChars(t *Terminal, msg string) {
	chars, _ := strconv.Atoi(msg)
	if chars == 0 {
		chars = 1
	}

	newCells := make([]widget.TextGridCell, chars)
	cellStyle := &widget.CustomTextGridStyle{FGColor: t.currentFG, BGColor: t.currentBG}
	for i := range newCells {
		newCells[i] = widget.TextGridCell{
			Rune:  ' ',
			Style: cellStyle,
		}
	}

	row := &t.content.Rows[t.cursorRow]
	row.Cells = append(row.Cells[:t.cursorCol], append(newCells, row.Cells[t.cursorCol:]...)...)
}

func escapeInsertLines(t *Terminal, msg string) {
	rows, _ := strconv.Atoi(msg)
	if rows == 0 {
		rows = 1
	}
	i := t.scrollBottom
	for ; i > t.cursorRow-rows+1; i-- {
		t.content.SetRow(i, t.content.Row(i-rows))
	}
	for ; i >= t.cursorRow; i-- {
		t.content.SetRow(i, widget.TextGridRow{})
	}
}

func escapeMoveCursorUp(t *Terminal, msg string) {
	rows, _ := strconv.Atoi(msg)
	if rows == 0 {
		rows = 1
	}
	t.moveCursor(t.cursorRow-rows, t.cursorCol)
}

func escapeMoveCursorDown(t *Terminal, msg string) {
	rows, _ := strconv.Atoi(msg)
	if rows == 0 {
		rows = 1
	}
	t.moveCursor(t.cursorRow+rows, t.cursorCol)
}

func escapeMoveCursorRight(t *Terminal, msg string) {
	cols, _ := strconv.Atoi(msg)
	if cols == 0 {
		cols = 1
	}
	t.moveCursor(t.cursorRow, t.cursorCol+cols)
}

func escapeMoveCursorLeft(t *Terminal, msg string) {
	cols, _ := strconv.Atoi(msg)
	if cols == 0 {
		cols = 1
	}
	t.moveCursor(t.cursorRow, t.cursorCol-cols)
}

func escapeMoveCursorRow(t *Terminal, msg string) {
	row, _ := strconv.Atoi(msg)
	t.moveCursor(row-1, t.cursorCol)
}

func escapeMoveCursorCol(t *Terminal, msg string) {
	col, _ := strconv.Atoi(msg)
	t.moveCursor(t.cursorRow, col-1)
}

func escapePrivateMode(t *Terminal, msg string, enable bool) {
	modes := strings.Split(msg, ";")
	for _, mode := range modes {
		switch mode {
		case "7":
			//TODO wrap around mode
		case "20":
			t.newLineMode = enable
		case "25":
			t.cursorHidden = !enable
			fyne.Do(t.refreshCursor)
		case "9":
			if enable {
				t.onMouseDown = t.handleMouseDownX10
				t.onMouseUp = t.handleMouseUpX10
			} else {
				t.onMouseDown = nil
				t.onMouseUp = nil
			}
		case "1000":
			if enable {
				t.onMouseDown = t.handleMouseDownV200
				t.onMouseUp = t.handleMouseUpV200
			} else {
				t.onMouseDown = nil
				t.onMouseUp = nil
			}
		case "1049":
			t.bufferMode = enable
		case "2004":
			t.bracketedPasteMode = enable
		case "47":
			// TODO save screen
			/*
				if enable {
					// save screen
				} else {
					// restore screen
				}
			*/
		default:
			m := "l"
			if enable {
				m = "h"
			}
			if t.debug {
				log.Println("Unknown private escape code", fmt.Sprintf("%s%s", mode, m))
			}
		}
	}
}

func escapePrivateModeOff(t *Terminal, msg string) {
	escapePrivateMode(t, msg[1:], false)
}

func escapePrivateModeOn(t *Terminal, msg string) {
	escapePrivateMode(t, msg[1:], true)
}

func escapeMoveCursor(t *Terminal, msg string) {
	if !strings.Contains(msg, ";") {
		t.moveCursor(0, 0)
		return
	}

	parts := strings.Split(msg, ";")
	row, _ := strconv.Atoi(parts[0])
	col := 1
	if len(parts) == 2 {
		col, _ = strconv.Atoi(parts[1])
	}

	t.moveCursor(row-1, col-1)
}

func escapeRestoreCursor(t *Terminal, s string) {
	if s != "" {
		if t.debug {
			log.Println("Corrupt restore cursor escape", s+"u")
		}
		return
	}
	t.moveCursor(t.savedRow, t.savedCol)
}

func escapeSaveCursor(t *Terminal, _ string) {
	t.savedRow = t.cursorRow
	t.savedCol = t.cursorCol
}

func escapeSetScrollArea(t *Terminal, msg string) {
	parts := strings.Split(msg, ";")
	start := 0
	end := int(t.config.Rows) - 1
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
}

func escapeScrollUp(t *Terminal, msg string) {
	lines, _ := strconv.Atoi(msg)
	if lines == 0 {
		lines = 1
	}

	// Ensure we are within the scrollable area
	if t.cursorRow < t.scrollTop || t.cursorRow > t.scrollBottom {
		return
	}

	// Calculate new cursor position after scrolling
	newCursorRow := t.cursorRow - lines

	// Make sure we don't scroll above the scroll top
	if newCursorRow < t.scrollTop {
		newCursorRow = t.scrollTop
	}

	// Move cursor to the new position
	t.moveCursor(newCursorRow, t.cursorCol)

	// Perform the actual scrolling action
	for i := t.scrollTop; i <= t.scrollBottom-lines; i++ {
		t.content.SetRow(i, t.content.Row(i+lines))
	}
	for i := t.scrollBottom - lines + 1; i <= t.scrollBottom; i++ {
		t.content.SetRow(i, widget.TextGridRow{}) // Clear the last lines
	}
}

func trimLeftZeros(s string) string {
	if s == "" {
		return s
	}

	i := 0
	for _, r := range s {
		if r > '0' {
			break
		}
		i++
	}

	return s[i:]
}

func escapePrinterMode(t *Terminal, code string) {
	switch code {
	case "5":
		t.state.printing = true
	case "4":
		t.state.printing = false
		if t.printData != nil {
			if t.printer != nil {
				// spool the printer
				t.printer.Print(t.printData)
			} else if t.debug {
				log.Println("Print data was received but no printer has been set")
			}

		}
		t.printData = nil
	default:
		if t.debug {
			log.Println("Unknown printer mode", code)
		}
	}
}

func escapeDeviceAttribute(t *Terminal, code string) {
	if len(code) == 0 { // query
		_, _ = t.in.Write([]byte{asciiEscape})
		_, _ = t.in.Write([]byte("[?2;22c")) // printer; color
		return
	}

	if t.debug {
		switch code[0] {
		case '>':
			log.Println("Unhandled secondary device attribute", code[1])
		case '=':
			log.Println("Unhandled tertiary device attribute", code[1])
		default:
			log.Println("Unknown device attribute", code[0])
		}
	}
}
