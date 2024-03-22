package terminal

import (
	"log"
	"strings"
)

// APCHandler handles a APC command for the given terminal.
type APCHandler func(*Terminal, string)

var apcHandlers = map[string]func(*Terminal, string){}

func (t *Terminal) handleAPC(code string) {
	for apcCommand, handler := range apcHandlers {
		if strings.HasPrefix(code, apcCommand) {
			// Extract the argument from the code
			arg := code[len(apcCommand):]
			// Invoke the corresponding handler function
			handler(t, arg)
			return
		}
	}

	if t.debug {
		// Handle other APC sequences or log the received APC code
		log.Println("Unrecognised APC", code)
	}
}

// RegisterAPCHandler registers a APC handler for the given APC command string.
func RegisterAPCHandler(APC string, handler APCHandler) {
	apcHandlers[APC] = handler
}
