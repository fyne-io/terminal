package terminal

import (
	"runtime"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
)

// TypedRune is called when the user types a visible character
func (t *Terminal) TypedRune(r rune) {
	_, _ = t.in.Write([]byte{byte(r)})
}

// TypedKey will be called if a non-printable keyboard event occurs
func (t *Terminal) TypedKey(e *fyne.KeyEvent) {
	cursorPrefix := byte('[')
	if t.bufferMode {
		cursorPrefix = 'O'
	}
	switch e.Name {
	case fyne.KeyEnter, fyne.KeyReturn:
		if t.bufferMode || runtime.GOOS == "windows" {
			_, _ = t.in.Write([]byte{'\r'})
		} else {
			_, _ = t.in.Write([]byte{'\n'})
		}
	case fyne.KeyTab:
		_, _ = t.in.Write([]byte{'\t'})
	case fyne.KeyEscape:
		_, _ = t.in.Write([]byte{asciiEscape})
	case fyne.KeyBackspace:
		_, _ = t.in.Write([]byte{asciiBackspace})
	case fyne.KeyDelete:
		_, _ = t.in.Write([]byte{asciiEscape, '[', '3', '~'})
	case fyne.KeyUp:
		_, _ = t.in.Write([]byte{asciiEscape, cursorPrefix, 'A'})
	case fyne.KeyDown:
		_, _ = t.in.Write([]byte{asciiEscape, cursorPrefix, 'B'})
	case fyne.KeyLeft:
		_, _ = t.in.Write([]byte{asciiEscape, cursorPrefix, 'D'})
	case fyne.KeyRight:
		_, _ = t.in.Write([]byte{asciiEscape, cursorPrefix, 'C'})
	}
}

// FocusGained notifies the terminal that it has focus
func (t *Terminal) FocusGained() {
	t.focused = true
	t.Refresh()
}

// TypedShortcut handles key combinations, we pass them on to the tty.
func (t *Terminal) TypedShortcut(s fyne.Shortcut) {
	// we need to override the default cut/copy/paste and do it ourselves
	if _, ok := s.(*fyne.ShortcutCut); ok {
		_, _ = t.in.Write([]byte{0x18})
	} else if _, ok := s.(*fyne.ShortcutCopy); ok {
		_, _ = t.in.Write([]byte{0x3})
	} else if _, ok := s.(*fyne.ShortcutPaste); ok {
		_, _ = t.in.Write([]byte{0x16})
	} else if _, ok := s.(*fyne.ShortcutSelectAll); ok {
		_, _ = t.in.Write([]byte{0x1})
	} else if ds, ok := s.(*desktop.CustomShortcut); ok {
		t.ShortcutHandler.TypedShortcut(s) // it's not clear how we can check if this consumed the event

		off := ds.KeyName[0] - 'A' + 1
		_, _ = t.in.Write([]byte{off})
	}
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
