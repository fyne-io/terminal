package terminal

import (
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"fyne.io/fyne"
	"fyne.io/fyne/widget"

	"github.com/creack/pty"
	termCrypt "golang.org/x/crypto/ssh/terminal"
)

// Config is the state of a terminal, updated upon certain actions or commands.
// Use Terminal.OnConfigure hook to register for changes.
type Config struct {
	Title string
}

// Terminal is a terminal widget that loads a shell and handles input/output.
type Terminal struct {
	content     *widget.TextGrid
	Config      Config
	OnConfigure func()

	pty      *os.File
	oldState *termCrypt.State
}

func (t *Terminal) bell() {
	if t.OnConfigure == nil {
		return // we're going to configure the title to "ring the bell"
	}

	add := "*BELL* "
	title := t.Config.Title
	if strings.Index(title, add) == 0 { // don't ring twice at once
		return
	}

	t.Config.Title = add + title
	t.OnConfigure()
	select {
	case <-time.After(time.Millisecond * 300):
		t.Config.Title = title
		t.OnConfigure()
	}
}

func (t *Terminal) handleOSC(code string) {
	if len(code) > 2 && code[1] == ';' {
		switch code[0] {
		case '0':
			t.Config.Title = code[2:]
			if t.OnConfigure != nil {
				t.OnConfigure()
			}
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
		// TODO clear from the cursor to end line
	default:
		log.Println("Unrecognised Escape:", code)
	}
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

func (t *Terminal) run(c fyne.Canvas) {
	// TODO fit to window size...
	size := &pty.Winsize{Cols: 80, Rows: 24}
	_ = pty.Setsize(t.pty, size)

	// TODO move to (Focusable) widget
	c.SetOnTypedRune(func(r rune) {
		_, _ = t.pty.WriteString(string(r))
	})
	c.SetOnTypedKey(func(e *fyne.KeyEvent) {
		switch e.Name {
		case fyne.KeyEnter, fyne.KeyReturn:
			_, _ = t.pty.Write([]byte{'\n'})
		case fyne.KeyBackspace:
			_, _ = t.pty.Write([]byte{8})
		case fyne.KeyUp:
			_, _ = t.pty.Write([]byte{0x1b, '[', 'A'})
		case fyne.KeyDown:
			_, _ = t.pty.Write([]byte{0x1b, '[', 'B'})
		}
	})

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
func (t *Terminal) Run(c fyne.Canvas) error { // TODO remove canvas param!
	err := t.open()
	if err != nil {
		return err
	}

	t.run(c)

	return t.close()
}

// BuildUI returns the interface component of this terminal
func (t *Terminal) BuildUI() fyne.CanvasObject { // TODO fix by having terminal a widget
	return t.content
}

// NewTerminal sets up a new terminal instance with the bash shell
func NewTerminal() *Terminal {
	t := &Terminal{}
	t.content = widget.NewTextGrid("")

	return t
}
