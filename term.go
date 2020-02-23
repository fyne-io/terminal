package terminal

import (
	"log"
	"os"
	"os/exec"

	"fyne.io/fyne"
	"fyne.io/fyne/widget"

	"github.com/creack/pty"
	termCrypt "golang.org/x/crypto/ssh/terminal"
)

type terminal struct {
	content *widget.TextGrid

	pty      *os.File
	oldState *termCrypt.State
}

func (t *terminal) handleOSC(code string) {
	log.Println("Unrecognised OSC:", code)
}

func (t *terminal) handleEscape(code string) {
	switch code {
	case "2J":
		t.content.SetText("")
	case "K":
		runes := []rune(t.content.Text())
		t.content.SetText(string(runes[:len(runes)-1]))
	default:
		log.Println("Unrecognised Escape:", code)
	}
}

func (t *terminal) open() error {
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

func (t *terminal) close() error {
	_ = termCrypt.Restore(int(os.Stdin.Fd()), t.oldState) // Best effort.

	return t.pty.Close()
}

func (t *terminal) run(c fyne.Canvas) {
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
		}
	})

	buf := make([]byte, 1024)
	for {
		num, err := t.pty.Read(buf)
		if err != nil {
			fyne.LogError("output err", err)
			break // presuming broken pipe
		}

		t.handleOutput(buf[:num])
	}
}

func (t *terminal) handleOutput(buf []byte) {
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
				t.handleOSC(string(buf[2:]))
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
		case '\r':
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

// TODO remove canvas param!
func (t *terminal) Run(c fyne.Canvas) error {
	err := t.open()
	if err != nil {
		return err
	}

	t.run(c)

	return t.close()
}

func (t *terminal) BuildUI() fyne.CanvasObject { // TODO fix by having terminal a widget
	return t.content
}

func NewTerminal() *terminal {
	t := &terminal{}
	t.content = widget.NewTextGrid("")

	return t
}
