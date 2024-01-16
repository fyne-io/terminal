package terminal

import (
	"bytes"
	"io"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
)

// NopCloser returns a WriteCloser with a no-op Close method wrapping
// the provided writer w.
func NopCloser(w io.Writer) io.WriteCloser {
	return nopCloser{w}
}

type nopCloser struct {
	io.Writer
}

func (n nopCloser) Close() error {
	return nil
}

func TestTerminal_TypedKey(t *testing.T) {
	tests := map[string]struct {
		key          fyne.KeyName
		bufferMode   bool
		shiftPressed bool
		want         []byte
	}{
		"F1":             {fyne.KeyF1, false, false, []byte{asciiEscape, 'O', 'P'}},
		"F2":             {fyne.KeyF2, false, false, []byte{asciiEscape, 'O', 'Q'}},
		"F3":             {fyne.KeyF3, false, false, []byte{asciiEscape, 'O', 'R'}},
		"F4":             {fyne.KeyF4, false, false, []byte{asciiEscape, 'O', 'S'}},
		"F5":             {fyne.KeyF5, false, false, []byte{asciiEscape, '[', '1', '5', '~'}},
		"F6":             {fyne.KeyF6, false, false, []byte{asciiEscape, '[', '1', '7', '~'}},
		"F7":             {fyne.KeyF7, false, false, []byte{asciiEscape, '[', '1', '8', '~'}},
		"F8":             {fyne.KeyF8, false, false, []byte{asciiEscape, '[', '1', '9', '~'}},
		"F9":             {fyne.KeyF9, false, false, []byte{asciiEscape, '[', '2', '0', '~'}},
		"F10":            {fyne.KeyF10, false, false, []byte{asciiEscape, '[', '2', '1', '~'}},
		"F11":            {fyne.KeyF11, false, false, []byte{asciiEscape, '[', '2', '3', '~'}},
		"F12":            {fyne.KeyF12, false, false, []byte{asciiEscape, '[', '2', '4', '~'}},
		"Shift+F1":       {fyne.KeyF1, false, true, []byte{asciiEscape, '[', '2', '5', '~'}},
		"Shift+F2":       {fyne.KeyF2, false, true, []byte{asciiEscape, '[', '2', '6', '~'}},
		"Shift+F3":       {fyne.KeyF3, false, true, []byte{asciiEscape, 'O', 'R', ';', '2', '~'}},
		"Shift+F4":       {fyne.KeyF4, false, true, []byte{asciiEscape, '[', '1', ';', '2', 'S'}},
		"Shift+F5":       {fyne.KeyF5, false, true, []byte{asciiEscape, '[', '1', '5', ';', '2', '~'}},
		"Shift+F6":       {fyne.KeyF6, false, true, []byte{asciiEscape, '[', '1', '7', ';', '2', '~'}},
		"Shift+F7":       {fyne.KeyF7, false, true, []byte{asciiEscape, '[', '1', '8', ';', '2', '~'}},
		"Shift+F8":       {fyne.KeyF8, false, true, []byte{asciiEscape, '[', '1', '9', ';', '2', '~'}},
		"Shift+F9":       {fyne.KeyF9, false, true, []byte{asciiEscape, '[', '2', '0', ';', '2', '~'}},
		"Shift+F10":      {fyne.KeyF10, false, true, []byte{asciiEscape, '[', '2', '1', ';', '2', '~'}},
		"Shift+F11":      {fyne.KeyF11, false, true, []byte{asciiEscape, '[', '2', '3', ';', '2', '~'}},
		"Shift+F12":      {fyne.KeyF12, false, true, []byte{asciiEscape, '[', '2', '4', ';', '2', '~'}},
		"Shift+PageUp":   {fyne.KeyPageUp, false, true, []byte{asciiEscape, '[', '5', ';', '2', '~'}},
		"Shift+PageDown": {fyne.KeyPageDown, false, true, []byte{asciiEscape, '[', '6', ';', '2', '~'}},
		"Shift+Home":     {fyne.KeyHome, false, true, []byte{asciiEscape, '[', '1', ';', '2', 'H'}},
		"Shift+Insert":   {fyne.KeyInsert, false, true, []byte{asciiEscape, '[', '2', ';', '2', '~'}},
		"Shift+Delete":   {fyne.KeyDelete, false, true, []byte{asciiEscape, '[', '3', ';', '2', '~'}},
		"Shift+End":      {fyne.KeyEnd, false, true, []byte{asciiEscape, '[', '1', ';', '2', 'F'}},
		"Shift+Up":       {fyne.KeyUp, false, true, []byte{asciiEscape, '[', 'A', ';', '2'}},
		"Shift+Down":     {fyne.KeyDown, false, true, []byte{asciiEscape, '[', 'B', ';', '2'}},
		"Shift+Left":     {fyne.KeyLeft, false, true, []byte{asciiEscape, '[', 'D', ';', '2'}},
		"Shift+Right":    {fyne.KeyRight, false, true, []byte{asciiEscape, '[', 'C', ';', '2'}},

		"PageUp":    {fyne.KeyPageUp, false, false, []byte{asciiEscape, '[', '5', '~'}},
		"PageDown":  {fyne.KeyPageDown, false, false, []byte{asciiEscape, '[', '6', '~'}},
		"Home":      {fyne.KeyHome, false, false, []byte{asciiEscape, 'O', 'H'}},
		"Insert":    {fyne.KeyInsert, false, false, []byte{asciiEscape, '[', '2', '~'}},
		"Delete":    {fyne.KeyDelete, false, false, []byte{asciiEscape, '[', '3', '~'}},
		"End":       {fyne.KeyEnd, false, false, []byte{asciiEscape, 'O', 'F'}},
		"Enter":     {fyne.KeyEnter, false, false, []byte{'\n'}}, // Modify as needed for Windows or bufferMode
		"Tab":       {fyne.KeyTab, false, false, []byte{'\t'}},
		"Escape":    {fyne.KeyEscape, false, false, []byte{asciiEscape}},
		"Backspace": {fyne.KeyBackspace, false, false, []byte{asciiBackspace}},
		"Up":        {fyne.KeyUp, false, false, []byte{asciiEscape, '[', 'A'}},
		"Down":      {fyne.KeyDown, false, false, []byte{asciiEscape, '[', 'B'}},
		"Left":      {fyne.KeyLeft, false, false, []byte{asciiEscape, '[', 'D'}},
		"Right":     {fyne.KeyRight, false, false, []byte{asciiEscape, '[', 'C'}},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// Creating a mock terminal
			inBuffer := bytes.NewBuffer([]byte{})
			term := &Terminal{in: NopCloser(inBuffer), bufferMode: tt.bufferMode}
			term.keyboardState.shiftPressed = tt.shiftPressed
			keyEvent := &fyne.KeyEvent{Name: tt.key}

			term.TypedKey(keyEvent)

			got := inBuffer.Bytes()
			if !bytes.Equal(got, tt.want) {
				t.Errorf("TypedKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTerminal_TypedKey_LineMode(t *testing.T) {
	tests := map[string]struct {
		key         fyne.KeyName
		newLineMode bool
		want        []byte
	}{

		"Enter":                 {fyne.KeyEnter, false, []byte{'\n'}},
		"Enter with line mode":  {fyne.KeyEnter, true, []byte{'\r'}},
		"Return":                {fyne.KeyReturn, false, []byte{'\r'}},
		"Return with line mode": {fyne.KeyReturn, true, []byte{'\r'}},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// Creating a mock terminal
			inBuffer := bytes.NewBuffer([]byte{})
			term := &Terminal{in: NopCloser(inBuffer), newLineMode: tt.newLineMode}
			keyEvent := &fyne.KeyEvent{Name: tt.key}

			term.TypedKey(keyEvent)

			got := inBuffer.Bytes()
			if !bytes.Equal(got, tt.want) {
				t.Errorf("TypedKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTerminal_TypedShortcut(t *testing.T) {
	tests := map[string]struct {
		shortcut fyne.Shortcut
		want     []byte
	}{
		"LeftOption+U": {
			shortcut: &desktop.CustomShortcut{
				Modifier: fyne.KeyModifierAlt,
				KeyName:  fyne.KeyU},
			want: []byte{},
		},
		"Control+@": {
			shortcut: &desktop.CustomShortcut{
				Modifier: fyne.KeyModifierControl,
				KeyName:  "@"},
			want: []byte{0},
		},
		"Control+Space": {
			shortcut: &desktop.CustomShortcut{
				Modifier: fyne.KeyModifierControl,
				KeyName:  fyne.KeySpace},
			want: []byte{0},
		},
		"Control+C": {
			shortcut: &desktop.CustomShortcut{
				Modifier: fyne.KeyModifierControl,
				KeyName:  fyne.KeyC},
			want: []byte{3},
		},
		"Control+_": {
			shortcut: &desktop.CustomShortcut{
				Modifier: fyne.KeyModifierControl,
				KeyName:  "_"},
			want: []byte{31},
		},
		"Control+X": {
			shortcut: &desktop.CustomShortcut{
				Modifier: fyne.KeyModifierControl,
				KeyName:  fyne.KeyX},
			want: []byte{24},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// Creating a mock terminal
			inBuffer := bytes.NewBuffer([]byte{})
			term := &Terminal{in: NopCloser(inBuffer)}

			term.TypedShortcut(tt.shortcut)

			got := inBuffer.Bytes()
			if !bytes.Equal(got, tt.want) {
				t.Errorf("TypedShortcut() = %v, want %v", got, tt.want)
			}
		})
	}
}
