package terminal

import "fyne.io/fyne"

// TypedRune is called when the user types a visible character
func (t *Terminal) TypedRune(r rune) {
	_, _ = t.pty.WriteString(string(r))
}

// TypedKey will be called if a non-printable keyboard event occurs
func (t *Terminal) TypedKey(e *fyne.KeyEvent) {
	switch e.Name {
	case fyne.KeyEnter, fyne.KeyReturn:
		_, _ = t.pty.Write([]byte{'\n'})
	case fyne.KeyEscape:
		_, _ = t.pty.Write([]byte{asciiEscape})
	case fyne.KeyBackspace:
		_, _ = t.pty.Write([]byte{asciiBackspace})
	case fyne.KeyUp:
		_, _ = t.pty.Write([]byte{asciiEscape, '[', 'A'})
	case fyne.KeyDown:
		_, _ = t.pty.Write([]byte{asciiEscape, '[', 'B'})
	}
}

// FocusGained notifies the terminal that it has focus
func (t *Terminal) FocusGained() {
	t.focused = true
	t.Refresh()
}

// FocusLost tells the terminal it no longer has focus
func (t *Terminal) FocusLost() {
	t.focused = false
	t.Refresh()
}

// Focused is used to determine if this terminal currently has focus
func (t *Terminal) Focused() bool {
	return t.focused
}
