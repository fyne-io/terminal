package terminal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTerminal_Backspace(t *testing.T) {
	term := NewTerminal()
	term.handleOutput([]byte("Hi"))
	assert.Equal(t, "Hi", term.content.Text())

	term.handleOutput([]byte{asciiBackspace})
	assert.Equal(t, "H", term.content.Text())
}
