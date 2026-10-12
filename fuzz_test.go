package terminal

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime/debug"
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
)

// fuzzTimeout is how long one input may take. An input that has not finished by then counts as a
// hang: output that makes the parser loop for ever freezes the app it runs in. Inputs take
// milliseconds; the bound is generous because a loaded machine stalls fuzzing workers for seconds.
const fuzzTimeout = 30 * time.Second

// fuzzSeeds are sequences that once crashed or froze the terminal, and some that real programs send.
var fuzzSeeds = []string{
	"\x1b[é",                       // CSI with a non-ASCII final character
	"\x1b[h", "\x1b[l", "\x1b[20h", // set and reset mode without "?"
	"\x1b[1;10H\x1b[P", // delete characters past the end of the row
	"\x1bM", "\x1bD",   // reverse index and index before any row exists
	"\x1b[9999Gx\x1b[2K\x1b[P",                               // delete characters after erasing the line
	"\x1b[999Gx\t",                                           // tab at the right margin
	"\x1b[5;1H\x1b[@",                                        // insert characters on a fresh screen
	"a\r\n\x1b[1;0r\x1bD",                                    // scroll region with a bottom of 0
	"x\x1b[9223372036854775807b",                             // repeat a character many times
	"\x1b[9223372036854775807M", "\x1b[9223372036854775807L", // delete and insert many lines
	"\x1b[9223372036854775807S", "\x1b[2;5r\x1b[9223372036854775807T", // scroll many lines
	"abcdef\x1b[1;3H\x1b[9223372036854775807P", // delete many characters
	"\x1b[>c", "\x1b[=c", "\x1b[c", // device attributes
	"\x1b[5ihidden\x1b[4i",                                    // media copy
	"\x1b]0;title\x07\x1b]7;file:///tmp\x1b\\",                // OSC
	"\x1bP+q544e\x1b\\", "\x1b_apc\x00", "\x1b_G;AAAA\x1b\\x", // DCS and APC
	"\x1b(0lqqk\x1b(B", "\x1b)0\x0eqq\x0f", // line drawing
	"\x1b[?1049h\x1b[2J\x1b[H\x1b[?1049l",                      // alternate screen
	"\x1b[?47h\x1b[3;10r\x1b[10;1H\n\n\x1b[r\x1b[?47l",         // alternate screen and scroll region
	"\x1b[1;5;31;42mred\x1b[38;5;208m\x1b[48;2;1;2;3mx\x1b[0m", // colours, blinking
	"\x1b7\x1b[10;10H\x1b8\x1b[s\x1b[u",                        // save and restore the cursor
	// top
	"\x1b[?1h\x1b=\x1b[?25l\x1b[H\x1b[2J\x1b(B\x1b[mtop - 10:00:01 up 3 days, load average: 0.10\x1b[K\r\n" +
		"\x1b[7m  PID USER      %CPU COMMAND             \x1b[m\x1b[K\r\n" +
		"\x1b[1m    1 root       0.3 init\x1b[m\x1b[K\r\n\x1b[J\x1b[3;1H\x1b[?25h",
	// tmux
	"\x1b[?1049h\x1b[22;0;0t\x1b[?1h\x1b=\x1b[H\x1b[2J\x1b[?12l\x1b[?25h\x1b[?1000l\x1b[?2004h" +
		"\x1b[1;23r\x1b[23;1H\r\n\x1b[24;1H\x1b[30m\x1b[42m[0] 0:sh*\x1b[K\x1b[39m\x1b[49m\x1b[1;1H" +
		"\x1b[2 q\x1b]52;c;aGVsbG8=\x07\x1b]112\x07\x1b[3J\x1b[2;23r\x1b[2;1H\x1b[M\x1b[r",
}

