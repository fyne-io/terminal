package terminal

import (
	"strconv"
	"strings"
	"testing"

	"fyne.io/fyne/v2"
	"github.com/stretchr/testify/assert"
)

func TestClearScreen(t *testing.T) {
	term := New()
	term.config.Columns = 5
	term.config.Rows = 2
	term.Refresh() // ensure visuals set up

	term.handleOutput([]byte("Hello"))
	assert.Equal(t, "Hello", term.content.Text())

	term.handleEscape("2J")
	assert.Equal(t, "", term.content.Text())
}

// test clearing the screen by using "scrollback"
// this is a method tmux uses to "clear the screen"
func TestScrollBack_Tmux(t *testing.T) {
	// Step 1: Setup a new terminal instance
	term := New()
	term.debug = true
	term.config.Columns = 80 // 80 columns (standard terminal width)
	term.config.Rows = 5     // Doesn't matter
	term.Refresh()           // ensure visuals set up

	// Step 2: Populate the entire screen with lines using cursor movement
	for i := 1; i <= 40; i++ {
		lineText := "Line " + strconv.Itoa(i)
		// Move the cursor to the beginning of each line using the escape sequence \x1b[{row};{col}H
		escapeMoveCursor := "\x1b[" + strconv.Itoa(i) + ";1H"
		term.handleOutput([]byte(escapeMoveCursor + lineText))
	}

	// Step 3: Set up the scroll region and scroll content away
	term.handleOutput([]byte("\x1b[1;47r")) // Set scroll region from lines 1 to 47
	term.handleOutput([]byte("\x1b[2;47r")) // Set scroll region again (redundant in most cases)
	term.handleOutput([]byte("\x1b[46S"))   // Scroll up by 46 lines (this should move almost all content out of view)

	// Step 4: Additional escape sequences to clear the screen
	term.handleOutput([]byte("\x1b[1;1H"))  // Move cursor to the top-left corner
	term.handleOutput([]byte("\x1b[K"))     // Clear the current line
	term.handleOutput([]byte("\x1b[1;48r")) // Restore scroll region to the full screen
	term.handleOutput([]byte("\x1b[1;1H"))  // Move cursor to top-left again
	term.handleOutput([]byte("\x1b(B"))     // Reset character set
	term.handleOutput([]byte("\x1b[m"))     // Reset all attributes

	// Step 5: Check the final content of the terminal
	// The scroll margins are clamped to the screen, so the grid holds its own rows, empty, rather
	// than the 46 an unclamped region of 1;47 grew it to.
	expectedContent := "" // After scrolling and clearing, the visible area should be empty
	for i := 0; i < int(term.config.Rows)-1; i++ {
		expectedContent += "\n" // Each row should be an empty line
	}

	assert.Equal(t, 0, term.cursorRow)
	assert.Equal(t, 0, term.cursorCol)

	assert.Equal(t, expectedContent, term.content.Text())
}

func TestScrollBack_With_Zero_Back_Buffer(t *testing.T) {
	// Define the test cases using a map
	tests := map[string]struct {
		linesToAdd        int
		scrollLines       int
		expectedOutput    string
		expectedCursorRow int
		expectedCursorCol int
	}{
		"when 5 lines added and scrolled up 4 lines, should show line 5": {
			linesToAdd:        5,
			scrollLines:       4,
			expectedOutput:    "Line 5",
			expectedCursorRow: 0, // Adjust as needed
			expectedCursorCol: 6, // Assuming cursor is at end of visible text
		},
		// Add more test cases here as needed
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// Setup: Create a new terminal instance with a zero back buffer
			term := New()
			term.debug = true
			term.config.Columns = 80 // 80 columns (standard terminal width)
			term.config.Rows = 5     // Doesn't matter
			term.Refresh()           // ensure visuals set up

			// Step 1: Populate the entire screen with lines using cursor movement
			for i := 1; i <= tt.linesToAdd; i++ {
				lineText := "Line " + strconv.Itoa(i)
				// Move the cursor to the beginning of each line using the escape sequence \x1b[{row};{col}H
				escapeMoveCursor := "\x1b[" + strconv.Itoa(i) + ";1H" // Move cursor to row i, column 1
				term.handleOutput([]byte(escapeMoveCursor + lineText))
			}
			term.handleOutput([]byte("\x1b[1;" + strconv.Itoa(tt.linesToAdd) + "r")) // Set scroll region from lines 1 to linesToAdd

			term.handleOutput([]byte("\x1b[" + strconv.Itoa(tt.scrollLines) + "S")) // Scroll up

			// Step 3: Get the current output after scrolling
			currentOutput := strings.TrimRight(term.Text(), "\n")

			// Step 4: Assert that the output matches
			assert.Equal(t, tt.expectedOutput, currentOutput)

			// Step 5: Check the final content of the terminal
			assert.Equal(t, tt.expectedCursorRow, term.cursorRow)
			assert.Equal(t, tt.expectedCursorCol, term.cursorCol)
		})
	}
}

