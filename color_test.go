package terminal

import (
	"fmt"
	"image/color"
	"reflect"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	widget2 "github.com/fyne-io/terminal/internal/widget"
	"github.com/stretchr/testify/assert"
)

func esc(s string) string {
	return fmt.Sprintf("%s%s", string(byte(asciiEscape)), s)
}

func testColor(t *testing.T, tests map[string]struct {
	inputSeq      string
	expectedFg    color.Color
	expectedBg    color.Color
	expectedStyle fyne.TextStyle
}) {
	// Iterate through the test cases
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			terminal := New()
			terminal.handleOutput([]byte(test.inputSeq))

			// Verify the actual results match the expected results
			if !reflect.DeepEqual(terminal.currentFG, test.expectedFg) {
				t.Errorf("Foreground color mismatch. Got %v, expected %v", terminal.currentFG, test.expectedFg)
			}

			if !reflect.DeepEqual(terminal.currentBG, test.expectedBg) {
				t.Errorf("Background color mismatch. Got %v, expected %v", terminal.currentBG, test.expectedBg)
			}
			if terminal.bold != test.expectedStyle.Bold {
				t.Errorf("Bold flag mismatch. Got %v, expected %v", terminal.bold, test.expectedStyle.Bold)
			}
			if terminal.italic != test.expectedStyle.Italic {
				t.Errorf("Italic flag mismatch. Got %v, expected %v", terminal.italic, test.expectedStyle.Italic)
			}
			if terminal.underline != test.expectedStyle.Underline {
				t.Errorf("Underline flag mismatch. Got %v, expected %v", terminal.underline, test.expectedStyle.Underline)
			}
			if terminal.strikethrough != test.expectedStyle.Strikethrough {
				t.Errorf("Strikethrough flag mismatch. Got %v, expected %v", terminal.strikethrough, test.expectedStyle.Strikethrough)
			}
		})
	}
}

func TestHandleOutput_Text(t *testing.T) {
	tests := map[string]struct {
		inputSeq      string
		expectedStyle fyne.TextStyle
	}{
		"bold": {
			inputSeq:      esc("[1m"),
			expectedStyle: fyne.TextStyle{Bold: true},
		},
		"italic": {
			inputSeq:      esc("[3m"),
			expectedStyle: fyne.TextStyle{Italic: true},
		},
		"bold and italic": {
			inputSeq:      esc("[1m") + esc("[3m"),
			expectedStyle: fyne.TextStyle{Bold: true, Italic: true},
		},
		"bold then disable bold": {
			inputSeq: esc("[1m") + esc("[22m"),
		},
		"italic then disable italic": {
			inputSeq: esc("[3m") + esc("[23m"),
		},
		"underline then disable underline": {
			inputSeq: esc("[4m") + esc("[24m"),
		},
		"all styles then reset": {
			inputSeq: esc("[1m") + esc("[3m") + esc("[4m") + esc("[9m") + esc("[0m"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			terminal := New()
			terminal.handleOutput([]byte(test.inputSeq))

			assert.Equal(t, test.expectedStyle.Bold, terminal.bold, "Bold mismatch")
			assert.Equal(t, test.expectedStyle.Italic, terminal.italic, "Italic mismatch")
			assert.Equal(t, test.expectedStyle.Underline, terminal.underline, "Underline mismatch")
			assert.Equal(t, test.expectedStyle.Strikethrough, terminal.strikethrough, "Strikethrough mismatch")
		})
	}
}

func TestHandleOutput_Normal_Text(t *testing.T) {
	tests := map[string]struct {
		inputSeq      string
		expectedFg    color.Color
		expectedBg    color.Color
		expectedStyle fyne.TextStyle
	}{
		"reverse video": {
			inputSeq:   esc("[7m"),
			expectedFg: color.NRGBA{R: 34, G: 34, B: 34, A: 255},
			expectedBg: color.NRGBA{R: 255, G: 255, B: 255, A: 255},
		},
		"reverse video and bold": {
			inputSeq:      esc("[7m") + esc("[1m"),
			expectedFg:    color.NRGBA{R: 34, G: 34, B: 34, A: 255},
			expectedBg:    color.NRGBA{R: 255, G: 255, B: 255, A: 255},
			expectedStyle: fyne.TextStyle{Bold: true},
		},
		"reverse video and bold then reset": {
			inputSeq:   esc("[7m") + esc("[1m") + esc("[m"),
			expectedFg: nil,
			expectedBg: nil,
		},
	}

	testColor(t, tests)
}

