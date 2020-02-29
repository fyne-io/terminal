package terminal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClearScreen(t *testing.T) {
	term := NewTerminal()
	term.content.SetText("Hello")
	assert.Equal(t, "Hello", term.content.Text())

	term.handleEscape("2J")
	assert.Equal(t, "", term.content.Text())
}

func TestEraseLine(t *testing.T) {
	term := NewTerminal()
	term.content.SetText("Hello")
	assert.Equal(t, "Hello", term.content.Text())

	term.handleEscape("K")
	assert.Equal(t, "", term.content.Text())
}

func TestCursorMove(t *testing.T) {
	term := NewTerminal()
	term.handleOutput([]byte("Hello"))
	assert.Equal(t, 0, term.cursorRow)
	assert.Equal(t, 5, term.cursorCol)

	term.handleEscape("0;2H")
	assert.Equal(t, 0, term.cursorRow)
	assert.Equal(t, 2, term.cursorCol)
}
