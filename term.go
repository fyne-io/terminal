package terminal

import (
	"image/color"
	"io"
	"math"
	"os"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/driver/mobile"
	"fyne.io/fyne/v2/widget"
)

// Config is the state of a terminal, updated upon certain actions or commands.
// Use Terminal.OnConfigure hook to register for changes.
type Config struct {
	Title         string
	Rows, Columns uint
}

// Terminal is a terminal widget that loads a shell and handles input/output.
type Terminal struct {
	widget.BaseWidget
	fyne.ShortcutHandler
	content      *widget.TextGrid
	config       Config
	listenerLock sync.Mutex
	listeners    []chan Config
	startDir     string

	pty io.Closer
	in  io.WriteCloser
	out io.Reader

	bell, bright, debug, focused bool
	currentFG, currentBG         color.Color
	cursorRow, cursorCol         int
	savedRow, savedCol           int
	scrollTop, scrollBottom      int

	cursor                   *canvas.Rectangle
	cursorHidden, bufferMode bool // buffer mode is an xterm extension that impacts control keys
	cursorMoved              func()

	onMouseDown, onMouseUp func(int, fyne.KeyModifier, fyne.Position)
}

// AcceptsTab indicates that this widget will use the Tab key (avoids loss of focus).
func (t *Terminal) AcceptsTab() bool {
	return true
}

// AddListener registers a new outgoing channel that will have our Config sent each time it changes.
func (t *Terminal) AddListener(listener chan Config) {
	t.listenerLock.Lock()
	defer t.listenerLock.Unlock()

	t.listeners = append(t.listeners, listener)
}

// MinSize provides a size large enough that a terminal could technically funcion.
func (t *Terminal) MinSize() fyne.Size {
	s := t.guessCellSize()
	return fyne.NewSize(s.Width*2.5, s.Height*1.2) // just enough to get a terminal init
}

// MouseDown handles the down action for desktop mouse events.
func (t *Terminal) MouseDown(ev *desktop.MouseEvent) {
	if t.onMouseDown == nil {
		return
	}

	if ev.Button == desktop.MouseButtonPrimary {
		t.onMouseDown(1, ev.Modifier, ev.Position)
	} else if ev.Button == desktop.MouseButtonSecondary {
		t.onMouseDown(2, ev.Modifier, ev.Position)
	}
}

// MouseUp handles the up action for desktop mouse events.
func (t *Terminal) MouseUp(ev *desktop.MouseEvent) {
	if t.onMouseDown == nil {
		return
	}

	if ev.Button == desktop.MouseButtonPrimary {
		t.onMouseUp(1, ev.Modifier, ev.Position)
	} else if ev.Button == desktop.MouseButtonSecondary {
		t.onMouseUp(2, ev.Modifier, ev.Position)
	}
}

// RemoveListener de-registers a Config channel and closes it
func (t *Terminal) RemoveListener(listener chan Config) {
	t.listenerLock.Lock()
	defer t.listenerLock.Unlock()

	for i, l := range t.listeners {
		if l == listener {
			if i < len(t.listeners)-1 {
				t.listeners = append(t.listeners[:i], t.listeners[i+1:]...)
			} else {
				t.listeners = t.listeners[:i]
			}
			close(l)
			return
		}
	}
}

// Resize is called when this terminal widget has been resized.
// It ensures that the virtual terminal is within the bounds of the widget.
func (t *Terminal) Resize(s fyne.Size) {
	if s.Width == t.Size().Width && s.Height == t.Size().Height {
		return
	}
	if s.Width < 20 { // not sure why we get tiny sizes
		return
	}
	t.BaseWidget.Resize(s)
	t.content.Resize(s)

	cellSize := t.guessCellSize()
	oldRows := int(t.config.Rows)

	t.config.Columns = uint(math.Floor(float64(s.Width) / float64(cellSize.Width)))
	t.config.Rows = uint(math.Floor(float64(s.Height) / float64(cellSize.Height)))
	if t.scrollBottom == 0 || t.scrollBottom == oldRows-1 {
		t.scrollBottom = int(t.config.Rows) - 1
	}
	t.onConfigure()

	go t.updatePTYSize()
}

// SetDebug turns on output about terminal codes and other errors if the parameter is `true`.
func (t *Terminal) SetDebug(debug bool) {
	t.debug = debug
}

// SetStartDir can be called before one of the Run calls to specify the initial directory.
func (t *Terminal) SetStartDir(path string) {
	t.startDir = path
}

// Tapped makes sure we ask for focus if user taps us.
func (t *Terminal) Tapped(ev *fyne.PointEvent) {
	fyne.CurrentApp().Driver().CanvasForObject(t).Focus(t)
}

