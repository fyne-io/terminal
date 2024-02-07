package terminal

import (
	"fmt"
	"image/color"
	"reflect"
	"testing"
)

func esc(s string) string {
	return fmt.Sprintf("%s%s", string(byte(asciiEscape)), s)
}

func testColor(t *testing.T, tests map[string]struct {
	inputSeq     string
	expectedFg   color.Color
	expectedBg   color.Color
	expectedBold bool
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
			if terminal.bold != test.expectedBold {
				t.Errorf("Bold flag mismatch. Got %v, expected %v", terminal.bold, test.expectedBold)
			}
		})
	}
}

func TestHandleOutput_Text(t *testing.T) {
	tests := map[string]struct {
		inputSeq        string
		expectBold      bool
		expectUnderline bool
	}{
		"bold": {
			inputSeq:   esc("[1m"),
			expectBold: true,
		},
		"underline": {
			inputSeq:        esc("[4m"),
			expectUnderline: true,
		},
		"bold and underline": {
			inputSeq:        esc("[1m") + esc("[4m"),
			expectBold:      true,
			expectUnderline: true,
		},
	}

	// Iterate through the test cases
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			terminal := New()
			terminal.handleOutput([]byte(test.inputSeq))

			// Verify the actual results match the expected results
			if terminal.underlined != test.expectUnderline {
				t.Errorf("Bold flag mismatch. Got %v, expected %v", terminal.underlined, test.expectUnderline)
			}

			if terminal.bold != test.expectBold {
				t.Errorf("Bold flag mismatch. Got %v, expected %v", terminal.bold, test.expectBold)
			}
		})
	}
}

func TestHandleOutput_Normal_Text(t *testing.T) {
	tests := map[string]struct {
		inputSeq     string
		expectedFg   color.Color
		expectedBg   color.Color
		expectedBold bool
	}{
		"reverse video": {
			inputSeq:     esc("[7m"),
			expectedFg:   color.NRGBA{R: 34, G: 34, B: 34, A: 255},
			expectedBg:   color.NRGBA{R: 255, G: 255, B: 255, A: 255},
			expectedBold: false,
		},
		"reverse video and bold": {
			inputSeq:     esc("[7m") + esc("[1m"),
			expectedFg:   color.NRGBA{R: 34, G: 34, B: 34, A: 255},
			expectedBg:   color.NRGBA{R: 255, G: 255, B: 255, A: 255},
			expectedBold: true,
		},
		"reverse video and bold then reset": {
			inputSeq:     esc("[7m") + esc("[1m") + esc("[m"),
			expectedFg:   nil,
			expectedBg:   nil,
			expectedBold: false,
		},
	}

	testColor(t, tests)
}

func TestHandleOutput_ANSI_Colors(t *testing.T) {
	tests := map[string]struct {
		inputSeq     string
		expectedFg   color.Color
		expectedBg   color.Color
		expectedBold bool
	}{
		"[30m": {
			inputSeq:     esc("[30m"),
			expectedFg:   color.Black,
			expectedBg:   nil,
			expectedBold: false,
		},
		"[31m": {
			inputSeq:     esc("[31m"),
			expectedFg:   &color.RGBA{170, 0, 0, 255},
			expectedBg:   nil,
			expectedBold: false,
		},
		"[32m": {
			inputSeq:     esc("[32m"),
			expectedFg:   &color.RGBA{0, 170, 0, 255},
			expectedBg:   nil,
			expectedBold: false,
		},
		"[33m": {
			inputSeq:     esc("[33m"),
			expectedFg:   &color.RGBA{170, 170, 0, 255},
			expectedBg:   nil,
			expectedBold: false,
		},
		"[34m": {
			inputSeq:     esc("[34m"),
			expectedFg:   &color.RGBA{0, 0, 170, 255},
			expectedBg:   nil,
			expectedBold: false,
		},
		"[35m": {
			inputSeq:     esc("[35m"),
			expectedFg:   &color.RGBA{170, 0, 170, 255},
			expectedBg:   nil,
			expectedBold: false,
		},
		"[36m": {
			inputSeq:     esc("[36m"),
			expectedFg:   &color.RGBA{0, 255, 255, 255},
			expectedBg:   nil,
			expectedBold: false,
		},
		"[37m": {
			inputSeq:     esc("[37m"),
			expectedFg:   &color.RGBA{170, 170, 170, 255},
			expectedBg:   nil,
			expectedBold: false,
		},
		"bold": {
			inputSeq:     esc("[1m") + esc("[37m"),
			expectedFg:   &color.RGBA{170, 170, 170, 255},
			expectedBg:   nil,
			expectedBold: true,
		},
		"reverse video": {
			inputSeq:     esc("[7m") + esc("[37m"),
			expectedFg:   &color.RGBA{170, 170, 170, 255},
			expectedBg:   color.NRGBA{255, 255, 255, 255},
			expectedBold: false,
		},
	}

	testColor(t, tests)
}

