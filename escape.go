package terminal

import "log"

func (t *Terminal) handleEscape(code string) {
	switch code {
	case "2J":
		t.clearScreen()
	case "K":
		t.content.SetText("")
		// TODO clear from the cursor to end line
	default:
		log.Println("Unrecognised Escape:", code)
	}
}

func (t *Terminal) clearScreen() {
	t.content.SetText("")
}
