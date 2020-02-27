package terminal

import (
	"image/color"
	"log"
	"math"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/widget"

	"github.com/creack/pty"
	termCrypt "golang.org/x/crypto/ssh/terminal"
)

// Config is the state of a terminal, updated upon certain actions or commands.
// Use Terminal.OnConfigure hook to register for changes.
type Config struct {
	Title         string
	Rows, Columns uint16
}

// Terminal is a terminal widget that loads a shell and handles input/output.
type Terminal struct {
	widget.BaseWidget
	content      *widget.TextGrid
	config       Config
	listenerLock sync.Mutex
	listeners    []chan Config

	pty      *os.File
	oldState *termCrypt.State
	focused  bool
}

// AddListener registers a new outgoing channel that will have our Config sent each time it changes.
func (t *Terminal) AddListener(listener chan Config) {
	t.listenerLock.Lock()
	t.listeners = append(t.listeners, listener)
	t.listenerLock.Unlock()
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

	t.config.Columns = uint16(math.Floor(float64(s.Width) / float64(cellSize.Width)))
	t.config.Rows = uint16(math.Floor(float64(s.Height) / float64(cellSize.Height)))
	t.onConfigure()

	scale := float32(1.0)
	c := fyne.CurrentApp().Driver().CanvasForObject(t)
	if c != nil {
		scale = c.Scale()
	}
	_ = pty.Setsize(t.pty, &pty.Winsize{
		Rows: t.config.Rows, Cols: t.config.Columns,
		X: uint16(float32(s.Width) * scale), Y: uint16(float32(s.Height) * scale)})
}

func (t *Terminal) bell() {
	add := "*BELL* "
	title := t.config.Title
	if strings.Index(title, add) == 0 { // don't ring twice at once
		return
	}

	t.config.Title = add + title
	t.onConfigure()
	select {
	case <-time.After(time.Millisecond * 300):
		t.config.Title = title
		t.onConfigure()
	}
}

func (t *Terminal) handleOSC(code string) {
	if len(code) > 2 && code[1] == ';' {
		switch code[0] {
		case '0':
			t.config.Title = code[2:]
			t.onConfigure()
		}
	} else {
		log.Println("Unrecognised OSC:", code)
	}
}

func (t *Terminal) handleEscape(code string) {
	switch code {
	case "2J":
		t.content.SetText("")
	case "K":
		t.content.SetText("")
		// TODO clear from the cursor to end line
	default:
		log.Println("Unrecognised Escape:", code)
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
	// Create shell command.
	c := exec.Command("bash")

	// Start the command with a pty.
	handle, err := pty.Start(c)
	if err != nil {
		return err
	}
	t.pty = handle

	// Set stdin in raw mode.
	t.oldState, err = termCrypt.MakeRaw(int(os.Stdin.Fd()))
	return err
}

func (t *Terminal) close() error {
	_ = termCrypt.Restore(int(os.Stdin.Fd()), t.oldState) // Best effort.

	return t.pty.Close()
}

// don't call often - should we cache?
func (t *Terminal) guessCellSize() fyne.Size {
	cell := canvas.NewText("M", color.White)
	cell.TextStyle.Monospace = true

	return cell.MinSize()
}

func (t *Terminal) run() {
	buf := make([]byte, 1024)
	for {
		num, err := t.pty.Read(buf)
		if err != nil {
			// this is the pre-go 1.13 way to check for the read failing (terminal closed)
			if err, ok := err.(*os.PathError); ok && err.Err.Error() == "input/output error" {
				break // broken pipe, terminal exit
			}

			fyne.LogError("pty read error", err)
		}

		t.handleOutput(buf[:num])
	}
}

func (t *Terminal) handleOutput(buf []byte) {
	out := ""
	esc := -5
	code := ""
	for i, r := range buf {
		if r == 0x1b {
			esc = i
			continue
		}
		if esc == i-1 {
			if r == '[' {
				continue
			} else if r == ']' {
				// TODO only up to BEL or ST
				t.handleOSC(string(buf[2 : len(buf)-1]))
				break
			} else {
				esc = -5
			}
		}
		if esc != -5 {
			if (r >= '0' && r <= '9') || r == ';' || r == '=' {
				code += string(r)
			} else {
				code += string(r)

				t.handleEscape(code)
				code = ""
				esc = -5
			}
			continue
		}

		switch r {
		case 8: // Backspace
			runes := []rune(t.content.Text())
			if len(runes) == 0 {
				continue
			}
			t.content.SetText(string(runes[:len(runes)-1]))
			continue
		case '\r':
			continue
		case 7: // Bell
			go t.bell()
			continue
		case '\t': // TODO remove silly approximation
			out += "    "
		default:
			out += string(r)
		}
		esc = -5
		code = ""
	}
	t.content.SetText(t.content.Text() + out)
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
	t.content = widget.NewTextGrid("")

	return t
}