// TouchCancel handles the tap action for mobile apps that lose focus during tap.
func (t *Terminal) TouchCancel(ev *mobile.TouchEvent) {
	if t.onMouseUp != nil {
		t.onMouseUp(1, 0, ev.Position)
	}
}

// TouchDown handles the down action for mobile touch events.
func (t *Terminal) TouchDown(ev *mobile.TouchEvent) {
	if t.onMouseDown != nil {
		t.onMouseDown(1, 0, ev.Position)
	}
}

// TouchUp handles the up action for mobile touch events.
func (t *Terminal) TouchUp(ev *mobile.TouchEvent) {
	if t.onMouseUp != nil {
		t.onMouseUp(1, 0, ev.Position)
	}
}

func (t *Terminal) onConfigure() {
	t.listenerLock.Lock()
	for _, l := range t.listeners {
		select {
		case l <- t.config:
		default:
			// channel blocked, might be closed
		}
	}
	t.listenerLock.Unlock()
}

func (t *Terminal) open() error {
	in, out, pty, err := t.startPTY()
	if err != nil {
		return err
	}
	t.in = in
	t.out = out
	t.pty = pty

	t.updatePTYSize()
	return nil
}

// Exit requests that this terminal exits.
// If there are embedded shells it will exit the child one only.
func (t *Terminal) Exit() {
	_, _ = t.Write([]byte{0x4})
}

func (t *Terminal) close() error {
	if t.in != t.pty {
		_ = t.in.Close() // we may already be closed
	}
	if t.pty == nil {
		return nil
	}

	return t.pty.Close()
}

// don't call often - should we cache?
func (t *Terminal) guessCellSize() fyne.Size {
	cell := canvas.NewText("M", color.White)
	cell.TextStyle.Monospace = true

	min := cell.MinSize()
	return fyne.NewSize(float32(math.Round(float64(min.Width))), float32(math.Round(float64(min.Height))))
}

func (t *Terminal) run() {
	bufLen := 4069
	buf := make([]byte, bufLen)
	for {
		num, err := t.out.Read(buf)
		if err != nil {
			// this is the pre-go 1.13 way to check for the read failing (terminal closed)
			if err.Error() == "EOF" {
				break // term exit on macOS
			} else if err, ok := err.(*os.PathError); ok && err.Err.Error() == "input/output error" {
				break // broken pipe, terminal exit
			}

			fyne.LogError("pty read error", err)
		}

		t.handleOutput(buf[:num])
		if num < bufLen {
			t.Refresh()
		}
	}
}

// RunLocalShell starts the terminal by loading a shell and starting to process the input/output.
func (t *Terminal) RunLocalShell() error {
	for t.config.Columns == 0 { // don't load the TTY until our output is configured
		time.Sleep(time.Millisecond * 50)
	}
	err := t.open()
	if err != nil {
		return err
	}

	t.run()

	return t.close()
}

// RunWithConnection starts the terminal by connecting to an external resource like an SSH connection.
func (t *Terminal) RunWithConnection(in io.WriteCloser, out io.Reader) error {
	for t.config.Columns == 0 { // don't load the TTY until our output is configured
		time.Sleep(time.Millisecond * 50)
	}
	t.in = in
	t.out = out

	t.run()

	return t.close()
}

// Write is used to send commands into an open terminal connection.
// Errors will be returned if the connection is not established, has closed, or there was a problem in transmission.
func (t *Terminal) Write(b []byte) (int, error) {
	if t.in == nil {
		return 0, io.EOF
	}

	return t.in.Write(b)
}

func (t *Terminal) setupShortcuts() {
	t.ShortcutHandler.AddShortcut(&desktop.CustomShortcut{KeyName: fyne.KeyV, Modifier: fyne.KeyModifierShift | fyne.KeyModifierControl},
		func(_ fyne.Shortcut) {
			a := fyne.CurrentApp()
			c := a.Driver().CanvasForObject(t)
			if c == nil {
				return
			}

			var win fyne.Window
			for _, w := range a.Driver().AllWindows() {
				if w.Canvas() == c {
					win = w
				}
			}
			if win == nil {
				return
			}

			_, _ = t.in.Write([]byte(win.Clipboard().Content()))
		})
}

func (t *Terminal) startingDir() string {
	if t.startDir == "" {
		home, err := os.UserHomeDir()
		if err == nil {
			return home
		}
	}

	return t.startDir
}

// New sets up a new terminal instance with the bash shell
func New() *Terminal {
	t := &Terminal{}
	t.ExtendBaseWidget(t)
	t.content = widget.NewTextGrid()
	t.setupShortcuts()

	return t
}