func TestInsertDeleteChars(t *testing.T) {
	term := New()
	term.config.Columns = 10
	term.config.Rows = 2
	term.Refresh() // ensure visuals set up

	term.handleOutput([]byte("Hello"))
	assert.Equal(t, "Hello", term.content.Text())

	term.moveCursor(0, 2)
	term.handleEscape("2@")
	assert.Equal(t, "He  llo", term.content.Text())
	term.handleEscape("3P")
	assert.Equal(t, "Helo", term.content.Text())
}

func TestInsertChars_Overflow(t *testing.T) {
	term := New()
	term.config.Columns = 5
	term.config.Rows = 2
	term.Refresh() // ensure visuals set up

	term.handleOutput([]byte("Hello"))
	term.moveCursor(0, 2)
	term.handleEscape("2@")
	assert.Equal(t, "He  l", term.content.Text())

	term.handleEscape("60000@")
	assert.Equal(t, "He   ", term.content.Text())
	assert.Equal(t, 5, len(term.content.Rows[0].Cells))
}

func TestEraseChars_Overflow(t *testing.T) {
	term := New()
	term.config.Columns = 5
	term.config.Rows = 2
	term.Refresh() // ensure visuals set up

	term.handleOutput([]byte("Hello"))
	term.moveCursor(0, 2)
	term.handleEscape("60000X")
	assert.Equal(t, "He   ", term.content.Text())
	assert.Equal(t, 5, len(term.content.Rows[0].Cells))
}

