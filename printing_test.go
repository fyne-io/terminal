package terminal

import (
	_ "embed"
	"testing"
	"unicode/utf16"

	"github.com/stretchr/testify/assert"
)

func TestHandleOutput_PrintMode(t *testing.T) {
	tests := map[string]struct {
		inputSeq string
	}{
		"pdf printing": {
			inputSeq: esc("_set printer:_editor:tmp/DBRRPT-20231026-022112-27.3446959766687.pdf") + "\000" + esc("\\"),
		},
	}

	// Iterate through the test cases
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			terminal := New()
			terminal.handleOutput([]byte(test.inputSeq))
		})
	}
}

func TestHandleOutput_Printing(t *testing.T) {
	tests := map[string]struct {
		inputSeq              []byte
		expectedPrintingState bool
		expectedPrintData     []byte
		expectedSpooledData   []byte
	}{
		"start printing": {
			inputSeq:              []byte(esc("[5ithisshouldbeprinted")),
			expectedPrintData:     []byte("thisshouldbeprinted"),
			expectedPrintingState: true,
		},
		"complete printing": {
			inputSeq:              []byte(esc("[5i") + "thisshouldbeprinted" + esc("[4i")),
			expectedSpooledData:   []byte("thisshouldbeprinted"),
			expectedPrintingState: false,
		},
		"printing with embedded esc": {
			inputSeq:              []byte(esc("[5i") + esc("����B�") + esc("[4i")),
			expectedSpooledData:   []byte(esc("����B�")),
			expectedPrintingState: false,
		},
		"UTF-8 Content": {
			inputSeq:              []byte(esc("[5i") + "Hello, 世界!" + esc("[4i")),
			expectedSpooledData:   []byte("Hello, 世界!"),
			expectedPrintingState: false,
		},
		"UTF-16 Content": {
			inputSeq: []byte(esc("[5i") + string(utf16.Decode([]uint16{
				0x0048, 0x0065, 0x006c, 0x006c, 0x006f, 0x002c, 0x0020, 0x4e16, 0x754c, 0x0021,
			})) + esc("[4i")),
			expectedSpooledData:   []byte("Hello, 世界!"),
			expectedPrintingState: false,
		},
		"ISO-8859-1 Content": {
			inputSeq:              []byte(esc("[5iH") + "\xe9llo, W\xf6rld!" + esc("[4i")),
			expectedSpooledData:   []byte{0x48, 0xe9, 0x6c, 0x6c, 0x6f, 0x2c, 0x20, 0x57, 0xf6, 0x72, 0x6c, 0x64, 0x21},
			expectedPrintingState: false,
		},
	}

	// Iterate through the test cases
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			terminal := New()
			var spooledData []byte
			terminal.printer = PrinterFunc(func(d []byte) {
				spooledData = d
			})
			terminal.handleOutput(test.inputSeq)

			assert.Equal(t, test.expectedPrintingState, terminal.state.printing)
			assert.Equal(t, test.expectedSpooledData, spooledData)
			assert.Equal(t, test.expectedPrintData, terminal.printData)
		})
	}
}

//go:embed test_data/chn.pdf
var examplePDFData []byte

func TestHandleOutput_Printing_PDF(t *testing.T) {
	terminal := New()
	var spooledData []byte
	terminal.printer = PrinterFunc(func(d []byte) {
		spooledData = d
	})

	data := []byte{asciiEscape, '[', '5', 'i'}
	data = append(data, examplePDFData...)
	data = append(data, []byte{asciiEscape, '[', '4', 'i'}...)

	for i := 0; i < len(data); i += bufLen {
		end := i + bufLen
		if end > len(data) {
			end = len(data)
		}
		t.Logf("sending chunk")
		terminal.handleOutput(data[i:end])
	}

	assert.Equal(t, spooledData, examplePDFData)
}
