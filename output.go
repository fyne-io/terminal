package terminal

import (
	"time"

	"fyne.io/fyne/widget"
)

func (t *Terminal) handleOutput(buf []byte) {
	esc := -5
	osc := false
	vt100 := rune(0)
	code := ""
	for i, r := range []rune(string(buf)) {
		if r == asciiEscape {
			esc = i
			continue
		}
		if esc == i-1 {
			if r == '[' {
				continue
			} else if r == ']' {
				osc = true
				continue
			} else if r == '(' || r == ')' {
				vt100 = r
				continue
			} else {
				esc = -5
			}
		}
		if osc {
			if r == asciiBell || r == 0 {
				t.handleOSC(code)
				code = ""
				osc = false
			} else {
				code += string(r)
			}
			continue
		} else if vt100 != 0 {
			t.handleVT100(string([]rune{vt100, r}))
			vt100 = 0
			continue
		} else if esc != -5 {
			code += string(r)
			if (r < '0' || r > '9') && r != ';' && r != '=' && r != '?' {
				t.handleEscape(code)
				code = ""
				esc = -5
			}
			continue
		}

		switch r {
		case asciiBackspace:
			row := t.content.Row(t.cursorRow)
			if len(row.Cells) == 0 {
				continue
			}
			t.content.SetRow(t.cursorRow, widget.TextGridRow{Cells: row.Cells[:len(row.Cells)-1]})
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
			if len(t.content.Rows)-1 < int(t.cursorRow) {
				t.content.Rows = append(t.content.Rows, widget.TextGridRow{})
			}

			if r == '\t' { // TODO handle tab
				r = ' '
			}

			cellStyle := &widget.CustomTextGridStyle{FGColor: currentFG, BGColor: currentBG}

			if len(t.content.Rows[t.cursorRow].Cells)-1 < int(t.cursorCol) {
				newcell := widget.TextGridCell{
					Rune:  r,
					Style: cellStyle,
				}
				t.content.Rows[t.cursorRow].Cells = append(t.content.Rows[t.cursorRow].Cells, newcell)
			} else {
				t.content.Rows[t.cursorRow].Cells[t.cursorCol].Rune = r
				t.content.Rows[t.cursorRow].Cells[t.cursorCol].Style = cellStyle
			}
			t.cursorCol++
		}
		esc = -5
		code = ""
	}

	if osc { // it could end at the buffer end
		t.handleOSC(code)
		return
	}
	t.Refresh()
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