func TestHandleOutput_xterm_bright(t *testing.T) {
	tests := map[string]struct {
		inputSeq     string
		expectedFg   color.Color
		expectedBg   color.Color
		expectedBold bool
	}{
		"[90m": {
			inputSeq:     esc("[90m"),
			expectedFg:   &color.RGBA{85, 85, 85, 255},
			expectedBg:   nil,
			expectedBold: false,
		},
		"[100m": {
			inputSeq:     esc("[100m"),
			expectedFg:   nil,
			expectedBg:   &color.RGBA{85, 85, 85, 255},
			expectedBold: false,
		},
		"[91m": {
			inputSeq:     esc("[91m"),
			expectedFg:   &color.RGBA{255, 85, 85, 255},
			expectedBg:   nil,
			expectedBold: false,
		},
		"[101m": {
			inputSeq:     esc("[101m"),
			expectedFg:   nil,
			expectedBg:   &color.RGBA{255, 85, 85, 255},
			expectedBold: false,
		},
		"[92m": {
			inputSeq:     esc("[92m"),
			expectedFg:   &color.RGBA{85, 255, 85, 255},
			expectedBg:   nil,
			expectedBold: false,
		},
		"[102m": {
			inputSeq:     esc("[102m"),
			expectedFg:   nil,
			expectedBg:   &color.RGBA{85, 255, 85, 255},
			expectedBold: false,
		},
		"[93m": {
			inputSeq:     esc("[93m"),
			expectedFg:   &color.RGBA{255, 255, 85, 255},
			expectedBg:   nil,
			expectedBold: false,
		},
		"[103m": {
			inputSeq:     esc("[103m"),
			expectedFg:   nil,
			expectedBg:   &color.RGBA{255, 255, 85, 255},
			expectedBold: false,
		},
		"[94m": {
			inputSeq:     esc("[94m"),
			expectedFg:   &color.RGBA{85, 85, 255, 255},
			expectedBg:   nil,
			expectedBold: false,
		},
		"[104m": {
			inputSeq:     esc("[104m"),
			expectedFg:   nil,
			expectedBg:   &color.RGBA{85, 85, 255, 255},
			expectedBold: false,
		},
		"[95m": {
			inputSeq:     esc("[95m"),
			expectedFg:   &color.RGBA{255, 85, 255, 255},
			expectedBg:   nil,
			expectedBold: false,
		},
		"[105m": {
			inputSeq:     esc("[105m"),
			expectedFg:   nil,
			expectedBg:   &color.RGBA{255, 85, 255, 255},
			expectedBold: false,
		},
		"[96m": {
			inputSeq:     esc("[96m"),
			expectedFg:   &color.RGBA{85, 255, 255, 255},
			expectedBg:   nil,
			expectedBold: false,
		},

		"[106m": {
			inputSeq:     esc("[106m"),
			expectedFg:   nil,
			expectedBg:   &color.RGBA{85, 255, 255, 255},
			expectedBold: false,
		},
		"[97m": {
			inputSeq:     esc("[97m"),
			expectedFg:   &color.RGBA{255, 255, 255, 255},
			expectedBg:   nil,
			expectedBold: false,
		},
		"[107m": {
			inputSeq:     esc("[107m"),
			expectedFg:   nil,
			expectedBg:   &color.RGBA{255, 255, 255, 255},
			expectedBold: false,
		},
	}

	testColor(t, tests)
}

