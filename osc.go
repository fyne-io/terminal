package terminal

import "log"

func (t *Terminal) handleOSC(code string) {
	if len(code) > 2 && code[1] == ';' {
		switch code[0] {
		case '0':
			t.config.Title = code[2:]
			t.onConfigure()
		}
	} else {
		log.Println("Unrecognised OSC:", code)
	}
}
