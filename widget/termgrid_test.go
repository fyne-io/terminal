package widget

import (
	"testing"

	"fyne.io/fyne/v2/widget"
)

func TestHighlightRange(t *testing.T) {
	// Define a bitmask
	bitmask := uint8(0xAA)
	// Define a test text grid
	textGrid := &TermGrid{
		&widget.TextGrid{
			Rows: []widget.TextGridRow{
				{Cells: []widget.TextGridCell{{Rune: 'A'}, {Rune: 'B'}, {Rune: 'C'}, {Rune: '*'}}},
				{Cells: []widget.TextGridCell{{Rune: 'D'}, {Rune: 'E'}, {Rune: 'F'}, {Rune: '*'}}},
				{Cells: []widget.TextGridCell{{Rune: 'G'}, {Rune: 'H'}, {Rune: 'I'}, {Rune: '*'}}},
				{Cells: []widget.TextGridCell{{Rune: 'J'}, {Rune: 'K'}, {Rune: 'L'}, {Rune: '*'}}},
			},
		},
	}

	textGrid.HighlightRange(false, 0, 0, 2, 2, WithInvert(bitmask))

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
					highlightedStyle, ok := cell.Style.(*HighlightedTextGridStyle)
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
	// Define a bitmask
	bitmask := uint8(0xAA)

	// Define a test text grid
	textGrid := &TermGrid{
		&widget.TextGrid{
			Rows: []widget.TextGridRow{
				{Cells: []widget.TextGridCell{{Rune: 'A'}, {Rune: 'B'}, {Rune: 'C'}, {Rune: '*'}}},
				{Cells: []widget.TextGridCell{{Rune: 'D'}, {Rune: 'E'}, {Rune: 'F'}, {Rune: '*'}}},
				{Cells: []widget.TextGridCell{{Rune: 'G'}, {Rune: 'H'}, {Rune: 'I'}, {Rune: '*'}}},
				{Cells: []widget.TextGridCell{{Rune: 'J'}, {Rune: 'K'}, {Rune: 'L'}, {Rune: '*'}}},
			},
		},
	}

	textGrid.HighlightRange(false, 0, 0, 2, 2, WithInvert(bitmask))
	textGrid.ClearHighlightRange(false, 0, 0, 2, 2)

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
					highlightedStyle, ok := cell.Style.(*HighlightedTextGridStyle)
					if ok && highlightedStyle.Highlighted != tt.wantHighlight {
						t.Errorf("unexpected highlighted flag at row=%d col=%d: got %v, want %v", row, col, highlightedStyle.Highlighted, tt.wantHighlight)
					}
				}
			}
		})
	}
}