func TestHandleOutput_xterm_256_1(t *testing.T) {
	tests := map[string]struct {
		inputSeq     string
		expectedFg   color.Color
		expectedBg   color.Color
		expectedBold bool
	}{
		"[48;5;16m": {
			inputSeq:     esc("[48;5;16m"),
			expectedFg:   nil,
			expectedBg:   &color.RGBA{0, 0, 0, 255},
			expectedBold: false,
		},
		"[48;5;232m": {
			inputSeq:     esc("[48;5;232m"),
			expectedFg:   nil,
			expectedBg:   &color.Gray{0},
			expectedBold: false,
		},
		"[48;5;233m": {
			inputSeq:     esc("[48;5;233m"),
			expectedFg:   nil,
			expectedBg:   &color.Gray{10},
			expectedBold: false,
		},
		"[48;5;234m": {
			inputSeq:     esc("[48;5;234m"),
			expectedFg:   nil,
			expectedBg:   &color.Gray{20},
			expectedBold: false,
		},
		"[48;5;235m": {
			inputSeq:     esc("[48;5;235m"),
			expectedFg:   nil,
			expectedBg:   &color.Gray{30},
			expectedBold: false,
		},
		"[48;5;236m": {
			inputSeq:     esc("[48;5;236m"),
			expectedFg:   nil,
			expectedBg:   &color.Gray{40},
			expectedBold: false,
		},
		"[48;5;237m": {
			inputSeq:     esc("[48;5;237m"),
			expectedFg:   nil,
			expectedBg:   &color.Gray{50},
			expectedBold: false,
		},
		"[48;5;238m": {
			inputSeq:     esc("[48;5;238m"),
			expectedFg:   nil,
			expectedBg:   &color.Gray{60},
			expectedBold: false,
		},
		"[48;5;239m": {
			inputSeq:     esc("[48;5;239m"),
			expectedFg:   nil,
			expectedBg:   &color.Gray{70},
			expectedBold: false,
		},
		"[48;5;240m": {
			inputSeq:     esc("[48;5;240m"),
			expectedFg:   nil,
			expectedBg:   &color.Gray{80},
			expectedBold: false,
		},
		"[48;5;241m": {
			inputSeq:     esc("[48;5;241m"),
			expectedFg:   nil,
			expectedBg:   &color.Gray{90},
			expectedBold: false,
		},
		"[48;5;242m": {
			inputSeq:     esc("[48;5;242m"),
			expectedFg:   nil,
			expectedBg:   &color.Gray{100},
			expectedBold: false,
		},
		"[48;5;243m": {
			inputSeq:     esc("[48;5;243m"),
			expectedFg:   nil,
			expectedBg:   &color.Gray{110},
			expectedBold: false,
		},
		"[48;5;244m": {
			inputSeq:     esc("[48;5;244m"),
			expectedFg:   nil,
			expectedBg:   &color.Gray{120},
			expectedBold: false,
		},
		"[48;5;245m": {
			inputSeq:     esc("[48;5;245m"),
			expectedFg:   nil,
			expectedBg:   &color.Gray{130},
			expectedBold: false,
		},
		"[48;5;246m": {
			inputSeq:     esc("[48;5;246m"),
			expectedFg:   nil,
			expectedBg:   &color.Gray{140},
			expectedBold: false,
		},
		"[48;5;247m": {
			inputSeq:     esc("[48;5;247m"),
			expectedFg:   nil,
			expectedBg:   &color.Gray{150},
			expectedBold: false,
		},
		"[48;5;248m": {
			inputSeq:     esc("[48;5;248m"),
			expectedFg:   nil,
			expectedBg:   &color.Gray{160},
			expectedBold: false,
		},
		"[48;5;249m": {
			inputSeq:     esc("[48;5;249m"),
			expectedFg:   nil,
			expectedBg:   &color.Gray{170},
			expectedBold: false,
		},
		"[48;5;250m": {
			inputSeq:     esc("[48;5;250m"),
			expectedFg:   nil,
			expectedBg:   &color.Gray{180},
			expectedBold: false,
		},
		"[48;5;251m": {
			inputSeq:     esc("[48;5;251m"),
			expectedFg:   nil,
			expectedBg:   &color.Gray{190},
			expectedBold: false,
		},
		"[48;5;252m": {
			inputSeq:     esc("[48;5;252m"),
			expectedFg:   nil,
			expectedBg:   &color.Gray{200},
			expectedBold: false,
		},
		"[48;5;253m": {
			inputSeq:     esc("[48;5;253m"),
			expectedFg:   nil,
			expectedBg:   &color.Gray{210},
			expectedBold: false,
		},
		"[48;5;254m": {
			inputSeq:     esc("[48;5;254m"),
			expectedFg:   nil,
			expectedBg:   &color.Gray{220},
			expectedBold: false,
		},
		"[48;5;255m": {
			inputSeq:     esc("[48;5;255m"),
			expectedFg:   nil,
			expectedBg:   &color.Gray{230},
			expectedBold: false,
		},
		"[48;5;231m": {
			inputSeq:     esc("[48;5;231m"),
			expectedFg:   nil,
			expectedBg:   &color.RGBA{255, 255, 255, 255},
			expectedBold: false,
		},
		"[31;48;5;16m": {
			inputSeq:     esc("[31;48;5;16m"),
			expectedFg:   &color.RGBA{170, 0, 0, 255},
			expectedBg:   &color.RGBA{0, 0, 0, 255},
			expectedBold: false,
		},
		"[48;5;52m": {
			inputSeq:     esc("[48;5;52m"),
			expectedFg:   nil,
			expectedBg:   &color.RGBA{95, 0, 0, 255},
			expectedBold: false,
		},
		"[48;5;88m": {
			inputSeq:     esc("[48;5;88m"),
			expectedFg:   nil,
			expectedBg:   &color.RGBA{135, 0, 0, 255},
			expectedBold: false,
		},
		"[48;5;124m": {
			inputSeq:     esc("[48;5;124m"),
			expectedFg:   nil,
			expectedBg:   &color.RGBA{175, 0, 0, 255},
			expectedBold: false,
		},
		"[48;5;160m": {
			inputSeq:     esc("[48;5;160m"),
			expectedFg:   nil,
			expectedBg:   &color.RGBA{215, 0, 0, 255},
			expectedBold: false,
		},
		"[48;5;196m": {
			inputSeq:     esc("[48;5;196m"),
			expectedFg:   nil,
			expectedBg:   &color.RGBA{255, 0, 0, 255},
			expectedBold: false,
		},
		"[32;48;5;16m": {
			inputSeq:     esc("[32;48;5;16m"),
			expectedFg:   &color.RGBA{0, 170, 0, 255},
			expectedBg:   &color.RGBA{0, 0, 0, 255},
			expectedBold: false,
		},
		"[48;5;22m": {
			inputSeq:     esc("[48;5;22m"),
			expectedFg:   nil,
			expectedBg:   &color.RGBA{0, 95, 0, 255},
			expectedBold: false,
		},
		"[48;5;28m": {
			inputSeq:     esc("[48;5;28m"),
			expectedFg:   nil,
			expectedBg:   &color.RGBA{0, 135, 0, 255},
			expectedBold: false,
		},
		"[48;5;34m": {
			inputSeq:     esc("[48;5;34m"),
			expectedFg:   nil,
			expectedBg:   &color.RGBA{0, 175, 0, 255},
			expectedBold: false,
		},
		"[48;5;40m": {
			inputSeq:     esc("[48;5;40m"),
			expectedFg:   nil,
			expectedBg:   &color.RGBA{0, 215, 0, 255},
			expectedBold: false,
		},
		"[48;5;46m": {
			inputSeq:     esc("[48;5;46m"),
			expectedFg:   nil,
			expectedBg:   &color.RGBA{0, 255, 0, 255},
			expectedBold: false,
		},
		"[34;48;5;16m": {
			inputSeq:     esc("[34;48;5;16m"),
			expectedFg:   &color.RGBA{0, 0, 170, 255},
			expectedBg:   &color.RGBA{0, 0, 0, 255},
			expectedBold: false,
		},
		"[48;5;17m": {
			inputSeq:     esc("[48;5;17m"),
			expectedFg:   nil,
			expectedBg:   &color.RGBA{0, 0, 95, 255},
			expectedBold: false,
		},
		"[48;5;18m": {
			inputSeq:     esc("[48;5;18m"),
			expectedFg:   nil,
			expectedBg:   &color.RGBA{0, 0, 135, 255},
			expectedBold: false,
		},
		"[48;5;19m": {
			inputSeq:     esc("[48;5;19m"),
			expectedFg:   nil,
			expectedBg:   &color.RGBA{0, 0, 175, 255},
			expectedBold: false,
		},
		"[48;5;20m": {
			inputSeq:     esc("[48;5;20m"),
			expectedFg:   nil,
			expectedBg:   &color.RGBA{0, 0, 215, 255},
			expectedBold: false,
		},
		"[48;5;21m": {
			inputSeq:     esc("[48;5;21m"),
			expectedFg:   nil,
			expectedBg:   &color.RGBA{0, 0, 255, 255},
			expectedBold: false,
		},
		"[33;48;5;16m": {
			inputSeq:     esc("[33;48;5;16m"),
			expectedFg:   &color.RGBA{170, 170, 0, 255},
			expectedBg:   &color.RGBA{0, 0, 0, 255},
			expectedBold: false,
		},
		"[48;5;58m": {
			inputSeq:     esc("[48;5;58m"),
			expectedFg:   nil,
			expectedBg:   &color.RGBA{95, 95, 0, 255},
			expectedBold: false,
		},
		"[48;5;100m": {
			inputSeq:     esc("[48;5;100m"),
			expectedFg:   nil,
			expectedBg:   &color.RGBA{135, 135, 0, 255},
			expectedBold: false,
		},
		"[48;5;142m": {
			inputSeq:     esc("[48;5;142m"),
			expectedFg:   nil,
			expectedBg:   &color.RGBA{175, 175, 0, 255},
			expectedBold: false,
		},
		"[48;5;184m": {
			inputSeq:     esc("[48;5;184m"),
			expectedFg:   nil,
			expectedBg:   &color.RGBA{215, 215, 0, 255},
			expectedBold: false,
		},
		"[48;5;226m": {
			inputSeq:     esc("[48;5;226m"),
			expectedFg:   nil,
			expectedBg:   &color.RGBA{255, 255, 0, 255},
			expectedBold: false,
		},
		"[35;48;5;16m": {
			inputSeq:     esc("[35;48;5;16m"),
			expectedFg:   &color.RGBA{170, 0, 170, 255},
			expectedBg:   &color.RGBA{0, 0, 0, 255},
			expectedBold: false,
		},
		"[48;5;53m": {
			inputSeq:     esc("[48;5;53m"),
			expectedFg:   nil,
			expectedBg:   &color.RGBA{95, 0, 95, 255},
			expectedBold: false,
		},
		"[48;5;90m": {
			inputSeq:     esc("[48;5;90m"),
			expectedFg:   nil,
			expectedBg:   &color.RGBA{135, 0, 135, 255},
			expectedBold: false,
		},
		"[48;5;127m": {
			inputSeq:     esc("[48;5;127m"),
			expectedFg:   nil,
			expectedBg:   &color.RGBA{175, 0, 175, 255},
			expectedBold: false,
		},
		"[48;5;164m": {
			inputSeq:     esc("[48;5;164m"),
			expectedFg:   nil,
			expectedBg:   &color.RGBA{215, 0, 215, 255},
			expectedBold: false,
		},
		"[48;5;201m": {
			inputSeq:     esc("[48;5;201m"),
			expectedFg:   nil,
			expectedBg:   &color.RGBA{255, 0, 255, 255},
			expectedBold: false,
		},
		"[36;48;5;16m": {
			inputSeq:     esc("[36;48;5;16m"),
			expectedFg:   &color.RGBA{0, 255, 255, 255},
			expectedBg:   &color.RGBA{0, 0, 0, 255},
			expectedBold: false,
		},
		"[48;5;23m": {
			inputSeq:     esc("[48;5;23m"),
			expectedFg:   nil,
			expectedBg:   &color.RGBA{0, 95, 95, 255},
			expectedBold: false,
		},
		"[48;5;30m": {
			inputSeq:     esc("[48;5;30m"),
			expectedFg:   nil,
			expectedBg:   &color.RGBA{0, 135, 135, 255},
			expectedBold: false,
		},
		"[48;5;37m": {
			inputSeq:     esc("[48;5;37m"),
			expectedFg:   nil,
			expectedBg:   &color.RGBA{0, 175, 175, 255},
			expectedBold: false,
		},
		"[48;5;44m": {
			inputSeq:     esc("[48;5;44m"),
			expectedFg:   nil,
			expectedBg:   &color.RGBA{0, 215, 215, 255},
			expectedBold: false,
		},
		"[48;5;51m": {
			inputSeq:     esc("[48;5;51m"),
			expectedFg:   nil,
			expectedBg:   &color.RGBA{0, 255, 255, 255},
			expectedBold: false,
		},
	}

	testColor(t, tests)
}

