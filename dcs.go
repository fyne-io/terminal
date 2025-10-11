package terminal

import (
	"encoding/hex"
	"log"
)

func (t *Terminal) handleDCS(code string) {
	if len(code) >= 2 && code[:2] == "+q" {
		query, _ := hex.DecodeString(code[2:]) // strip the +q
		if t.debug {
			log.Println("unhandled DCS query", query)
		}

		_, _ = t.in.Write([]byte{asciiEscape})
		_, _ = t.in.Write([]byte("P0+r")) // return not recognised - TODO actually return results
		_, _ = t.in.Write([]byte{asciiEscape, '\\', 0})
	} else {
		if t.debug {
			log.Println("unknown DCS query", code)
		}
	}
}
