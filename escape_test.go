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
	expectedContent := "" // After scrolling and clearing, the visible area should be empty
	for i := 0; i < 46; i++ {
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
	term.config.Columns = 5
	term.config.Rows = 2
	term.handleOutput([]byte("Hello"))
	assert.Equal(t, "Hello", term.content.Text())

	term.moveCursor(0, 2)
	term.handleEscape("2@")
	assert.Equal(t, "He  llo", term.content.Text())
	term.handleEscape("3P")
	assert.Equal(t, "Helo", term.content.Text())
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
			term.handleOutput([]byte(testCase.input))
			actual := term.content.Text()
			if actual != testCase.expected {
				t.Errorf("Expected: %s, Got: %s", testCase.expected, actual)
			}
		})
	}
}
