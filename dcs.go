package terminal

import (
	"encoding/hex"
	"log"
)

func (t *Terminal) handleDCS(code string) {
	if len(code) >= 2 && code[:2] == "+q" {
		query, _ := hex.DecodeString(code[2:]) // strip the +q
		if t.debug {
			log.Println("unhandled DCS query", string(query))
		}

		// return not recognised - TODO actually return results
		_, _ = t.in.Write([]byte{asciiEscape, 'P', '0', '+', 'r', asciiEscape, '\\'})
	} else {
		if t.debug {
			log.Println("unknown DCS query", code)
		}
	}
}