func TestHandleOutput_xterm_256_2(t *testing.T) {
	tests := map[string]struct {
		inputSeq     string
		expectedFg   color.Color
		expectedBg   color.Color
		expectedBold bool
	}{
		"[38;5;0m": {
			inputSeq:     esc("[38;5;0m"),
			expectedFg:   color.Black,
			expectedBg:   nil,
			expectedBold: false,
		},
		"[38;5;1m": {
			inputSeq:     esc("[38;5;1m"),
			expectedFg:   &color.RGBA{170, 0, 0, 255},
			expectedBg:   nil,
			expectedBold: false,
		},
		"[38;5;2m": {
			inputSeq:     esc("[38;5;2m"),
			expectedFg:   &color.RGBA{0, 170, 0, 255},
			expectedBg:   nil,
			expectedBold: false,
		},
		"[38;5;3m": {
			inputSeq:     esc("[38;5;3m"),
			expectedFg:   &color.RGBA{170, 170, 0, 255},
			expectedBg:   nil,
			expectedBold: false,
		},
		"[38;5;4m": {
			inputSeq:     esc("[38;5;4m"),
			expectedFg:   &color.RGBA{0, 0, 170, 255},
			expectedBg:   nil,
			expectedBold: false,
		},
		"[38;5;5m": {
			inputSeq:     esc("[38;5;5m"),
			expectedFg:   &color.RGBA{170, 0, 170, 255},
			expectedBg:   nil,
			expectedBold: false,
		},
		"[38;5;6m": {
			inputSeq:     esc("[38;5;6m"),
			expectedFg:   &color.RGBA{0, 255, 255, 255},
			expectedBg:   nil,
			expectedBold: false,
		},
		"[38;5;7m": {
			inputSeq:     esc("[38;5;7m"),
			expectedFg:   &color.RGBA{170, 170, 170, 255},
			expectedBg:   nil,
			expectedBold: false,
		},
		"[38;5;8m": {
			inputSeq:     esc("[38;5;8m"),
			expectedFg:   &color.RGBA{85, 85, 85, 255},
			expectedBg:   nil,
			expectedBold: false,
		},
		"[38;5;9m": {
			inputSeq:     esc("[38;5;9m"),
			expectedFg:   &color.RGBA{255, 85, 85, 255},
			expectedBg:   nil,
			expectedBold: false,
		},
		"[38;5;10m": {
			inputSeq:     esc("[38;5;10m"),
			expectedFg:   &color.RGBA{85, 255, 85, 255},
			expectedBg:   nil,
			expectedBold: false,
		},
		"[38;5;11m": {
			inputSeq:     esc("[38;5;11m"),
			expectedFg:   &color.RGBA{255, 255, 85, 255},
			expectedBg:   nil,
			expectedBold: false,
		},
		"[38;5;12m": {
			inputSeq:     esc("[38;5;12m"),
			expectedFg:   &color.RGBA{85, 85, 255, 255},
			expectedBg:   nil,
			expectedBold: false,
		},
		"[38;5;13m": {
			inputSeq:     esc("[38;5;13m"),
			expectedFg:   &color.RGBA{255, 85, 255, 255},
			expectedBg:   nil,
			expectedBold: false,
		},
		"[38;5;14m": {
			inputSeq:     esc("[38;5;14m"),
			expectedFg:   &color.RGBA{85, 255, 255, 255},
			expectedBg:   nil,
			expectedBold: false,
		},
		"[38;5;15m": {
			inputSeq:     esc("[38;5;15m"),
			expectedFg:   &color.RGBA{255, 255, 255, 255}, // White
			expectedBg:   nil,
			expectedBold: false,
		},
		"[38;5;16m": {
			inputSeq:     esc("[38;5;16m"),
			expectedFg:   &color.RGBA{0, 0, 0, 255},
			expectedBg:   nil,
			expectedBold: false,
		},
		"[38;5;17m": {
			inputSeq:     esc("[38;5;17m"),
			expectedFg:   &color.RGBA{0, 0, 95, 255}, // Dark Blue
			expectedBg:   nil,
			expectedBold: false,
		},
		"[38;5;18m": {
			inputSeq:     esc("[38;5;18m"),
			expectedFg:   &color.RGBA{0, 0, 135, 255}, // Dark Green
			expectedBg:   nil,
			expectedBold: false,
		},
		"[38;5;19m": {
			inputSeq:     esc("[38;5;19m"),
			expectedFg:   &color.RGBA{0, 0, 175, 255}, // Dark Cyan
			expectedBg:   nil,
			expectedBold: false,
		},
		"[38;5;20m": {
			inputSeq:     esc("[38;5;20m"),
			expectedFg:   &color.RGBA{0, 0, 215, 255}, // Dark Red
			expectedBg:   nil,
			expectedBold: false,
		},
		"[38;5;21m": {
			inputSeq:     esc("[38;5;21m"),
			expectedFg:   &color.RGBA{0, 0, 255, 255}, // Dark Purple
			expectedBg:   nil,
			expectedBold: false,
		},
		"[38;5;22m": {
			inputSeq:     esc("[38;5;22m"),
			expectedFg:   &color.RGBA{0, 95, 0, 255},
			expectedBg:   nil,
			expectedBold: false,
		},
		"[38;5;23m": {
			inputSeq:     esc("[38;5;23m"),
			expectedFg:   &color.RGBA{0, 95, 95, 255},
			expectedBg:   nil,
			expectedBold: false,
		},
		"[38;5;24m": {
			inputSeq:     esc("[38;5;24m"),
			expectedFg:   &color.RGBA{0, 95, 135, 255},
			expectedBg:   nil,
			expectedBold: false,
		},
		"[38;5;25m": {
			inputSeq:     esc("[38;5;25m"),
			expectedFg:   &color.RGBA{0, 95, 175, 255},
			expectedBg:   nil,
			expectedBold: false,
		},
		"[38;5;26m": {
			inputSeq:     esc("[38;5;26m"),
			expectedFg:   &color.RGBA{0, 95, 215, 255},
			expectedBg:   nil,
			expectedBold: false,
		},
		"[38;5;27m": {
			inputSeq:     esc("[38;5;27m"),
			expectedFg:   &color.RGBA{0, 95, 255, 255},
			expectedBg:   nil,
			expectedBold: false,
		},
		"[38;5;28m": {
			inputSeq:     esc("[38;5;28m"),
			expectedFg:   &color.RGBA{0, 135, 0, 255},
			expectedBg:   nil,
			expectedBold: false,
		},
		"[38;5;29m": {
			inputSeq:     esc("[38;5;29m"),
			expectedFg:   &color.RGBA{0, 135, 95, 255},
			expectedBg:   nil,
			expectedBold: false,
		},
		"[38;5;30m": {
			inputSeq:     esc("[38;5;30m"),
			expectedFg:   &color.RGBA{0, 135, 135, 255}, // Cyan
			expectedBg:   nil,
			expectedBold: false,
		},
		"[38;5;31m": {
			inputSeq:     esc("[38;5;31m"),
			expectedFg:   &color.RGBA{0, 135, 175, 255}, // Dark Green
			expectedBg:   nil,
			expectedBold: false,
		},
	}

	testColor(t, tests)
}

