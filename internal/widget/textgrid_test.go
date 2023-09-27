package widget

import (
	"testing"

	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
)

func TestHighlightRange(t *testing.T) {
	// start the test app for the purpose of the test
	test.NewApp()
	// Define a bitmask
	bitmask := uint8(0xAA)
	// Define a test text grid
	textGrid := widget.NewTextGrid()
	textGrid.Rows = []widget.TextGridRow{
		{Cells: []widget.TextGridCell{{Rune: 'A'}, {Rune: 'B'}, {Rune: 'C'}, {Rune: '*'}}},
		{Cells: []widget.TextGridCell{{Rune: 'D'}, {Rune: 'E'}, {Rune: 'F'}, {Rune: '*'}}},
		{Cells: []widget.TextGridCell{{Rune: 'G'}, {Rune: 'H'}, {Rune: 'I'}, {Rune: '*'}}},
		{Cells: []widget.TextGridCell{{Rune: 'J'}, {Rune: 'K'}, {Rune: 'L'}, {Rune: '*'}}},
	}

	HighlightRange(textGrid, false, 0, 0, 2, 2, bitmask)

	tests := map[string]struct {
		startRow, startCol, endRow, endCol int
		wantHighlight                      bool
	}{
		"0:0 ; 1:1": {0, 0, 1, 1, true},
		"0:0 ; 0:0": {0, 0, 0, 0, true},
		"2:3 ; 3:3": {2, 3, 3, 3, false},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			for row := tt.startRow; row <= tt.endRow; row++ {
				for col := tt.startCol; col <= tt.endCol; col++ {
					cell := &textGrid.Rows[row].Cells[col]
					highlightedStyle, ok := cell.Style.(*TermTextGridStyle)
					if ok != tt.wantHighlight {
						t.Errorf("unexpected highlight status at row=%d col=%d: got %v, want %v", row, col, ok, tt.wantHighlight)
					}
					if ok && highlightedStyle.Highlighted != tt.wantHighlight {
						t.Errorf("unexpected highlighted flag at row=%d col=%d: got %v, want %v", row, col, highlightedStyle.Highlighted, tt.wantHighlight)
					}
				}
			}
		})
	}
}

func TestClearHighlightRange(t *testing.T) {
	// start the test app for the purpose of the test
	test.NewApp()
	// Define a bitmask
	bitmask := uint8(0xAA)

	// Define a test text grid
	textGrid := widget.NewTextGrid()
	textGrid.Rows = []widget.TextGridRow{
		{Cells: []widget.TextGridCell{{Rune: 'A'}, {Rune: 'B'}, {Rune: 'C'}, {Rune: '*'}}},
		{Cells: []widget.TextGridCell{{Rune: 'D'}, {Rune: 'E'}, {Rune: 'F'}, {Rune: '*'}}},
		{Cells: []widget.TextGridCell{{Rune: 'G'}, {Rune: 'H'}, {Rune: 'I'}, {Rune: '*'}}},
		{Cells: []widget.TextGridCell{{Rune: 'J'}, {Rune: 'K'}, {Rune: 'L'}, {Rune: '*'}}},
	}

	HighlightRange(textGrid, false, 0, 0, 2, 2, bitmask)
	ClearHighlightRange(textGrid, false, 0, 0, 2, 2)

	tests := map[string]struct {
		startRow, startCol, endRow, endCol int
		wantHighlight                      bool
	}{
		"0:0 ; 1:1": {0, 0, 1, 1, false},
		"0:0 ; 0:0": {0, 0, 0, 0, false},
		"2:3 ; 3:3": {2, 3, 3, 3, false},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			for row := tt.startRow; row <= tt.endRow; row++ {
				for col := tt.startCol; col <= tt.endCol; col++ {
					cell := &textGrid.Rows[row].Cells[col]
					highlightedStyle, ok := cell.Style.(*TermTextGridStyle)
					if ok && highlightedStyle.Highlighted != tt.wantHighlight {
						t.Errorf("unexpected highlighted flag at row=%d col=%d: got %v, want %v", row, col, highlightedStyle.Highlighted, tt.wantHighlight)
					}
				}
			}
		})
	}
}

func TestGetTextRange(t *testing.T) {
	// start the test app for the purpose of the test
	test.NewApp()
	// Prepare the text grid for the tests
	textGrid := widget.NewTextGrid()
	textGrid.Rows = []widget.TextGridRow{
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
			got := GetTextRange(textGrid, tc.blockMode, tc.startRow, tc.startCol, tc.endRow, tc.endCol)
			if got != tc.want {
				t.Fatalf("GetTextRange() = %v; want %v", got, tc.want)
			}
		})
	}
}