func TestHandleOutput_ANSI_Colors(t *testing.T) {
	tests := map[string]struct {
		inputSeq      string
		expectedFg    color.Color
		expectedBg    color.Color
		expectedStyle fyne.TextStyle
	}{
		"[30m": {
			inputSeq:   esc("[30m"),
			expectedFg: color.Black,
			expectedBg: nil,
		},
		"[31m": {
			inputSeq:   esc("[31m"),
			expectedFg: &color.RGBA{170, 0, 0, 255},
			expectedBg: nil,
		},
		"[32m": {
			inputSeq:   esc("[32m"),
			expectedFg: &color.RGBA{0, 170, 0, 255},
			expectedBg: nil,
		},
		"[33m": {
			inputSeq:   esc("[33m"),
			expectedFg: &color.RGBA{170, 170, 0, 255},
			expectedBg: nil,
		},
		"[34m": {
			inputSeq:   esc("[34m"),
			expectedFg: &color.RGBA{0, 0, 170, 255},
			expectedBg: nil,
		},
		"[35m": {
			inputSeq:   esc("[35m"),
			expectedFg: &color.RGBA{170, 0, 170, 255},
			expectedBg: nil,
		},
		"[36m": {
			inputSeq:   esc("[36m"),
			expectedFg: &color.RGBA{0, 255, 255, 255},
			expectedBg: nil,
		},
		"[37m": {
			inputSeq:   esc("[37m"),
			expectedFg: &color.RGBA{170, 170, 170, 255},
			expectedBg: nil,
		},
		"bold": {
			inputSeq:      esc("[1m") + esc("[37m"),
			expectedFg:    &color.RGBA{170, 170, 170, 255},
			expectedBg:    nil,
			expectedStyle: fyne.TextStyle{Bold: true},
		},
		"reverse video": {
			inputSeq:   esc("[7m") + esc("[37m"),
			expectedFg: &color.RGBA{170, 170, 170, 255},
			expectedBg: color.NRGBA{255, 255, 255, 255},
		},
		"underline": {
			inputSeq:      esc("[4m"),
			expectedStyle: fyne.TextStyle{Underline: true},
		},
		"bold and underline": {
			inputSeq:      esc("[1m") + esc("[4m"),
			expectedStyle: fyne.TextStyle{Bold: true, Underline: true},
		},
		"strikethrough": {
			inputSeq:      esc("[9m"),
			expectedStyle: fyne.TextStyle{Strikethrough: true},
		},
		"strikethrough and underline": {
			inputSeq:      esc("[9m") + esc("[4m"),
			expectedStyle: fyne.TextStyle{Strikethrough: true, Underline: true},
		},
	}

	testColor(t, tests)
}

func TestHandleOutput_xterm_bright(t *testing.T) {
	tests := map[string]struct {
		inputSeq      string
		expectedFg    color.Color
		expectedBg    color.Color
		expectedStyle fyne.TextStyle
	}{
		"[90m": {
			inputSeq:   esc("[90m"),
			expectedFg: &color.RGBA{85, 85, 85, 255},
			expectedBg: nil,
		},
		"[100m": {
			inputSeq:   esc("[100m"),
			expectedFg: nil,
			expectedBg: &color.RGBA{85, 85, 85, 255},
		},
		"[91m": {
			inputSeq:   esc("[91m"),
			expectedFg: &color.RGBA{255, 85, 85, 255},
			expectedBg: nil,
		},
		"[101m": {
			inputSeq:   esc("[101m"),
			expectedFg: nil,
			expectedBg: &color.RGBA{255, 85, 85, 255},
		},
		"[92m": {
			inputSeq:   esc("[92m"),
			expectedFg: &color.RGBA{85, 255, 85, 255},
			expectedBg: nil,
		},
		"[102m": {
			inputSeq:   esc("[102m"),
			expectedFg: nil,
			expectedBg: &color.RGBA{85, 255, 85, 255},
		},
		"[93m": {
			inputSeq:   esc("[93m"),
			expectedFg: &color.RGBA{255, 255, 85, 255},
			expectedBg: nil,
		},
		"[103m": {
			inputSeq:   esc("[103m"),
			expectedFg: nil,
			expectedBg: &color.RGBA{255, 255, 85, 255},
		},
		"[94m": {
			inputSeq:   esc("[94m"),
			expectedFg: &color.RGBA{85, 85, 255, 255},
			expectedBg: nil,
		},
		"[104m": {
			inputSeq:   esc("[104m"),
			expectedFg: nil,
			expectedBg: &color.RGBA{85, 85, 255, 255},
		},
		"[95m": {
			inputSeq:   esc("[95m"),
			expectedFg: &color.RGBA{255, 85, 255, 255},
			expectedBg: nil,
		},
		"[105m": {
			inputSeq:   esc("[105m"),
			expectedFg: nil,
			expectedBg: &color.RGBA{255, 85, 255, 255},
		},
		"[96m": {
			inputSeq:   esc("[96m"),
			expectedFg: &color.RGBA{85, 255, 255, 255},
			expectedBg: nil,
		},

		"[106m": {
			inputSeq:   esc("[106m"),
			expectedFg: nil,
			expectedBg: &color.RGBA{85, 255, 255, 255},
		},
		"[97m": {
			inputSeq:   esc("[97m"),
			expectedFg: &color.RGBA{255, 255, 255, 255},
			expectedBg: nil,
		},
		"[107m": {
			inputSeq:   esc("[107m"),
			expectedFg: nil,
			expectedBg: &color.RGBA{255, 255, 255, 255},
		},
	}

	testColor(t, tests)
}

