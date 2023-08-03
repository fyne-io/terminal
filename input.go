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

	if t.keyboardState.shiftPressed {
		t.KeyTypedWithShift(e)
		return
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
	case fyne.KeyF1:
		_, _ = t.in.Write([]byte{asciiEscape, 'O', 'P'})
	case fyne.KeyF2:
		_, _ = t.in.Write([]byte{asciiEscape, 'O', 'Q'})
	case fyne.KeyF3:
		_, _ = t.in.Write([]byte{asciiEscape, 'O', 'R'})
	case fyne.KeyF4:
		_, _ = t.in.Write([]byte{asciiEscape, 'O', 'S'})
	case fyne.KeyF5:
		_, _ = t.in.Write([]byte{asciiEscape, '[', '1', '5', '~'})
	case fyne.KeyF6:
		_, _ = t.in.Write([]byte{asciiEscape, '[', '1', '7', '~'})
	case fyne.KeyF7:
		_, _ = t.in.Write([]byte{asciiEscape, '[', '1', '8', '~'})
	case fyne.KeyF8:
		_, _ = t.in.Write([]byte{asciiEscape, '[', '1', '9', '~'})
	case fyne.KeyF9:
		_, _ = t.in.Write([]byte{asciiEscape, '[', '2', '0', '~'})
	case fyne.KeyF10:
		_, _ = t.in.Write([]byte{asciiEscape, '[', '2', '1', '~'})
	case fyne.KeyF11:
		_, _ = t.in.Write([]byte{asciiEscape, '[', '2', '3', '~'})
	case fyne.KeyF12:
		_, _ = t.in.Write([]byte{asciiEscape, '[', '2', '4', '~'})
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
	case fyne.KeyPageUp:
		_, _ = t.in.Write([]byte{asciiEscape, '[', '5', '~'})
	case fyne.KeyPageDown:
		_, _ = t.in.Write([]byte{asciiEscape, '[', '6', '~'})
	case fyne.KeyHome:
		_, _ = t.in.Write([]byte{asciiEscape, 'O', 'H'})
	case fyne.KeyInsert:
		_, _ = t.in.Write([]byte{asciiEscape, '[', '2', '~'})
	case fyne.KeyEnd:
		_, _ = t.in.Write([]byte{asciiEscape, 'O', 'F'})
	}

}

func (t *Terminal) KeyTypedWithShift(e *fyne.KeyEvent) {
	switch e.Name {
	case fyne.KeyF1:
		_, _ = t.in.Write([]byte{asciiEscape, '[', '2', '5', '~'})
	case fyne.KeyF2:
		_, _ = t.in.Write([]byte{asciiEscape, '[', '2', '6', '~'})
	case fyne.KeyF3:
		_, _ = t.in.Write([]byte{asciiEscape, 'O', 'R', ';', '2', '~'})
	case fyne.KeyF4:
		_, _ = t.in.Write([]byte{asciiEscape, '[', '1', ';', '2', 'S'})
	case fyne.KeyF5:
		_, _ = t.in.Write([]byte{asciiEscape, '[', '1', '5', ';', '2', '~'})
	case fyne.KeyF6:
		_, _ = t.in.Write([]byte{asciiEscape, '[', '1', '7', ';', '2', '~'})
	case fyne.KeyF7:
		_, _ = t.in.Write([]byte{asciiEscape, '[', '1', '8', ';', '2', '~'})
	case fyne.KeyF8:
		_, _ = t.in.Write([]byte{asciiEscape, '[', '1', '9', ';', '2', '~'})
	case fyne.KeyF9:
		_, _ = t.in.Write([]byte{asciiEscape, '[', '2', '0', ';', '2', '~'})
	case fyne.KeyF10:
		_, _ = t.in.Write([]byte{asciiEscape, '[', '2', '1', ';', '2', '~'})
	case fyne.KeyF11:
		_, _ = t.in.Write([]byte{asciiEscape, '[', '2', '3', ';', '2', '~'})
	case fyne.KeyF12:
		_, _ = t.in.Write([]byte{asciiEscape, '[', '2', '4', ';', '2', '~'})
	case fyne.KeyPageUp:
		_, _ = t.in.Write([]byte{asciiEscape, '[', '5', ';', '2', '~'})
	case fyne.KeyPageDown:
		_, _ = t.in.Write([]byte{asciiEscape, '[', '6', ';', '2', '~'})
	case fyne.KeyHome:
		_, _ = t.in.Write([]byte{asciiEscape, '[', '1', ';', '2', 'H'})
	case fyne.KeyInsert:
		_, _ = t.in.Write([]byte{asciiEscape, '[', '2', ';', '2', '~'})
	case fyne.KeyDelete:
		_, _ = t.in.Write([]byte{asciiEscape, '[', '3', ';', '2', '~'})
	case fyne.KeyEnd:
		_, _ = t.in.Write([]byte{asciiEscape, '[', '1', ';', '2', 'F'})
	case fyne.KeyUp:
		_, _ = t.in.Write([]byte{asciiEscape, '[', 'A', ';', '2'})
	case fyne.KeyDown:
		_, _ = t.in.Write([]byte{asciiEscape, '[', 'B', ';', '2'})
	case fyne.KeyLeft:
		_, _ = t.in.Write([]byte{asciiEscape, '[', 'D', ';', '2'})
	case fyne.KeyRight:
		_, _ = t.in.Write([]byte{asciiEscape, '[', 'C', ';', '2'})
	}
}

func (t *Terminal) TrackKeyboardState(down bool, e *fyne.KeyEvent) {
	switch e.Name {
	case desktop.KeyShiftLeft:
		t.keyboardState.shiftPressed = down
	case desktop.KeyAltLeft:
		t.keyboardState.altPressed = down
	case desktop.KeyControlLeft:
		t.keyboardState.ctrlPressed = down
	case desktop.KeyShiftRight:
		t.keyboardState.shiftPressed = down
	case desktop.KeyAltRight:
		t.keyboardState.altPressed = down
	case desktop.KeyControlRight:
		t.keyboardState.ctrlPressed = down
	}
}

func (t *Terminal) KeyDown(e *fyne.KeyEvent) {
	t.TrackKeyboardState(true, e)
}

func (t *Terminal) KeyUp(e *fyne.KeyEvent) {
	t.TrackKeyboardState(false, e)
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
