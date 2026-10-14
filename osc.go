package terminal

import (
	"fmt"
	"image/color"
	"log"
	"net/url"
	"strings"

	"fyne.io/fyne/v2"
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

// setDirectory records the directory the shell reports (OSC 7) in the configuration. It does
// not change the working directory of the process, which the terminal shares with the app it is
// embedded in (and with every other terminal in it): output from a remote shell must not move it.
func (t *Terminal) setDirectory(uri string) {
	t.config.PWD = directoryOf(uri)
	t.onConfigure()
}

// directoryOf returns the path of the file URI OSC 7 reports ("file://host/path", the path
// percent-encoded), or the text as it is when it is not one.
func directoryOf(uri string) string {
	if u, err := url.Parse(uri); err == nil && u.Scheme == "file" {
		return u.Path
	}
	return uri
}

func (t *Terminal) setTitle(title string) {
	t.config.Title = title
	t.onConfigure()
}
