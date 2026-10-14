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

func TestTerminal_WrapBeyondEdge(t *testing.T) {
	term := New()
	term.Resize(fyne.NewSize(45, 45))
	assert.Equal(t, uint(5), term.config.Columns)

	term.cursorCol = 8 // as if restored from a wider terminal
	term.handleOutput([]byte("Hello World"))

	for _, row := range term.content.Rows {
		assert.LessOrEqual(t, len(row.Cells), 5)
	}
}
