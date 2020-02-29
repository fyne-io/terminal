package terminal

import "log"

func (t *Terminal) handleEscape(code string) {
	switch code {
	case "2J":
		t.clearScreen()
	case "K":
		row := t.content.Row(t.cursorRow)
		t.content.SetRow(t.cursorRow, row[:t.cursorCol])
	default:
		log.Println("Unrecognised Escape:", code)
	}
}

func (t *Terminal) clearScreen() {
	t.content.SetText("")
	t.cursorCol = 0
	t.cursorRow = 0
}
