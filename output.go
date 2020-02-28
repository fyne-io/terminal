package terminal

import (
	"strings"
	"time"
)

func (t *Terminal) handleOutput(buf []byte) {
	out := ""
	esc := -5
	code := ""
	for i, r := range buf {
		if r == asciiEscape {
			esc = i
			continue
		}
		if esc == i-1 {
			if r == '[' {
				continue
			} else if r == ']' {
				// TODO only up to BEL or ST
				t.handleOSC(string(buf[2 : len(buf)-1]))
				break
			} else {
				esc = -5
			}
		}
		if esc != -5 {
			if (r >= '0' && r <= '9') || r == ';' || r == '=' {
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
			runes := []rune(t.content.Text())
			if len(runes) == 0 {
				continue
			}
			t.content.SetText(string(runes[:len(runes)-1]))
			continue
		case '\r':
			continue
		case asciiBell:
			go t.bell()
			continue
		case '\t': // TODO remove silly approximation
			out += "    "
		default:
			out += string(r)
		}
		esc = -5
		code = ""
	}
	t.content.SetText(t.content.Text() + out)
}

func (t *Terminal) bell() {
	add := "*BELL* "
	title := t.config.Title
	if strings.Index(title, add) == 0 { // don't ring twice at once
		return
	}

	t.config.Title = add + title
	t.onConfigure()
	select {
	case <-time.After(time.Millisecond * 300):
		t.config.Title = title
		t.onConfigure()
	}
}