func TestHandleOutput_xterm_256_1(t *testing.T) {
	tests := map[string]struct {
		inputSeq      string
		expectedFg    color.Color
		expectedBg    color.Color
		expectedStyle fyne.TextStyle
	}{
		"[48;5;16m": {
			inputSeq:   esc("[48;5;16m"),
			expectedFg: nil,
			expectedBg: &color.RGBA{0, 0, 0, 255},
		},
		"[48;5;232m": {
			inputSeq:   esc("[48;5;232m"),
			expectedFg: nil,
			expectedBg: &color.Gray{0},
		},
		"[48;5;233m": {
			inputSeq:   esc("[48;5;233m"),
			expectedFg: nil,
			expectedBg: &color.Gray{10},
		},
		"[48;5;234m": {
			inputSeq:   esc("[48;5;234m"),
			expectedFg: nil,
			expectedBg: &color.Gray{20},
		},
		"[48;5;235m": {
			inputSeq:   esc("[48;5;235m"),
			expectedFg: nil,
			expectedBg: &color.Gray{30},
		},
		"[48;5;236m": {
			inputSeq:   esc("[48;5;236m"),
			expectedFg: nil,
			expectedBg: &color.Gray{40},
		},
		"[48;5;237m": {
			inputSeq:   esc("[48;5;237m"),
			expectedFg: nil,
			expectedBg: &color.Gray{50},
		},
		"[48;5;238m": {
			inputSeq:   esc("[48;5;238m"),
			expectedFg: nil,
			expectedBg: &color.Gray{60},
		},
		"[48;5;239m": {
			inputSeq:   esc("[48;5;239m"),
			expectedFg: nil,
			expectedBg: &color.Gray{70},
		},
		"[48;5;240m": {
			inputSeq:   esc("[48;5;240m"),
			expectedFg: nil,
			expectedBg: &color.Gray{80},
		},
		"[48;5;241m": {
			inputSeq:   esc("[48;5;241m"),
			expectedFg: nil,
			expectedBg: &color.Gray{90},
		},
		"[48;5;242m": {
			inputSeq:   esc("[48;5;242m"),
			expectedFg: nil,
			expectedBg: &color.Gray{100},
		},
		"[48;5;243m": {
			inputSeq:   esc("[48;5;243m"),
			expectedFg: nil,
			expectedBg: &color.Gray{110},
		},
		"[48;5;244m": {
			inputSeq:   esc("[48;5;244m"),
			expectedFg: nil,
			expectedBg: &color.Gray{120},
		},
		"[48;5;245m": {
			inputSeq:   esc("[48;5;245m"),
			expectedFg: nil,
			expectedBg: &color.Gray{130},
		},
		"[48;5;246m": {
			inputSeq:   esc("[48;5;246m"),
			expectedFg: nil,
			expectedBg: &color.Gray{140},
		},
		"[48;5;247m": {
			inputSeq:   esc("[48;5;247m"),
			expectedFg: nil,
			expectedBg: &color.Gray{150},
		},
		"[48;5;248m": {
			inputSeq:   esc("[48;5;248m"),
			expectedFg: nil,
			expectedBg: &color.Gray{160},
		},
		"[48;5;249m": {
			inputSeq:   esc("[48;5;249m"),
			expectedFg: nil,
			expectedBg: &color.Gray{170},
		},
		"[48;5;250m": {
			inputSeq:   esc("[48;5;250m"),
			expectedFg: nil,
			expectedBg: &color.Gray{180},
		},
		"[48;5;251m": {
			inputSeq:   esc("[48;5;251m"),
			expectedFg: nil,
			expectedBg: &color.Gray{190},
		},
		"[48;5;252m": {
			inputSeq:   esc("[48;5;252m"),
			expectedFg: nil,
			expectedBg: &color.Gray{200},
		},
		"[48;5;253m": {
			inputSeq:   esc("[48;5;253m"),
			expectedFg: nil,
			expectedBg: &color.Gray{210},
		},
		"[48;5;254m": {
			inputSeq:   esc("[48;5;254m"),
			expectedFg: nil,
			expectedBg: &color.Gray{220},
		},
		"[48;5;255m": {
			inputSeq:   esc("[48;5;255m"),
			expectedFg: nil,
			expectedBg: &color.Gray{230},
		},
		"[48;5;231m": {
			inputSeq:   esc("[48;5;231m"),
			expectedFg: nil,
			expectedBg: &color.RGBA{255, 255, 255, 255},
		},
		"[31;48;5;16m": {
			inputSeq:   esc("[31;48;5;16m"),
			expectedFg: &color.RGBA{170, 0, 0, 255},
			expectedBg: &color.RGBA{0, 0, 0, 255},
		},
		"[48;5;52m": {
			inputSeq:   esc("[48;5;52m"),
			expectedFg: nil,
			expectedBg: &color.RGBA{95, 0, 0, 255},
		},
		"[48;5;88m": {
			inputSeq:   esc("[48;5;88m"),
			expectedFg: nil,
			expectedBg: &color.RGBA{135, 0, 0, 255},
		},
		"[48;5;124m": {
			inputSeq:   esc("[48;5;124m"),
			expectedFg: nil,
			expectedBg: &color.RGBA{175, 0, 0, 255},
		},
		"[48;5;160m": {
			inputSeq:   esc("[48;5;160m"),
			expectedFg: nil,
			expectedBg: &color.RGBA{215, 0, 0, 255},
		},
		"[48;5;196m": {
			inputSeq:   esc("[48;5;196m"),
			expectedFg: nil,
			expectedBg: &color.RGBA{255, 0, 0, 255},
		},
		"[32;48;5;16m": {
			inputSeq:   esc("[32;48;5;16m"),
			expectedFg: &color.RGBA{0, 170, 0, 255},
			expectedBg: &color.RGBA{0, 0, 0, 255},
		},
		"[48;5;22m": {
			inputSeq:   esc("[48;5;22m"),
			expectedFg: nil,
			expectedBg: &color.RGBA{0, 95, 0, 255},
		},
		"[48;5;28m": {
			inputSeq:   esc("[48;5;28m"),
			expectedFg: nil,
			expectedBg: &color.RGBA{0, 135, 0, 255},
		},
		"[48;5;34m": {
			inputSeq:   esc("[48;5;34m"),
			expectedFg: nil,
			expectedBg: &color.RGBA{0, 175, 0, 255},
		},
		"[48;5;40m": {
			inputSeq:   esc("[48;5;40m"),
			expectedFg: nil,
			expectedBg: &color.RGBA{0, 215, 0, 255},
		},
		"[48;5;46m": {
			inputSeq:   esc("[48;5;46m"),
			expectedFg: nil,
			expectedBg: &color.RGBA{0, 255, 0, 255},
		},
		"[34;48;5;16m": {
			inputSeq:   esc("[34;48;5;16m"),
			expectedFg: &color.RGBA{0, 0, 170, 255},
			expectedBg: &color.RGBA{0, 0, 0, 255},
		},
		"[48;5;17m": {
			inputSeq:   esc("[48;5;17m"),
			expectedFg: nil,
			expectedBg: &color.RGBA{0, 0, 95, 255},
		},
		"[48;5;18m": {
			inputSeq:   esc("[48;5;18m"),
			expectedFg: nil,
			expectedBg: &color.RGBA{0, 0, 135, 255},
		},
		"[48;5;19m": {
			inputSeq:   esc("[48;5;19m"),
			expectedFg: nil,
			expectedBg: &color.RGBA{0, 0, 175, 255},
		},
		"[48;5;20m": {
			inputSeq:   esc("[48;5;20m"),
			expectedFg: nil,
			expectedBg: &color.RGBA{0, 0, 215, 255},
		},
		"[48;5;21m": {
			inputSeq:   esc("[48;5;21m"),
			expectedFg: nil,
			expectedBg: &color.RGBA{0, 0, 255, 255},
		},
		"[33;48;5;16m": {
			inputSeq:   esc("[33;48;5;16m"),
			expectedFg: &color.RGBA{170, 170, 0, 255},
			expectedBg: &color.RGBA{0, 0, 0, 255},
		},
		"[48;5;58m": {
			inputSeq:   esc("[48;5;58m"),
			expectedFg: nil,
			expectedBg: &color.RGBA{95, 95, 0, 255},
		},
		"[48;5;100m": {
			inputSeq:   esc("[48;5;100m"),
			expectedFg: nil,
			expectedBg: &color.RGBA{135, 135, 0, 255},
		},
		"[48;5;142m": {
			inputSeq:   esc("[48;5;142m"),
			expectedFg: nil,
			expectedBg: &color.RGBA{175, 175, 0, 255},
		},
		"[48;5;184m": {
			inputSeq:   esc("[48;5;184m"),
			expectedFg: nil,
			expectedBg: &color.RGBA{215, 215, 0, 255},
		},
		"[48;5;226m": {
			inputSeq:   esc("[48;5;226m"),
			expectedFg: nil,
			expectedBg: &color.RGBA{255, 255, 0, 255},
		},
		"[35;48;5;16m": {
			inputSeq:   esc("[35;48;5;16m"),
			expectedFg: &color.RGBA{170, 0, 170, 255},
			expectedBg: &color.RGBA{0, 0, 0, 255},
		},
		"[48;5;53m": {
			inputSeq:   esc("[48;5;53m"),
			expectedFg: nil,
			expectedBg: &color.RGBA{95, 0, 95, 255},
		},
		"[48;5;90m": {
			inputSeq:   esc("[48;5;90m"),
			expectedFg: nil,
			expectedBg: &color.RGBA{135, 0, 135, 255},
		},
		"[48;5;127m": {
			inputSeq:   esc("[48;5;127m"),
			expectedFg: nil,
			expectedBg: &color.RGBA{175, 0, 175, 255},
		},
		"[48;5;164m": {
			inputSeq:   esc("[48;5;164m"),
			expectedFg: nil,
			expectedBg: &color.RGBA{215, 0, 215, 255},
		},
		"[48;5;201m": {
			inputSeq:   esc("[48;5;201m"),
			expectedFg: nil,
			expectedBg: &color.RGBA{255, 0, 255, 255},
		},
		"[36;48;5;16m": {
			inputSeq:   esc("[36;48;5;16m"),
			expectedFg: &color.RGBA{0, 255, 255, 255},
			expectedBg: &color.RGBA{0, 0, 0, 255},
		},
		"[48;5;23m": {
			inputSeq:   esc("[48;5;23m"),
			expectedFg: nil,
			expectedBg: &color.RGBA{0, 95, 95, 255},
		},
		"[48;5;30m": {
			inputSeq:   esc("[48;5;30m"),
			expectedFg: nil,
			expectedBg: &color.RGBA{0, 135, 135, 255},
		},
		"[48;5;37m": {
			inputSeq:   esc("[48;5;37m"),
			expectedFg: nil,
			expectedBg: &color.RGBA{0, 175, 175, 255},
		},
		"[48;5;44m": {
			inputSeq:   esc("[48;5;44m"),
			expectedFg: nil,
			expectedBg: &color.RGBA{0, 215, 215, 255},
		},
		"[48;5;51m": {
			inputSeq:   esc("[48;5;51m"),
			expectedFg: nil,
			expectedBg: &color.RGBA{0, 255, 255, 255},
		},
	}

	testColor(t, tests)
}

