package terminal

import (
	"time"

	"fyne.io/fyne/v2/widget"
)

const noEscape = 5000

var previous *parseState

type parseState struct {
	code  string
	esc   int
	osc   bool
	vt100 rune
}

func (t *Terminal) handleOutput(buf []byte) {
	state := &parseState{}
	if previous != nil {
		state = previous
		previous = nil
	} else {
		state.esc = noEscape
	}

	for i, r := range []rune(string(buf)) {
		if r == asciiEscape {
			state.esc = i
			continue
		}
		if state.esc == i-1 {
			if r == '[' {
				continue
			} else if r == ']' {
				state.osc = true
				state.esc = noEscape
				continue
			} else if r == '(' || r == ')' {
				state.vt100 = r
				continue
			} else if r == '7' {
				t.savedRow = t.cursorRow
				t.savedCol = t.cursorCol
				state.esc = noEscape
				continue
			} else if r == '8' {
				t.cursorRow = t.savedRow
				t.cursorCol = t.savedCol
				state.esc = noEscape
				continue
			} else {
				state.esc = noEscape
			}
		}
		if state.osc {
			if r == asciiBell || r == 0 {
				t.handleOSC(state.code)
				state.code = ""
				state.osc = false
			} else {
				state.code += string(r)
			}
			continue
		} else if state.vt100 != 0 {
			t.handleVT100(string([]rune{state.vt100, r}))
			state.vt100 = 0
			continue
		} else if state.esc != noEscape {
			state.code += string(r)
			if (r < '0' || r > '9') && r != ';' && r != '=' && r != '?' {
				t.handleEscape(state.code)
				state.code = ""
				state.esc = noEscape
			}
			continue
		}

		switch r {
		case asciiBackspace:
			row := t.content.Row(t.cursorRow)
			if len(row.Cells) == 0 {
				continue
			}
			t.moveCursor(t.cursorRow, t.cursorCol-1)
			continue
		case '\n': // line feed
			if t.cursorRow == int(t.config.Rows-1) {
				for i = 0; i < t.cursorRow; i++ {
					t.content.SetRow(i, t.content.Row(i+1))
				}
				t.content.SetRow(i, widget.TextGridRow{})
			} else {
				t.moveCursor(t.cursorRow+1, t.cursorCol)
			}
			continue
		case '\r': // carriage return
			t.moveCursor(t.cursorRow, 0)
			continue
		case asciiBell:
			go t.ringBell()
			continue
		default:
			if t.cursorCol >= int(t.config.Columns) || t.cursorRow >= int(t.config.Rows) {
				break // TODO handle wrap?
			}
			for len(t.content.Rows)-1 < t.cursorRow {
				t.content.Rows = append(t.content.Rows, widget.TextGridRow{})
			}

			if r == '\t' { // TODO handle tab
				r = ' '
			}

			cellStyle := &widget.CustomTextGridStyle{FGColor: currentFG, BGColor: currentBG}

			for len(t.content.Rows[t.cursorRow].Cells)-1 < t.cursorCol {
				newCell := widget.TextGridCell{
					Rune:  ' ',
					Style: cellStyle,
				}
				t.content.Rows[t.cursorRow].Cells = append(t.content.Rows[t.cursorRow].Cells, newCell)
			}

			cell := t.content.Rows[t.cursorRow].Cells[t.cursorCol]
			if cell.Rune != r || cell.Style.TextColor() != cellStyle.FGColor || cell.Style.BackgroundColor() != cellStyle.BGColor {
				cell.Rune = r
				cell.Style = cellStyle
				t.content.SetCell(t.cursorRow, t.cursorCol, cell)
			}
			t.cursorCol++
		}
	}

	// record progress for next chunk of buffer
	if state.esc != noEscape {
		state.esc = -1 - (len(state.code))
		previous = state
	}
}

func (t *Terminal) ringBell() {
	t.bell = true
	t.Refresh()

	select {
	case <-time.After(time.Millisecond * 300):
		t.bell = false
		t.Refresh()
	}
}