func TestEraseLine(t *testing.T) {
	term := New()
	term.config.Columns = 5
	term.config.Rows = 2
	term.Refresh() // ensure visuals set up

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
	term.Refresh() // ensure visuals set up

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
	term.Refresh() // ensure visuals set up

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

func TestHandleOutput_NewLineMode(t *testing.T) {
	tests := []struct {
		name                    string
		input                   string // Escape codes to feed into handleOutput
		expectedCursorRow       int    // Expected cursorRow after processing escape codes
		expectedCursorCol       int    // Expected cursorCol after processing escape codes
		expectedNewLineMode     bool   // Expected value of newLineMode after processing escape codes
		expectedContentText     string // Expected content text after processing escape codes
		expectedContentRowCount int    // Expected number of rows in content after processing escape codes
	}{
		{
			name:                    "single line",
			input:                   "hello",
			expectedCursorRow:       0,
			expectedCursorCol:       5,
			expectedNewLineMode:     false,
			expectedContentText:     "hello",
			expectedContentRowCount: 1,
		},
		{
			name:                    "Default - carriage return new line",
			input:                   "hello\r\nworld",
			expectedCursorRow:       1,
			expectedCursorCol:       5,
			expectedNewLineMode:     false,
			expectedContentText:     "hello\nworld",
			expectedContentRowCount: 2,
		},
		{
			name:                    "Default - new line",
			input:                   "hello\nworld",
			expectedCursorRow:       1,
			expectedCursorCol:       10,
			expectedNewLineMode:     false,
			expectedContentText:     "hello\n     world",
			expectedContentRowCount: 2,
		},
		{
			name:                    "Enable New Line Mode",
			input:                   "\x1b[?20hhello\nworld",
			expectedCursorRow:       1,
			expectedCursorCol:       5,
			expectedNewLineMode:     true,
			expectedContentText:     "hello\nworld",
			expectedContentRowCount: 2,
		},
		{
			name:                    "Enable then disable New Line Mode",
			input:                   "\x1b[?20h\x1b[?20lhello\nworld",
			expectedCursorRow:       1,
			expectedCursorCol:       10,
			expectedNewLineMode:     false,
			expectedContentText:     "hello\n     world",
			expectedContentRowCount: 2,
		},
		{
			name:                    "Enable new line mode - lf vt ff",
			input:                   "\x1b[?20hhello\n\v\fworld",
			expectedCursorRow:       3,
			expectedCursorCol:       5,
			expectedNewLineMode:     true,
			expectedContentText:     "hello\n\n\nworld",
			expectedContentRowCount: 4,
		},
		{
			name:                    "Default new line mode - lf vt ff",
			input:                   "hello\n\v\fworld",
			expectedCursorRow:       3,
			expectedCursorCol:       10,
			expectedNewLineMode:     false,
			expectedContentText:     "hello\n\n\n     world",
			expectedContentRowCount: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			term := New()
			term.Resize(fyne.NewSize(500, 500))

			term.handleOutput([]byte(tt.input))

			assert.Equal(t, tt.expectedCursorRow, term.cursorRow)
			assert.Equal(t, tt.expectedCursorCol, term.cursorCol)
			assert.Equal(t, tt.expectedNewLineMode, term.newLineMode)
			assert.Equal(t, tt.expectedContentText, term.content.Text())
			assert.Equal(t, tt.expectedContentRowCount, len(term.content.Rows))
		})
	}
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
			term.Refresh() // ensure visuals set up

			term.handleOutput([]byte(testCase.input))
			actual := term.content.Text()
			if actual != testCase.expected {
				t.Errorf("Expected: %s, Got: %s", testCase.expected, actual)
			}
		})
	}
}

func TestHandleEscape_NonASCIIFinal(t *testing.T) {
	term := New()
	term.config.Columns = 10
	term.config.Rows = 2
	term.Refresh() // ensure visuals set up

	term.handleOutput([]byte("\x1b[é\x1b[1;2€ok"))
	assert.Equal(t, "ok", term.content.Text())
}

func TestDeleteChars(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected string
	}{
		"within the row":   {input: "\x1b[1;3H\x1b[2P", expected: "abef\n2"},
		"past the row":     {input: "\x1b[1;3H\x1b[9P", expected: "ab\n2"},
		"cursor past row":  {input: "\x1b[1;9H\x1b[P", expected: "abcdef\n2"},
		"largest":          {input: "\x1b[1;3H\x1b[9223372036854775807P", expected: "ab\n2"},
		"largest, at last": {input: "\x1b[1;6H\x1b[9223372036854775807P", expected: "abcde\n2"},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			term := New()
			term.config.Columns = 10
			term.config.Rows = 2
			term.scrollBottom = 1
			term.Refresh() // ensure visuals set up
			term.handleOutput([]byte("abcdef\r\n2"))
			term.handleOutput([]byte(test.input))
			assert.Equal(t, test.expected, term.content.Text())
		})
	}
}

func TestInsertChars_PastTheRow(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected string
	}{
		"fresh screen":     {input: "\x1b[2;3H\x1b[2@x", expected: "\n  x "},
		"past a short row": {input: "ab\x1b[1;5H\x1b[@x", expected: "ab  x"},
		"within the row":   {input: "abc\x1b[1;2H\x1b[@x", expected: "axbc"},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			term := New()
			term.config.Columns = 10
			term.config.Rows = 2
			term.scrollBottom = 1
			term.Refresh() // ensure visuals set up
			term.handleOutput([]byte(test.input))
			assert.Equal(t, test.expected, term.content.Text())
		})
	}
}

