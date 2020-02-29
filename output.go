package terminal

import (
	"time"
)

func (t *Terminal) handleOutput(buf []byte) {
	out := ""
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
				t.handleOSC(out)
				out = ""
				osc = false
				continue
			}
		} else if vt100 != 0 {
			t.handleVT100(string([]rune{vt100, r}))
			vt100 = 0
			continue
		} else if esc != -5 {
			if (r >= '0' && r <= '9') || r == ';' || r == '=' || r == '?' {
				code += string(r)
			} else {
				code += string(r)

				t.handleEscape(code)
				code = ""
				esc = -5
			}
			continue
		}

		switch r {
		case asciiBackspace:
			row := t.content.Row(t.cursorRow)
			if len(row) == 0 {
				continue
			}
			t.content.SetRow(t.cursorRow, row[:len(row)-1])
			t.cursorCol--
			t.cursorMoved()
			continue
		case '\n':
			row := t.content.Row(t.cursorRow)
			row = append(row, []rune(out)...)
			t.content.SetRow(t.cursorRow, row)

			if t.cursorRow == int(t.config.Rows-1) {
				for i = 0; i < t.cursorRow; i++ {
					t.content.SetRow(i, t.content.Row(i+1))
				}
				t.content.SetRow(i, []rune{})
			} else {
				t.cursorRow++
			}

			out = ""
			t.cursorCol = 0
			continue
		case '\r':
			t.cursorCol = 0
			continue
		case asciiBell:
			go t.ringBell()
			continue
		case '\t': // TODO remove silly approximation
			out += "    "
		default:
			out += string(r)
		}
		esc = -5
		code = ""
	}

	if osc {
		t.handleOSC(out)
		return
	}
	row := t.content.Row(t.cursorRow)
	row = append(row, []rune(out)...)
	t.content.SetRow(t.cursorRow, row)
	t.cursorCol += len(out)
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
