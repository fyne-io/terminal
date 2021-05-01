package terminal

import (
	"time"

	"fyne.io/fyne/v2/widget"
)

const (
	asciiBell      = 7
	asciiBackspace = 8
	asciiEscape    = 27

	noEscape = 5000
	tabWidth = 8
)

var specialChars = map[rune]func(t *Terminal){
	asciiBell:      handleOutputBell,
	asciiBackspace: handleOutputBackspace,
	'\n':           handleOutputLineFeed,
	'\r':           handleOutputCarriageReturn,
	'\t':           handleOutputTab,
	0x0e:           nil, // handle switch to G1 character set
	0x0f:           nil, // handle switch to G1 character set
}

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
			}
			switch r {
			case '\\':
				t.handleOSC(state.code)
				state.code = ""
				state.osc = false
			case ']':
				state.osc = true
			case '(', ')':
				state.vt100 = r
			case '7':
				t.savedRow = t.cursorRow
				t.savedCol = t.cursorCol
			case '8':
				t.cursorRow = t.savedRow
				t.cursorCol = t.savedCol
			case 'D':
				t.scrollDown()
			case 'M':
				t.scrollUp()
			case '=', '>':
			}
			state.esc = noEscape
			continue
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

		if out, ok := specialChars[r]; ok {
			if out == nil {
				continue
			}
			out(t)
		} else {
			t.handleOutputChar(r)
		}
	}

	// record progress for next chunk of buffer
	if state.esc != noEscape {
		state.esc = -1 - (len(state.code))
		previous = state
	}
}

func (t *Terminal) handleOutputChar(r rune) {
	if t.cursorCol >= int(t.config.Columns) || t.cursorRow >= int(t.config.Rows) {
		return // TODO handle wrap?
	}
	for len(t.content.Rows)-1 < t.cursorRow {
		t.content.Rows = append(t.content.Rows, widget.TextGridRow{})
	}

	cellStyle := &widget.CustomTextGridStyle{FGColor: t.currentFG, BGColor: t.currentBG}
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

func (t *Terminal) ringBell() {
	t.bell = true
	t.Refresh()

	time.Sleep(time.Millisecond * 300)
	t.bell = false
	t.Refresh()
}

func (t *Terminal) scrollUp() {
	for i := t.scrollBottom; i > t.scrollTop; i-- {
		t.content.Rows[i] = t.content.Row(i - 1)
	}
	t.content.Rows[t.scrollTop] = widget.TextGridRow{}
	t.content.Refresh()
}

func (t *Terminal) scrollDown() {
	i := t.scrollTop
	for ; i < t.scrollBottom && i < len(t.content.Rows)-1; i++ {
		t.content.Rows[i] = t.content.Row(i + 1)
	}
	for ; i < len(t.content.Rows); i++ {
		if len(t.content.Rows) > t.scrollBottom {
			t.content.Rows[t.scrollBottom] = widget.TextGridRow{}
		} else {
			t.content.Rows = append(t.content.Rows, widget.TextGridRow{})
		}
	}
	t.content.Refresh()
}

func handleOutputBackspace(t *Terminal) {
	row := t.content.Row(t.cursorRow)
	if len(row.Cells) == 0 {
		return
	}
	t.moveCursor(t.cursorRow, t.cursorCol-1)
}

func handleOutputBell(t *Terminal) {
	go t.ringBell()
}

func handleOutputCarriageReturn(t *Terminal) {
	t.moveCursor(t.cursorRow, 0)
}

func handleOutputLineFeed(t *Terminal) {
	if t.cursorRow == t.scrollBottom {
		t.scrollDown()
	} else {
		t.moveCursor(t.cursorRow+1, t.cursorCol)
	}
}

func handleOutputTab(t *Terminal) {
	end := t.cursorCol - t.cursorCol%tabWidth + tabWidth
	for t.cursorCol < end {
		t.handleOutputChar(' ')
	}
}
