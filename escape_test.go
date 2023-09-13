package terminal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClearScreen(t *testing.T) {
	term := New()
	term.config.Columns = 5
	term.config.Rows = 2
	term.handleOutput([]byte("Hello"))
	assert.Equal(t, "Hello", term.content.Text())

	term.handleEscape("2J")
	assert.Equal(t, "", term.content.Text())
}

func TestEraseLine(t *testing.T) {
	term := New()
	term.config.Columns = 5
	term.config.Rows = 2
	term.handleOutput([]byte("Hello"))
	assert.Equal(t, "Hello", term.content.Text())

	term.moveCursor(0, 2)
	term.handleEscape("K")
	assert.Equal(t, "He", term.content.Text())
}

func TestCursorMove(t *testing.T) {
	term := New()
	term.config.Columns = 5
	term.config.Rows = 2
	term.handleOutput([]byte("Hello"))
	assert.Equal(t, 0, term.cursorRow)
	assert.Equal(t, 5, term.cursorCol)

	term.handleEscape("1;4H")
	assert.Equal(t, 0, term.cursorRow)
	assert.Equal(t, 3, term.cursorCol)

	term.handleEscape("2D")
	assert.Equal(t, 0, term.cursorRow)
	assert.Equal(t, 1, term.cursorCol)

	term.handleEscape("2C")
	assert.Equal(t, 0, term.cursorRow)
	assert.Equal(t, 3, term.cursorCol)

	term.handleEscape("1B")
	assert.Equal(t, 1, term.cursorRow)
	assert.Equal(t, 3, term.cursorCol)

	term.handleEscape("1A")
	assert.Equal(t, 0, term.cursorRow)
	assert.Equal(t, 3, term.cursorCol)
}

func TestCursorMove_Overflow(t *testing.T) {
	term := New()
	term.config.Columns = 2
	term.config.Rows = 2
	term.handleEscape("2;2H")
	assert.Equal(t, 1, term.cursorRow)
	assert.Equal(t, 1, term.cursorCol)

	term.handleEscape("2D")
	assert.Equal(t, 1, term.cursorRow)
	assert.Equal(t, 0, term.cursorCol)

	term.handleEscape("5C")
	assert.Equal(t, 1, term.cursorRow)
	assert.Equal(t, 1, term.cursorCol)

	term.handleEscape("5A")
	assert.Equal(t, 0, term.cursorRow)
	assert.Equal(t, 1, term.cursorCol)

	term.handleEscape("4B")
	assert.Equal(t, 1, term.cursorRow)
	assert.Equal(t, 1, term.cursorCol)
}

func TestTrimLeftZeros(t *testing.T) {
	assert.Equal(t, "1", trimLeftZeros(string([]byte{0, 0, '1'})))
}

func TestTerminalEscapeSequences(t *testing.T) {
	testCases := []struct {
		input       string
		expected    string
		description string
	}{
		{
			input:       string([]byte{asciiEscape}) + "(BHello",
			expected:    "Hello",
			description: "Test set G0 to ASCII charset",
		},
		{
			input:       string([]byte{asciiEscape}) + ")BHola",
			expected:    "Hola",
			description: "Test set G1 to ASCII charset",
		},
		{
			input:       string([]byte{asciiEscape}) + "(0oooo",
			expected:    "⎺⎺⎺⎺", // Using decSpecialGraphics map
			description: "Test set G0 to DEC charset",
		},
		{
			input:       string([]byte{asciiEscape, ')', '0', 0x0e}) + "oooo",
			expected:    "⎺⎺⎺⎺",
			description: "Test set G1 to DEC charset and 'SO' to switch to G1",
		},
		{
			input:       string([]byte{asciiEscape, ')', '0', 0x0e}) + "oooo" + string([]byte{0x0f, 'o'}),
			expected:    "⎺⎺⎺⎺o",
			description: "Test set G1 to DEC charset and 'SO' to switch to G1, then 'SI' to G0",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			term := New()
			term.config.Columns = 10
			term.config.Rows = 1
			term.handleOutput([]byte(testCase.input))
			actual := term.content.Text()
			if actual != testCase.expected {
				t.Errorf("Expected: %s, Got: %s", testCase.expected, actual)
			}
		})
	}
}