func TestHandleOutput_xterm_256_2(t *testing.T) {
	tests := map[string]struct {
		inputSeq      string
		expectedFg    color.Color
		expectedBg    color.Color
		expectedStyle fyne.TextStyle
	}{
		"[38;5;0m": {
			inputSeq:   esc("[38;5;0m"),
			expectedFg: color.Black,
			expectedBg: nil,
		},
		"[38;5;1m": {
			inputSeq:   esc("[38;5;1m"),
			expectedFg: &color.RGBA{170, 0, 0, 255},
			expectedBg: nil,
		},
		"[38;5;2m": {
			inputSeq:   esc("[38;5;2m"),
			expectedFg: &color.RGBA{0, 170, 0, 255},
			expectedBg: nil,
		},
		"[38;5;3m": {
			inputSeq:   esc("[38;5;3m"),
			expectedFg: &color.RGBA{170, 170, 0, 255},
			expectedBg: nil,
		},
		"[38;5;4m": {
			inputSeq:   esc("[38;5;4m"),
			expectedFg: &color.RGBA{0, 0, 170, 255},
			expectedBg: nil,
		},
		"[38;5;5m": {
			inputSeq:   esc("[38;5;5m"),
			expectedFg: &color.RGBA{170, 0, 170, 255},
			expectedBg: nil,
		},
		"[38;5;6m": {
			inputSeq:   esc("[38;5;6m"),
			expectedFg: &color.RGBA{0, 255, 255, 255},
			expectedBg: nil,
		},
		"[38;5;7m": {
			inputSeq:   esc("[38;5;7m"),
			expectedFg: &color.RGBA{170, 170, 170, 255},
			expectedBg: nil,
		},
		"[38;5;8m": {
			inputSeq:   esc("[38;5;8m"),
			expectedFg: &color.RGBA{85, 85, 85, 255},
			expectedBg: nil,
		},
		"[38;5;9m": {
			inputSeq:   esc("[38;5;9m"),
			expectedFg: &color.RGBA{255, 85, 85, 255},
			expectedBg: nil,
		},
		"[38;5;10m": {
			inputSeq:   esc("[38;5;10m"),
			expectedFg: &color.RGBA{85, 255, 85, 255},
			expectedBg: nil,
		},
		"[38;5;11m": {
			inputSeq:   esc("[38;5;11m"),
			expectedFg: &color.RGBA{255, 255, 85, 255},
			expectedBg: nil,
		},
		"[38;5;12m": {
			inputSeq:   esc("[38;5;12m"),
			expectedFg: &color.RGBA{85, 85, 255, 255},
			expectedBg: nil,
		},
		"[38;5;13m": {
			inputSeq:   esc("[38;5;13m"),
			expectedFg: &color.RGBA{255, 85, 255, 255},
			expectedBg: nil,
		},
		"[38;5;14m": {
			inputSeq:   esc("[38;5;14m"),
			expectedFg: &color.RGBA{85, 255, 255, 255},
			expectedBg: nil,
		},
		"[38;5;15m": {
			inputSeq:   esc("[38;5;15m"),
			expectedFg: &color.RGBA{255, 255, 255, 255}, // White
			expectedBg: nil,
		},
		"[38;5;16m": {
			inputSeq:   esc("[38;5;16m"),
			expectedFg: &color.RGBA{0, 0, 0, 255},
			expectedBg: nil,
		},
		"[38;5;17m": {
			inputSeq:   esc("[38;5;17m"),
			expectedFg: &color.RGBA{0, 0, 95, 255}, // Dark Blue
			expectedBg: nil,
		},
		"[38;5;18m": {
			inputSeq:   esc("[38;5;18m"),
			expectedFg: &color.RGBA{0, 0, 135, 255}, // Dark Green
			expectedBg: nil,
		},
		"[38;5;19m": {
			inputSeq:   esc("[38;5;19m"),
			expectedFg: &color.RGBA{0, 0, 175, 255}, // Dark Cyan
			expectedBg: nil,
		},
		"[38;5;20m": {
			inputSeq:   esc("[38;5;20m"),
			expectedFg: &color.RGBA{0, 0, 215, 255}, // Dark Red
			expectedBg: nil,
		},
		"[38;5;21m": {
			inputSeq:   esc("[38;5;21m"),
			expectedFg: &color.RGBA{0, 0, 255, 255}, // Dark Purple
			expectedBg: nil,
		},
		"[38;5;22m": {
			inputSeq:   esc("[38;5;22m"),
			expectedFg: &color.RGBA{0, 95, 0, 255},
			expectedBg: nil,
		},
		"[38;5;23m": {
			inputSeq:   esc("[38;5;23m"),
			expectedFg: &color.RGBA{0, 95, 95, 255},
			expectedBg: nil,
		},
		"[38;5;24m": {
			inputSeq:   esc("[38;5;24m"),
			expectedFg: &color.RGBA{0, 95, 135, 255},
			expectedBg: nil,
		},
		"[38;5;25m": {
			inputSeq:   esc("[38;5;25m"),
			expectedFg: &color.RGBA{0, 95, 175, 255},
			expectedBg: nil,
		},
		"[38;5;26m": {
			inputSeq:   esc("[38;5;26m"),
			expectedFg: &color.RGBA{0, 95, 215, 255},
			expectedBg: nil,
		},
		"[38;5;27m": {
			inputSeq:   esc("[38;5;27m"),
			expectedFg: &color.RGBA{0, 95, 255, 255},
			expectedBg: nil,
		},
		"[38;5;28m": {
			inputSeq:   esc("[38;5;28m"),
			expectedFg: &color.RGBA{0, 135, 0, 255},
			expectedBg: nil,
		},
		"[38;5;29m": {
			inputSeq:   esc("[38;5;29m"),
			expectedFg: &color.RGBA{0, 135, 95, 255},
			expectedBg: nil,
		},
		"[38;5;30m": {
			inputSeq:   esc("[38;5;30m"),
			expectedFg: &color.RGBA{0, 135, 135, 255}, // Cyan
			expectedBg: nil,
		},
		"[38;5;31m": {
			inputSeq:   esc("[38;5;31m"),
			expectedFg: &color.RGBA{0, 135, 175, 255}, // Dark Green
			expectedBg: nil,
		},
	}

	testColor(t, tests)
}

