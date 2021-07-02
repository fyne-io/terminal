package terminal

import (
	"log"
	"os"

	"fyne.io/fyne/v2/storage"
)

func (t *Terminal) handleOSC(code string) {
	if len(code) <= 2 || code[1] != ';' {
		return
	}

	switch code[0] {
	case '0':
		// set icon name, if Fyne supports in the future
		t.setTitle(code[2:])
	case '1':
		// set icon name, if Fyne supports in the future
	case '2':
		t.setTitle(code[2:])
	case '7':
		t.setDirectory(code[2:])
	default:
		if t.debug {
			log.Println("Unrecognised OSC:", code)
		}
	}
}

func (t *Terminal) setDirectory(uri string) {
	u, err := storage.ParseURI(uri)
	if err == nil {
		// working around a Fyne bug where file URI does not parse host
		off := 4
		count := 0
		for count < 3 && off < len(uri) {
			off++
			if uri[off] == '/' {
				count++
			}

		}
		os.Chdir(uri[off:])
		return
	}

	// fallback to guessing it's a path
	os.Chdir(u.Path())
}

func (t *Terminal) setTitle(title string) {
	t.config.Title = title
	t.onConfigure()
}
