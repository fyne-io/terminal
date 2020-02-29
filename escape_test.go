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