// FuzzHandleOutput feeds arbitrary output to the terminal at arbitrary sizes, optionally resizing
// it part way through, and fails on a panic or on output that does not finish within fuzzTimeout.
// The seeds are fuzzSeeds and the output of real programs in test_data/sessions.
func FuzzHandleOutput(f *testing.F) {
	for _, seed := range fuzzSeeds {
		f.Add([]byte(seed), uint8(80), uint8(24), uint16(0), uint8(0), uint8(0), false)
		f.Add([]byte(seed), uint8(2), uint8(1), uint16(1), uint8(1), uint8(3), true)
	}
	sessions, _ := filepath.Glob(filepath.Join("test_data", "sessions", "*.vt"))
	if len(sessions) == 0 {
		f.Fatal("no sessions in test_data/sessions")
	}
	for _, name := range sessions {
		data, err := os.ReadFile(name)
		if err != nil {
			f.Fatal(err)
		}
		f.Add(data, uint8(80), uint8(24), uint16(len(data)/2), uint8(40), uint8(10), false)
	}

	// The output is parsed on this goroutine, as on the goroutine reading a connection, and the
	// widget changes on the main loop, as in an app.
	loop := useMainLoop(f)
	var ringing []ringingTerm
	f.Fuzz(func(t *testing.T, data []byte, cols, rows uint8, split uint16, cols2, rows2 uint8, debugOn bool) {
		if debugOn {
			log.SetOutput(io.Discard)
			defer log.SetOutput(os.Stderr)
		}
		term := New()
		loop.do(func() {
			term.SetDebug(debugOn)
			term.Refresh() // ensure visuals set up
			resizeTo(term, cols, rows)
		})

		done := make(chan string, 1)
		go func() {
			defer func() {
				if r := recover(); r != nil {
					done <- fmt.Sprintf("panic: %v\n%s", r, debug.Stack())
				}
			}()
			rest := data
			if int(split) > 0 && int(split) < len(data) {
				term.handleOutput(data[:split])
				loop.flush()
				loop.do(func() {
					releaseRenderers(t, term) // rows a smaller size puts aside would be missed
					resizeTo(term, cols2, rows2)
				})
				rest = data[split:]
			}
			term.handleOutput(rest)
			loop.do(func() {
				term.Refresh()
				_ = term.Text()
			})
			done <- ""
		}()

		where := fmt.Sprintf("%dx%d (then %dx%d at %d), %q", cols, rows, cols2, rows2, split, data)
		select {
		case msg := <-done:
			if msg != "" {
				t.Fatalf("%s: %s", where, msg)
			}
		case <-time.After(fuzzTimeout):
			t.Fatalf("%s: not done within %v", where, fuzzTimeout)
		}
		if msg := loop.panicked(); msg != "" {
			t.Fatalf("%s: %s", where, msg)
		}

		// Refreshes queue more work on the main loop: it has to be done before the renderers are
		// collected, or it makes them again after the test has dropped them.
		var rang bool
		loop.flush()
		loop.do(func() {
			releaseRenderers(t, term)
			rang = term.bell
			// A blinking cell keeps a goroutine refreshing the grid until none is left. Emptied
			// before the renderers are collected, the grid would hide its rows from them.
			term.content.Rows = nil
			term.content.Refresh()
		})
		loop.flush()
		// A bell clears itself 300 ms later and redraws the terminal, which makes its renderers
		// again; those are collected once that has happened.
		if rang {
			ringing = append(ringing, ringingTerm{term, time.Now()})
		}
		for len(ringing) > 0 && time.Since(ringing[0].since) > 400*time.Millisecond {
			old := ringing[0].term
			ringing = ringing[1:]
			loop.do(func() { releaseRenderers(t, old) })
			loop.flush()
		}
	})
}

// ringingTerm is a fuzzed terminal whose bell rang.
type ringingTerm struct {
	term  *Terminal
	since time.Time
}

// resizeTo resizes the terminal to the given number of columns and rows.
func resizeTo(term *Terminal, cols, rows uint8) {
	cell := term.guessCellSize()
	term.Resize(fyne.NewSize((float32(cols)+0.5)*cell.Width, (float32(rows)+0.5)*cell.Height))
}

// releaseRenderers makes Fyne drop the cached renderers of o and of everything they show when the
// test ends. Fyne's test driver never expires them, and each keeps its widget: without this a fuzz
// worker keeps every input's terminal (about 650 KB each) and slows down until inputs time out.
func releaseRenderers(t *testing.T, o fyne.CanvasObject) {
	switch o := o.(type) {
	case fyne.Widget:
		for _, child := range test.TempWidgetRenderer(t, o).Objects() {
			releaseRenderers(t, child)
		}
	case *fyne.Container:
		for _, child := range o.Objects {
			releaseRenderers(t, child)
		}
	}
}
