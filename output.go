package terminal

import (
	"bytes"
	"log"
	"time"
	"unicode/utf8"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	widget2 "github.com/fyne-io/terminal/internal/widget"
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

type parseState struct {
	code          string
	esc           int
	osc, apc, dcs bool
	vt100         rune
	printing      bool
}

func (t *Terminal) handleOutput(buf []byte) []byte {
	if t.hasSelectedText() {
		t.clearSelectedText()
	}
	if t.state == nil {
		t.state = &parseState{
			esc: noEscape,
		}
	}
	var (
		size int
		r    rune
		i    = -1
	)
	for {
		i += size
		buf = buf[size:]
		r, size = utf8.DecodeRune(buf)
		if size == 0 {
			break
		}
		if r == utf8.RuneError && size == 1 { // not UTF-8
			if !t.state.printing {
				if t.debug {
					log.Println("Invalid UTF-8", buf[0])
				}
				continue
			}
		}

		if t.state.printing {
			t.parsePrinting(buf, size)
			continue
		}
		if r == asciiEscape {
			t.state.esc = i
			continue
		}
		if t.state.dcs {
			t.parseDCS(r)
			continue
		}
		if t.state.esc == i-1 {
			if cont := t.parseEscState(r); cont {
				continue
			}
			t.state.esc = noEscape
			continue
		}
		if t.state.apc {
			t.parseAPC(r)
			continue
		}
		if t.state.osc {
			t.parseOSC(r)
			continue
		} else if t.state.vt100 != 0 {
			t.handleVT100(string([]rune{t.state.vt100, r}))
			t.state.vt100 = 0
			continue
		} else if t.state.esc != noEscape {
			t.parseEscape(r)
			continue
		}

		if out, ok := specialChars[r]; ok {
			if out == nil {
				continue
			}
			fyne.Do(func() {
				out(t)
			})
		} else {
			// check to see which charset to use
			if t.useG1CharSet {
				chr := charSetMap[t.g1Charset](r)
				fyne.Do(func() {
					t.handleOutputChar(chr)
				})
			} else {
				chr := charSetMap[t.g0Charset](r)
				fyne.Do(func() {
					t.handleOutputChar(chr)
				})
			}
		}
	}

	// record progress for next chunk of buffer
	if t.state.esc != noEscape {
		t.state.esc = t.state.esc - i
	}
	return buf
}

func (t *Terminal) parseEscState(r rune) (shouldContinue bool) {
	switch r {
	case '[':
		return true
	case '\\':
		if t.state.osc {
			code := t.state.code
			fyne.Do(func() {
				t.handleOSC(code)
			})
		}
		t.state.code = ""
		t.state.osc = false
	case ']':
		t.state.osc = true
	case '(', ')':
		t.state.vt100 = r
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
	case 'P':
		t.state.dcs = true
	case '_':
		t.state.apc = true
	case '=', '>':
	}
	return false
}

func (t *Terminal) parseEscape(r rune) {
	t.state.code += string(r)
	if (r < '0' || r > '9') && r != ';' && r != '=' && r != '?' && r != '>' {
		code := t.state.code
		fyne.Do(func() {
			t.handleEscape(code)
		})
		t.state.code = ""
		t.state.esc = noEscape
	}
}

func (t *Terminal) parsePrinting(buf []byte, size int) {
	t.printData = append(t.printData, buf[:size]...)
	if bytes.HasSuffix(t.printData, []byte{asciiEscape, '[', '4', 'i'}) {
		// Handle the end of printing
		t.printData = t.printData[:len(t.printData)-4]
		escapePrinterMode(t, "4")
		t.state.esc = noEscape
	}
}

func (t *Terminal) parseAPC(r rune) {
	if r == 0 {
		code := t.state.code
		fyne.Do(func() {
			t.handleAPC(code)
		})
		t.state.code = ""
		t.state.apc = false
	} else {
		t.state.code += string(r)
	}
}

func (t *Terminal) parseOSC(r rune) {
	if r == asciiBell || r == 0 {
		code := t.state.code
		fyne.Do(func() {
			t.handleOSC(code)
		})
		t.state.code = ""
		t.state.osc = false
	} else {
		t.state.code += string(r)
	}
}

func (t *Terminal) parseDCS(r rune) {
	if r == '\\' {
		code := t.state.code
		fyne.Do(func() {
			t.handleDCS(code)
		})
		t.state.code = ""
		t.state.dcs = false
	} else {
		t.state.code += string(r)
	}
}

func (t *Terminal) handleOutputChar(r rune) {
	if t.cursorCol == int(t.config.Columns) {
		if !t.disableAutoWrap {
			t.cursorCol = 0
			handleOutputLineFeed(t)
		} else {
			// In non-wrap mode, overwrite the last character
			t.cursorCol = int(t.config.Columns) - 1
		}
	}

	var cellStyle widget.TextGridStyle
	cellStyle = &widget.CustomTextGridStyle{FGColor: t.currentFG, BGColor: t.currentBG}
	if t.blinking {
		cellStyle = widget2.NewTermTextGridStyle(t.currentFG, t.currentBG, highlightBitMask, t.blinking)
	}

	row, col := t.cursorRow, t.cursorCol
	cell := widget.TextGridCell{Rune: r, Style: cellStyle}
	oldLen := 0
	if len(t.content.Rows) > row {
		oldLen = len(t.content.Rows[row].Cells)
	}
	t.content.SetCell(row, col, cell)

	for i := oldLen; i < col; i++ {
		if t.content.Rows[row].Cells[i].Rune == 0 {
			t.content.Rows[row].Cells[i].Rune = ' '
		}
	}
	t.cursorCol++
}

func (t *Terminal) ringBell() {
	t.bell = true
	t.Refresh()

	go func() {
		time.Sleep(time.Millisecond * 300)
		t.bell = false
		fyne.Do(t.Refresh)
	}()
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
	t.ringBell()
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

// SetPrinterFunc sets the printer function which is executed when printing.
func (t *Terminal) SetPrinterFunc(printerFunc PrinterFunc) {
	t.printer = printerFunc
}
