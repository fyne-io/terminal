package terminal

import (
	"testing"

	"fyne.io/fyne/v2"
)

// The tests in this file show their failures with the race detector (go test -race): the
// terminal parses output on the goroutine that reads it, and everything that touches the widget's
// state runs on Fyne's main goroutine, which useMainLoop stands in for.

func TestRace_SelectionClearedByOutput(t *testing.T) {
	loop := useMainLoop(t)
	term := New()
	loop.do(func() { term.Resize(fyne.NewSize(100, 100)) })

	done := make(chan struct{})
	go func() { // the user selects text while output arrives
		defer close(done)
		for i := 0; i < 100; i++ {
			loop.do(func() {
				term.selStart = &position{Row: 1, Col: 1}
				term.selEnd = &position{Row: 1, Col: 2}
			})
		}
	}()
	for i := 0; i < 100; i++ {
		term.handleOutput(nil) // a read that brings nothing to show still looks at the selection
	}
	<-done
	loop.flush()
}

func TestRace_CharacterSets(t *testing.T) {
	loop := useMainLoop(t)
	term := New()
	loop.do(func() { term.Resize(fyne.NewSize(100, 100)) })

	// shift out and shift in switch the character set on the main goroutine; designations and
	// the text in between are parsed on this one
	for i := 0; i < 50; i++ {
		term.handleOutput([]byte("\x1b)0\x0eqqq\x0fqqq\x1b(B"))
	}
	loop.flush()
}