func TestSetScrollArea_Invalid(t *testing.T) {
	tests := map[string]struct {
		input       string
		top, bottom int
	}{
		"top below bottom": {input: "\x1b[3;1r\x1bD\x1bM", top: 0, bottom: 4},
		"past the screen":  {input: "\x1b[1;47r\x1bM", top: 0, bottom: 4},
		"bottom past":      {input: "\x1b[4;9r\x1b[4;1H\x1bM", top: 3, bottom: 4},
		"bottom of zero":   {input: "\x1b[1;0r\x1b[5;1H\n\n", top: 0, bottom: 4},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			term := New()
			term.config.Columns = 5
			term.config.Rows = 5
			term.scrollBottom = 4
			term.Refresh() // ensure visuals set up
			term.handleOutput([]byte(test.input))
			assert.Equal(t, test.top, term.scrollTop)
			assert.Equal(t, test.bottom, term.scrollBottom)
		})
	}
}

func TestSetMode_WithoutPrivateMarker(t *testing.T) {
	term := New()
	term.config.Columns = 10
	term.config.Rows = 2
	term.Refresh() // ensure visuals set up

	term.handleOutput([]byte("\x1b[h\x1b[l\x1b[;h\x1b[4h\x1b[4l"))
	assert.False(t, term.newLineMode)

	term.handleOutput([]byte("\x1b[20h"))
	assert.True(t, term.newLineMode)
	term.handleOutput([]byte("\x1b[20l"))
	assert.False(t, term.newLineMode)

	term.handleOutput([]byte("\x1b[125l")) // not the private mode 25
	assert.False(t, term.cursorHidden)
	term.handleOutput([]byte("\x1b[?25l"))
	assert.True(t, term.cursorHidden)
}

func TestDeviceAttributes_Debug(t *testing.T) {
	term := New()
	term.debug = true
	term.config.Columns = 10
	term.config.Rows = 2
	term.Refresh() // ensure visuals set up

	term.handleOutput([]byte("\x1b[>c\x1b[=c\x1b[>0c\x1b[?1c"))
	assert.Equal(t, "", term.content.Text())
}

func TestRepeatChar_LargeCount(t *testing.T) {
	newTerm := func(cols, rows uint) *Terminal {
		term := New()
		term.config.Columns = cols
		term.config.Rows = rows
		term.scrollBottom = int(rows) - 1
		term.Refresh() // ensure visuals set up
		return term
	}
	sizes := []struct{ cols, rows uint }{{5, 3}, {3, 1}, {1, 1}, {0, 2}, {4, 0}}
	prefixes := map[string]string{
		"top left":        "",
		"after text":      "ab\r\ncd",
		"scroll region":   "1\r\n2\r\n3\x1b[2;3r\x1b[2;2H",
		"below region":    "\x1b[1;2r\x1b[3;1H",
		"without wrap":    "\x1b[?7l",
		"line drawing":    "\x1b(0",
		"past the margin": "\x1b[1;5Hxy",
	}

	// repeating a character must look the same as writing it that many times
	for _, size := range sizes {
		for name, prefix := range prefixes {
			for _, count := range []int{1, 14, 15, 16, 29, 30, 31, 47, 100, 1001} {
				written := newTerm(size.cols, size.rows)
				written.handleOutput([]byte(prefix + strings.Repeat("q", count+1)))
				repeated := newTerm(size.cols, size.rows)
				repeated.handleOutput([]byte(prefix + "q\x1b[" + strconv.Itoa(count) + "b"))

				where := name + " " + strconv.Itoa(int(size.cols)) + "x" + strconv.Itoa(int(size.rows)) + " " + strconv.Itoa(count)
				assert.Equal(t, written.content.Text(), repeated.content.Text(), where)
				assert.Equal(t, written.cursorRow, repeated.cursorRow, where)
				assert.Equal(t, written.cursorCol, repeated.cursorCol, where)
			}
		}
	}

	term := newTerm(80, 24)
	term.handleOutput([]byte("x\x1b[9223372036854775807b"))
	assert.Equal(t, 23, term.cursorRow)
}
