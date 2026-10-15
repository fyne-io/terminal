package terminal

import (
	"testing"

	"fyne.io/fyne/v2"
	"github.com/stretchr/testify/assert"
)

func TestAPC(t *testing.T) {
	var APCString string

	RegisterAPCHandler("set apcstring:", func(terminal *Terminal, s string) {
		APCString = s
	})

	testCases := map[string]struct {
		input    []byte
		expected string
	}{
		"starts with APC command": {
			input:    append([]byte("\x1b_set apcstring:Hello"), 0),
			expected: "Hello",
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			term := New()
			term.Resize(fyne.NewSize(50, 50))
			term.handleOutput(testCase.input)

			assert.Equal(t, testCase.expected, APCString)
		})
	}
}
