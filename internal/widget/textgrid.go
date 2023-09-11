package widget

import (
	"image/color"

	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// HighlightRange highlight options to the given range
// if highlighting has previously been applied it is enabled
func HighlightRange(t *widget.TextGrid, blockMode bool, startRow, startCol, endRow, endCol int, bitmask byte) {
	applyHighlight := func(cell *widget.TextGridCell) {
		// Check if already highlighted
		if h, ok := cell.Style.(*TermTextGridStyle); !ok {
			if cell.Style != nil {
				cell.Style = NewTermTextGridStyle(cell.Style.TextColor(), cell.Style.BackgroundColor(), bitmask, false)
			} else {
				cell.Style = NewTermTextGridStyle(nil, nil, bitmask, false)
			}
			cell.Style.(*TermTextGridStyle).Highlighted = true

		} else {
			h.Highlighted = true
		}
	}

	forRange(t, blockMode, startRow, startCol, endRow, endCol, applyHighlight, nil)
}

// ClearHighlightRange disables the highlight style for the given range
func ClearHighlightRange(t *widget.TextGrid, blockMode bool, startRow, startCol, endRow, endCol int) {
	clearHighlight := func(cell *widget.TextGridCell) {
		// Check if already highlighted
		if h, ok := cell.Style.(*TermTextGridStyle); ok {
			h.Highlighted = false
		}
	}
	forRange(t, blockMode, startRow, startCol, endRow, endCol, clearHighlight, nil)
}

// GetTextRange retrieves a text range from the TextGrid. It collects the text
// within the specified grid coordinates, starting from (startRow, startCol) and
// ending at (endRow, endCol), and returns it as a string. The behavior of the
// selection depends on the blockMode parameter. If blockMode is true, then
// startCol and endCol apply to each row in the range, creating a block selection.
// If blockMode is false, startCol applies only to the first row, and endCol
// applies only to the last row, resulting in a continuous range.
//
// Parameters:
//   - blockMode: A boolean flag indicating whether to use block mode.
//   - startRow:  The starting row index of the text range.
//   - startCol:  The starting column index of the text range.
//   - endRow:    The ending row index of the text range.
//   - endCol:    The ending column index of the text range.
//
// Returns:
//   - string: The text content within the specified range as a string.
func GetTextRange(t *widget.TextGrid, blockMode bool, startRow, startCol, endRow, endCol int) string {
	var result []rune

	forRange(t, blockMode, startRow, startCol, endRow, endCol, func(cell *widget.TextGridCell) {
		result = append(result, cell.Rune)
	}, func(row *widget.TextGridRow) {
		result = append(result, '\n')
	})

	return string(result)
}

// forRange iterates over a range of cells and rows within a TermGrid, optionally applying a function to each cell and row.
//
// Parameters:
// - blockMode (bool): If true, the iteration is done in block mode, meaning it iterates through rows and applies the cell function for each cell in the specified column range.
// - startRow (int): The starting row index for the iteration. Rows are 0-indexed.
// - startCol (int): The starting column index for the iteration within the starting row. Columns are 0-indexed.
// - endRow (int): The ending row index for the iteration.
// - endCol (int): The ending column index for the iteration within the ending row.
// - eachCell (func(cell *widget.TextGridCell)): A function that takes a pointer to a TextGridCell and is applied to each cell in the specified range. Pass `nil` if you don't want to apply a cell function.
// - eachRow (func(row *widget.TextGridRow)): A function that takes a pointer to a TextGridRow and is applied to each row in the specified range. Pass `nil` if you don't want to apply a row function.
//
// Note:
// - If startRow or endRow are out of bounds (negative or greater/equal to the number of rows in the TextGrid), they will be adjusted to valid values.
// - If startRow and endRow are the same, the iteration will be limited to the specified column range within that row.
// - When blockMode is true, it iterates through rows from startRow to endRow, applying the cell function for each cell in the specified column range.
// - When blockMode is false, it iterates through individual cells row by row, applying the cell function for each cell and optionally applying the row function for each row.
//
// Example Usage:
// forRange(termGrid, true, 0, 1, 2, 3, cellFunc, rowFunc) // Iterate in block mode, applying cellFunc to cells in columns 1 to 3 and rowFunc to rows 0 to 2.
// forRange(termGrid, false, 1, 0, 3, 2, cellFunc, rowFunc) // Iterate cell by cell, applying cellFunc to all cells and rowFunc to rows 1 and 2.
func forRange(t *widget.TextGrid, blockMode bool, startRow, startCol, endRow, endCol int, eachCell func(cell *widget.TextGridCell), eachRow func(row *widget.TextGridRow)) {
	if startRow >= len(t.Rows) || endRow < 0 {
		return
	}
	if startRow < 0 {
		startRow = 0
		startCol = 0
	}
	if endRow >= len(t.Rows) {
		endRow = len(t.Rows) - 1
		endCol = len(t.Rows[endRow].Cells) - 1
	}

	if startRow == endRow {
		if len(t.Rows[startRow].Cells)-1 < endCol {
			endCol = len(t.Rows[startRow].Cells) - 1
		}
		for col := startCol; col <= endCol; col++ {
			if eachCell != nil {
				eachCell(&t.Rows[startRow].Cells[col])
			}
		}
		return
	}

	if blockMode {
		// Iterate through the rows
		for rowNum := startRow; rowNum <= endRow; rowNum++ {
			row := &t.Rows[rowNum]
			if rowNum != startRow && eachRow != nil {
				eachRow(row)
			}

			// Apply the cell function for the cells in the given column range
			for col := startCol; col <= endCol && col < len(row.Cells); col++ {
				if eachCell != nil {
					eachCell(&row.Cells[col])
				}
			}
		}
		return
	}

	// first row
	if eachCell != nil {
		for col := startCol; col < len(t.Rows[startRow].Cells); col++ {
			eachCell(&t.Rows[startRow].Cells[col])
		}
	}

	// possible middle rows
	for rowNum := startRow + 1; rowNum < endRow; rowNum++ {
		if eachRow != nil {
			eachRow(&t.Rows[rowNum])
		}
		for col := 0; col < len(t.Rows[rowNum].Cells); col++ {
			if eachCell != nil {
				eachCell(&t.Rows[rowNum].Cells[col])
			}
		}
	}

	if len(t.Rows[endRow].Cells)-1 < endCol {
		endCol = len(t.Rows[endRow].Cells) - 1
	}
	if eachRow != nil {
		eachRow(&t.Rows[endRow])
	}
	// last row
	for col := 0; col <= endCol; col++ {
		if eachCell != nil {
			eachCell(&t.Rows[endRow].Cells[col])
		}
	}
}

// TermTextGridStyle defines a style that can be original or highlighted.
type TermTextGridStyle struct {
	OriginalTextColor       color.Color
	OriginalBackgroundColor color.Color
	InvertedTextColor       color.Color
	InvertedBackgroundColor color.Color
	Highlighted             bool
	BlinkEnabled            bool
	Blink                   bool
}

// TextColor returns the color of the text, depending on whether it is highlighted.
func (h *TermTextGridStyle) TextColor() color.Color {
	if h.Highlighted {
		if h.BlinkEnabled && h.Blink {
			return h.InvertedBackgroundColor
		}
		return h.InvertedTextColor
	}
	if h.BlinkEnabled && h.Blink {
		return h.OriginalBackgroundColor
	}
	return h.OriginalTextColor
}

// BackgroundColor returns the background color, depending on whether it is highlighted.
func (h *TermTextGridStyle) BackgroundColor() color.Color {
	if h.Highlighted {
		return h.InvertedBackgroundColor
	}
	return h.OriginalBackgroundColor
}

// HighlightOption defines a function type that can modify a TermTextGridStyle.
type HighlightOption func(h *TermTextGridStyle)

// NewTermTextGridStyle creates a new TextGridStyle with the specified foreground (fg) and background (bg)
// colors, as well as a bitmask to control the inversion of colors. If fg or bg is nil, the function
// will use default foreground and background colors from the theme. The bitmask is used to determine
// which color channels should be inverted.
//
// Parameters:
//   - fg: The foreground color.
//   - bg: The background color.
//   - bitmask: The bitmask to control color inversion.
//   - blinkEnabled: Should this cell blink when told to.
//
// Returns:
//
//	A pointer to a TermTextGridStyle initialized with the provided colors and inversion settings.
func NewTermTextGridStyle(fg, bg color.Color, bitmask byte, blinkEnabled bool) widget.TextGridStyle {
	// calculate the inverted colors
	var invertedFg, invertedBg color.Color
	if fg == nil {
		invertedFg = invertColor(theme.ForegroundColor(), bitmask)
	} else {
		invertedFg = invertColor(fg, bitmask)
	}
	if bg == nil {
		invertedBg = invertColor(theme.BackgroundColor(), bitmask)
	} else {
		invertedBg = invertColor(bg, bitmask)
	}

	return &TermTextGridStyle{
		OriginalTextColor:       fg,
		OriginalBackgroundColor: bg,
		InvertedTextColor:       invertedFg,
		InvertedBackgroundColor: invertedBg,
		Highlighted:             false,
		BlinkEnabled:            blinkEnabled,
		Blink:                   false,
	}
}

// invertColor inverts a color c with the given bitmask
func invertColor(c color.Color, bitmask uint8) color.Color {
	r, g, b, a := c.RGBA()
	return color.RGBA{
		R: uint8(r>>8) ^ bitmask,
		G: uint8(g>>8) ^ bitmask,
		B: uint8(b>>8) ^ bitmask,
		A: uint8(a >> 8),
	}
}