func TestHandleOutput_24_bit_colour(t *testing.T) {
	tests := map[string]struct {
		inputSeq      string
		expectedFg    color.Color
		expectedBg    color.Color
		expectedStyle fyne.TextStyle
	}{
		"SlateGrey": {
			inputSeq:   esc("[38;2;112;128;144m"),
			expectedFg: &color.RGBA{112, 128, 144, 255},
			expectedBg: nil,
		},
		"OliveDrab": {
			inputSeq:   esc("[38;2;107;142;35m"),
			expectedFg: &color.RGBA{107, 142, 35, 255},
			expectedBg: nil,
		},
		"goldenrod": {
			inputSeq:   esc("[38;2;218;165;32m"),
			expectedFg: &color.RGBA{218, 165, 32, 255},
			expectedBg: nil,
		},
		"SaddleBrown": {
			inputSeq:   esc("[38;2;139;69;19m"),
			expectedFg: &color.RGBA{139, 69, 19, 255},
			expectedBg: nil,
		},
		"DarkViolet (bg)": {
			inputSeq:   esc("[30;48;2;148;0;211m"),
			expectedFg: color.Black,
			expectedBg: &color.RGBA{148, 0, 211, 255},
		},
	}

	testColor(t, tests)
}

func TestHandleOutput_BufferCutoff(t *testing.T) {
	term := New()
	termsize := fyne.NewSize(80, 50)
	term.Resize(termsize)
	term.handleOutput([]byte("\x1b[38;5;64"))
	term.handleOutput([]byte("m40\x1b[38;5;65m41"))
	tg := widget2.NewTermGrid()
	tg.Resize(termsize)
	c1 := &color.RGBA{R: 95, G: 135, A: 255}
	c2 := &color.RGBA{R: 95, G: 135, B: 95, A: 255}

	mono := fyne.TextStyle{Monospace: true}
	tg.Rows = []widget.TextGridRow{
		{
			Cells: []widget.TextGridCell{
				{Rune: '4', Style: &widget.CustomTextGridStyle{FGColor: c1, BGColor: nil, TextStyle: mono}},
				{Rune: '0', Style: &widget.CustomTextGridStyle{FGColor: c1, BGColor: nil, TextStyle: mono}},
				{Rune: '4', Style: &widget.CustomTextGridStyle{FGColor: c2, BGColor: nil, TextStyle: mono}},
				{Rune: '1', Style: &widget.CustomTextGridStyle{FGColor: c2, BGColor: nil, TextStyle: mono}},
			},
		},
	}
	assert.Equal(t, tg.Rows, term.content.Rows)
}
