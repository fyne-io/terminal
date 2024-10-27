package terminal

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	widget2 "github.com/fyne-io/terminal/internal/widget"
	"github.com/stretchr/testify/assert"
)

func TestGetSelectedRange(t *testing.T) {
	tests := map[string]struct {
		selStart, selEnd                                   position
		blockMode                                          bool
		wantStartRow, wantStartCol, wantEndRow, wantEndCol int
	}{
		"Positive Selection": {
			selStart:     position{Row: 1, Col: 1},
			selEnd:       position{Row: 1, Col: 5},
			blockMode:    false,
			wantStartRow: 0, wantStartCol: 0, wantEndRow: 0, wantEndCol: 4,
		},
		"Negative Selection Same Row": {
			selStart:     position{Row: 1, Col: 5},
			selEnd:       position{Row: 1, Col: 1},
			blockMode:    false,
			wantStartRow: 0, wantStartCol: 0, wantEndRow: 0, wantEndCol: 4,
		},
		"Positive Selection Different Rows": {
			selStart:     position{Row: 2, Col: 3},
			selEnd:       position{Row: 4, Col: 3},
			blockMode:    false,
			wantStartRow: 1, wantStartCol: 2, wantEndRow: 3, wantEndCol: 2,
		},
		"Negative Selection Different Rows": {
			selStart:     position{Row: 4, Col: 3},
			selEnd:       position{Row: 2, Col: 3},
			blockMode:    false,
			wantStartRow: 1, wantStartCol: 2, wantEndRow: 3, wantEndCol: 2,
		},
		"Block Mode Positive Selection": {
			selStart:     position{Row: 1, Col: 5},
			selEnd:       position{Row: 4, Col: 6},
			blockMode:    true,
			wantStartRow: 0, wantStartCol: 4, wantEndRow: 3, wantEndCol: 5,
		},
		"Block Mode Negative Selection Same Row": {
			selStart:     position{Row: 1, Col: 5},
			selEnd:       position{Row: 1, Col: 1},
			blockMode:    true,
			wantStartRow: 0, wantStartCol: 0, wantEndRow: 0, wantEndCol: 4,
		},
		"Block Mode Negative Column, Positive Rows": {
			selStart:     position{Row: 4, Col: 3},
			selEnd:       position{Row: 2, Col: 2},
			blockMode:    true,
			wantStartRow: 1, wantStartCol: 1, wantEndRow: 3, wantEndCol: 2,
		},
		"Block Mode Negative Column, Negative Rows": {
			selStart:     position{Row: 4, Col: 4},
			selEnd:       position{Row: 3, Col: 3},
			blockMode:    true,
			wantStartRow: 2, wantStartCol: 2, wantEndRow: 3, wantEndCol: 3,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			term := &Terminal{selStart: &tt.selStart, selEnd: &tt.selEnd, blockMode: tt.blockMode}
			gotStartRow, gotStartCol, gotEndRow, gotEndCol := term.getSelectedRange()
			if gotStartRow != tt.wantStartRow || gotStartCol != tt.wantStartCol || gotEndRow != tt.wantEndRow || gotEndCol != tt.wantEndCol {
				t.Errorf("getSelectedRange() = (%d, %d, %d, %d), want (%d, %d, %d, %d)", gotStartRow, gotStartCol, gotEndRow, gotEndCol, tt.wantStartRow, tt.wantStartCol, tt.wantEndRow, tt.wantEndCol)
			}
		})
	}
}

func TestGetTextRange(t *testing.T) {
	grid := widget2.NewTermGrid()
	grid.Rows = []widget.TextGridRow{
		{Cells: []widget.TextGridCell{{Rune: 'A'}, {Rune: 'B'}, {Rune: 'C'}}},
		{Cells: []widget.TextGridCell{{Rune: 'D'}, {Rune: 'E'}, {Rune: 'F'}}},
		{Cells: []widget.TextGridCell{{Rune: 'G'}, {Rune: 'H'}, {Rune: 'I'}}},
	}

	tests := map[string]struct {
		startRow  int
		startCol  int
		endRow    int
		endCol    int
		blockMode bool
		want      string
	}{
		"Full grid":        {0, 0, 2, 2, false, "ABC\nDEF\nGHI"},
		"Almost Full grid": {0, 1, 2, 1, false, "BC\nDEF\nGH"},
		"Sub grid":         {1, 1, 2, 2, false, "EF\nGHI"},
		"Single cell":      {0, 0, 0, 0, false, "A"},
		"Single row":       {0, 0, 0, 2, false, "ABC"},
		"Single full row":  {0, 0, 1, -1, false, "ABC\n"},
		"Block mode":       {0, 1, 2, 2, true, "BC\nEF\nHI"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := widget2.GetTextRange(grid, tc.blockMode, tc.startRow, tc.startCol, tc.endRow, tc.endCol)
			if got != tc.want {
				t.Fatalf("GetTextRange() = %v; want %v", got, tc.want)
			}
		})
	}
}

func TestDoubleTapped(t *testing.T) {
	grid := widget2.NewTermGrid()
	grid.Rows = []widget.TextGridRow{
		{Cells: []widget.TextGridCell{
			{Rune: 'H'}, {Rune: 'e'}, {Rune: 'l'}, {Rune: 'l'}, {Rune: 'o'},
			{Rune: ' '}, {Rune: 'W'}, {Rune: 'o'}, {Rune: 'r'}, {Rune: 'l'},
			{Rune: 'd'}, {Rune: '!'},
		}},
		{Cells: []widget.TextGridCell{
			{Rune: 'T'}, {Rune: 'e'}, {Rune: 's'}, {Rune: 't'}, {Rune: 'i'},
			{Rune: 'n'}, {Rune: 'g'}, {Rune: ' '}, {Rune: '1'}, {Rune: '2'},
			{Rune: '3'}, {Rune: '.'},
		}},
	}

	term := &Terminal{
		content: grid,
	}
	term.Resize(fyne.NewSize(500, 500))

	tests := map[string]struct {
		clickPosition fyne.Position
		expectedWord  string
	}{
		"Double tap on 'Hello'": {
			clickPosition: term.getTextPosition(position{Row: 1, Col: 1}),
			expectedWord:  "Hello",
		},
		"Double tap on 'World'": {
			clickPosition: term.getTextPosition(position{Row: 1, Col: 7}),
			expectedWord:  "World",
		},
		"Double tap on '123'": {
			clickPosition: term.getTextPosition(position{Row: 2, Col: 9}),
			expectedWord:  "123",
		},
		"Double tap on '!' should not select": {
			clickPosition: term.getTextPosition(position{Row: 1, Col: 12}),
			expectedWord:  "",
		},
		"Double tap on '.' should not select": {
			clickPosition: term.getTextPosition(position{Row: 2, Col: 12}),
			expectedWord:  "",
		},
		"Double tap on space between words": {
			clickPosition: term.getTextPosition(position{Row: 1, Col: 6}),
			expectedWord:  "",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			term.clearSelectedText()
			term.DoubleTapped(&fyne.PointEvent{
				Position: tc.clickPosition,
			})

			selectedWord := ""
			if term.hasSelectedText() {
				selectedWord = term.SelectedText()
			}

			assert.Equal(t, tc.expectedWord, selectedWord)
		})
	}
}
