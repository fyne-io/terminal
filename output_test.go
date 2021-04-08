package terminal

import (
	"testing"

	"fyne.io/fyne/v2"

	"github.com/stretchr/testify/assert"
)

func TestTerminal_Backspace(t *testing.T) {
	term := New()
	term.Resize(fyne.NewSize(50, 50))
	term.handleOutput([]byte("Hi"))
	assert.Equal(t, "Hi", term.content.Text())

	term.handleOutput([]byte{asciiBackspace})
	term.handleOutput([]byte("ello"))

	assert.Equal(t, "Hello", term.content.Text())
}
