package terminal

import (
	"fmt"
	"image/color"
	"log"
	"os"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
)

const (
	oscTerminatorBell = "\a"
	oscTerminatorST   = "\x1b\\"
)

func (t *Terminal) handleOSC(code string) {
	t.handleOSCTerm(code, oscTerminatorBell)
}

// handleOSCTerm processes an OSC command, terminator is the string that ended
// the sequence so that any reply can be terminated in the same way.
func (t *Terminal) handleOSCTerm(code, terminator string) {
	cmd, arg, ok := strings.Cut(code, ";")
	if !ok {
		return
	}

	switch cmd {
	case "0":
		// set icon name, if Fyne supports in the future
		t.setTitle(arg)
	case "1":
		// set icon name, if Fyne supports in the future
	case "2":
		t.setTitle(arg)
	case "7":
		t.setDirectory(arg)
	case "10":
		if arg == "?" {
			t.replyColor(cmd, theme.ColorNameForeground, terminator)
		}
	case "11":
		if arg == "?" {
			t.replyColor(cmd, theme.ColorNameBackground, terminator)
		}
	default:
		if t.debug {
			log.Println("Unrecognised OSC:", code)
		}
	}
}

// replyColor answers an OSC colour query with the current theme colour
// in the xterm "rgb:rrrr/gggg/bbbb" form.
func (t *Terminal) replyColor(cmd string, name fyne.ThemeColorName, terminator string) {
	var c color.Color = color.Black
	if t.Theme() != nil {
		c = t.Theme().Color(name, fyne.CurrentApp().Settings().ThemeVariant())
	}
	r, g, b, _ := c.RGBA()
	_, _ = fmt.Fprintf(t.in, "\x1b]%s;rgb:%04x/%04x/%04x%s", cmd, r, g, b, terminator)
}

func (t *Terminal) setDirectory(uri string) {
	u, err := storage.ParseURI(uri)
	if err != nil {
		// working around a Fyne bug where file URI does not parse host
		off := 4
		count := 0
		for count < 3 && off < len(uri) {
			off++
			if uri[off] == '/' {
				count++
			}

		}
		_ = os.Chdir(uri[off:])
		return
	}

	// fallback to guessing it's a path
	_ = os.Chdir(u.Path())
}

func (t *Terminal) setTitle(title string) {
	t.config.Title = title
	t.onConfigure()
}
