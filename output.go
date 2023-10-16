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

var charSetMap = map[charSet]func(rune) rune{
	charSetANSII: func(r rune) rune {
		return r
	},
	charSetDECSpecialGraphics: func(r rune) rune {
		m, ok := decSpecialGraphics[r]
		if ok {
			return m
		}
		return r
	},
	charSetAlternate: func(r rune) rune {
		return r
	},
}

var specialChars = map[rune]func(t *Terminal){
	asciiBell:      handleOutputBell,
	asciiBackspace: handleOutputBackspace,
	'\n':           handleOutputLineFeed,
	'\v':           handleOutputLineFeed,
	'\f':           handleOutputLineFeed,
	'\r':           handleOutputCarriageReturn,
	'\t':           handleOutputTab,
	0x0e:           handleShiftOut, // handle switch to G1 character set
	0x0f:           handleShiftIn,  // handle switch to G0 character set
}

// decSpecialGraphics is for ESC(0 graphics mode
// https://en.wikipedia.org/wiki/DEC_Special_Graphics
var decSpecialGraphics = map[rune]rune{
	'`': '◆', // filled in diamond
	'a': '▒', // filled in box
	'b': '␉', // horizontal tab symbol
	'c': '␌', // form feed symbol
	'd': '␍', // carriage return symbol
	'e': '␊', // line feed symbol
	'f': '°', // degree symbol
	'g': '±', // plus-minus sign
	'h': '␤', // new line symbol
	'i': '␋', // vertical tab symbol
	'j': '┘', // bottom right
	'k': '┐', // top right
	'l': '┌', // top left
	'm': '└', // bottom left
	'n': '┼', // cross
	'o': '⎺', // scan line 1
	'p': '⎻', // scan line 2
	'q': '─', // scan line 3
	'r': '─', // scan line 4
	's': '⎽', // scan line 5
	't': '├', // vertical and right
	'u': '┤', // vertical and left
	'v': '┴', // horizontal and up
	'w': '┬', // horizontal and down
	'x': '│', // vertical bar
	'y': '≤', // less or equal
	'z': '≥', // greater or equal
	'{': 'π', // pi
	'|': '≠', // not equal
	'}': '£', // Pounds currency symbol
	'~': '·', // centered dot
}

var previous *parseState

type parseState struct {
	code  string
	esc   int
	osc   bool
	vt100 rune
}

func (t *Terminal) handleOutput(buf []byte) {
	if t.hasSelectedText() {
		t.clearSelectedText()
	}
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
			if (r < '0' || r > '9') && r != ';' && r != '=' && r != '?' && r != '>' {
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
			// check to see which charset to use
			if t.useG1CharSet {
				t.handleOutputChar(charSetMap[t.g1Charset](r))

			} else {
				t.handleOutputChar(charSetMap[t.g0Charset](r))
			}
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

	t.content.SetCell(t.cursorRow, t.cursorCol, widget.TextGridCell{Rune: r, Style: cellStyle})
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
		if t.newLineMode {
			t.moveCursor(t.cursorRow, 0)
		}
		return
	}
	if t.newLineMode {
		t.moveCursor(t.cursorRow+1, 0)
		return
	}
	t.moveCursor(t.cursorRow+1, t.cursorCol)

}

func handleOutputTab(t *Terminal) {
	end := t.cursorCol - t.cursorCol%tabWidth + tabWidth
	for t.cursorCol < end {
		t.handleOutputChar(' ')
	}
}

func handleShiftOut(t *Terminal) {
	t.useG1CharSet = true
}

func handleShiftIn(t *Terminal) {
	t.useG1CharSet = false
}