func TestHandleOutput_24_bit_colour(t *testing.T) {
	tests := map[string]struct {
		inputSeq     string
		expectedFg   color.Color
		expectedBg   color.Color
		expectedBold bool
	}{
		"SlateGrey": {
			inputSeq:     esc("[38;2;112;128;144m"),
			expectedFg:   &color.RGBA{112, 128, 144, 255},
			expectedBg:   nil,
			expectedBold: false,
		},
		"OliveDrab": {
			inputSeq:     esc("[38;2;107;142;35m"),
			expectedFg:   &color.RGBA{107, 142, 35, 255},
			expectedBg:   nil,
			expectedBold: false,
		},
		"goldenrod": {
			inputSeq:     esc("[38;2;218;165;32m"),
			expectedFg:   &color.RGBA{218, 165, 32, 255},
			expectedBg:   nil,
			expectedBold: false,
		},
		"SaddleBrown": {
			inputSeq:     esc("[38;2;139;69;19m"),
			expectedFg:   &color.RGBA{139, 69, 19, 255},
			expectedBg:   nil,
			expectedBold: false,
		},
		"DarkViolet (bg)": {
			inputSeq:     esc("[30;48;2;148;0;211m"),
			expectedFg:   color.Black,
			expectedBg:   &color.RGBA{148, 0, 211, 255},
			expectedBold: false,
		},
	}

	testColor(t, tests)
}
