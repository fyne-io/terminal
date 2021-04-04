package terminal

import (
	"image/color"
	"math"
	"os"
	"os/exec"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"

	"github.com/creack/pty"
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
	content      *widget.TextGrid
	config       Config
	listenerLock sync.Mutex
	listeners    []chan Config

	pty                  *os.File
	focused, bell        bool
	cursorRow, cursorCol int
	homeRow, homeCol     int
	savedRow, savedCol   int
	cursorMoved          func()
}

// AddListener registers a new outgoing channel that will have our Config sent each time it changes.
func (t *Terminal) AddListener(listener chan Config) {
	t.listenerLock.Lock()
	defer t.listenerLock.Unlock()

	t.listeners = append(t.listeners, listener)
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

	t.config.Columns = uint(math.Floor(float64(s.Width) / float64(cellSize.Width)))
	t.config.Rows = uint(math.Floor(float64(s.Height) / float64(cellSize.Height)))
	t.onConfigure()

	t.updatePTYSize()
}

func (t *Terminal) updatePTYSize() {
	scale := float32(1.0)
	c := fyne.CurrentApp().Driver().CanvasForObject(t)
	if c != nil {
		scale = c.Scale()
	}
	_ = pty.Setsize(t.pty, &pty.Winsize{
		Rows: uint16(t.config.Rows), Cols: uint16(t.config.Columns),
		X: uint16(float32(t.Size().Width) * scale), Y: uint16(float32(t.Size().Height) * scale)})
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
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "bash"
	}
	env := os.Environ()
	env = append(env, "TERM=xterm-256color")
	c := exec.Command(shell)
	c.Env = env

	// Start the command with a pty.
	handle, err := pty.Start(c)
	if err != nil {
		return err
	}
	t.pty = handle

	t.updatePTYSize()
	return nil
}

// Exit requests that this terminal exits.
// If there are embedded shells it will exit the child one only.
func (t *Terminal) Exit() {
	_, _ = t.pty.Write([]byte("exit\n"))
}

func (t *Terminal) close() error {
	return t.pty.Close()
}

// don't call often - should we cache?
func (t *Terminal) guessCellSize() fyne.Size {
	cell := canvas.NewText("M", color.White)
	cell.TextStyle.Monospace = true

	return cell.MinSize()
}

func (t *Terminal) run() {
	bufLen := 4069
	buf := make([]byte, bufLen)
	for {
		num, err := t.pty.Read(buf)
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

// Run starts the terminal by loading a shell and starting to process the input/output
func (t *Terminal) Run() error {
	err := t.open()
	if err != nil {
		return err
	}

	t.run()

	return t.close()
}

// NewTerminal sets up a new terminal instance with the bash shell
func NewTerminal() *Terminal {
	t := &Terminal{}
	t.ExtendBaseWidget(t)
	t.content = widget.NewTextGrid()

	return t
}
