package widget

import (
	"image/color"

	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type TermGrid struct {
	widget.TextGrid
}

func NewTermGrid() *TermGrid {
	tg := &TermGrid{}
	tg.ExtendBaseWidget(tg)
	return tg
}

// HighlightRange highlight options to the given range
// if highlighting has previously been applied it is enabled
func (t *TermGrid) HighlightRange(blockMode bool, startRow, startCol, endRow, endCol int, o ...HighlightOption) {
	applyHighlight := func(cell *widget.TextGridCell) {
		// Check if already highlighted
		if h, ok := cell.Style.(*HighlightedTextGridStyle); !ok {
			highlightedStyle := &HighlightedTextGridStyle{OriginalStyle: cell.Style, Highlighted: true}
			highlightedStyle.With(o...)
			cell.Style = highlightedStyle
		} else {
			h.Highlighted = true
		}
	}

	t.ForRange(blockMode, startRow, startCol, endRow, endCol, applyHighlight, nil)
}

// ClearHighlightRange disables the highlight style for the given range
func (t *TermGrid) ClearHighlightRange(blockMode bool, startRow, startCol, endRow, endCol int) {
	clearHighlight := func(cell *widget.TextGridCell) {
		// Check if already highlighted
		if h, ok := cell.Style.(*HighlightedTextGridStyle); ok {
			h.Highlighted = false
		}
	}
	t.ForRange(blockMode, startRow, startCol, endRow, endCol, clearHighlight, nil)
}

// HighlightedTextGridStyle defines a style that can be original or highlighted.
type HighlightedTextGridStyle struct {
	OriginalStyle    widget.TextGridStyle
	HighlightedStyle widget.TextGridStyle
	Highlighted      bool
}

// TextColor returns the color of the text, depending on whether it is highlighted.
func (h *HighlightedTextGridStyle) TextColor() color.Color {
	if h.Highlighted {
		return h.HighlightedStyle.TextColor()
	}
	return h.OriginalStyle.TextColor()
}

// BackgroundColor returns the background color, depending on whether it is highlighted.
func (h *HighlightedTextGridStyle) BackgroundColor() color.Color {
	if h.Highlighted {
		return h.HighlightedStyle.BackgroundColor()
	}
	return h.OriginalStyle.BackgroundColor()
}

// HighlightOption defines a function type that can modify a HighlightedTextGridStyle.
type HighlightOption func(h *HighlightedTextGridStyle)

// InvertColor inverts a color c with the given bitmask
func InvertColor(c color.Color, bitmask uint8) color.Color {
	r, g, b, a := c.RGBA()
	return color.RGBA{
		R: uint8(r>>8) ^ bitmask,
		G: uint8(g>>8) ^ bitmask,
		B: uint8(b>>8) ^ bitmask,
		A: uint8(a >> 8),
	}
}

// WithInvert returns a HighlightOption that inverts the colors of the HighlightedTextGridStyle using the provided bitmask.
func WithInvert(bitmask uint8) HighlightOption {
	return func(h *HighlightedTextGridStyle) {
		var fg, bg color.Color
		if h.OriginalStyle != nil {
			fg = h.OriginalStyle.TextColor()
			bg = h.OriginalStyle.BackgroundColor()
		}
		if fg == nil {
			fg = theme.ForegroundColor()
		}
		if bg == nil {
			bg = theme.BackgroundColor()
		}

		h.HighlightedStyle = &widget.CustomTextGridStyle{
			FGColor: InvertColor(fg, bitmask),
			BGColor: InvertColor(bg, bitmask),
		}
	}
}

// With applies one or more HighlightOption functions to the HighlightedTextGridStyle,
// allowing customization of the style.
func (h *HighlightedTextGridStyle) With(options ...HighlightOption) {
	for _, option := range options {
		option(h)
	}
}

// ForRange iterates over a range of cells and rows within a TermGrid, optionally applying a function to each cell and row.
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
// - If startRow or endRow are out of bounds (negative or greater/equal to the number of rows in the TermGrid), they will be adjusted to valid values.
// - If startRow and endRow are the same, the iteration will be limited to the specified column range within that row.
// - When blockMode is true, it iterates through rows from startRow to endRow, applying the cell function for each cell in the specified column range.
// - When blockMode is false, it iterates through individual cells row by row, applying the cell function for each cell and optionally applying the row function for each row.
//
// Example Usage:
// termGrid.ForRange(true, 0, 1, 2, 3, cellFunc, rowFunc) // Iterate in block mode, applying cellFunc to cells in columns 1 to 3 and rowFunc to rows 0 to 2.
// termGrid.ForRange(false, 1, 0, 3, 2, cellFunc, rowFunc) // Iterate cell by cell, applying cellFunc to all cells and rowFunc to rows 1 and 2.
func (t *TermGrid) ForRange(blockMode bool, startRow, startCol, endRow, endCol int, eachCell func(cell *widget.TextGridCell), eachRow func(row *widget.TextGridRow)) {
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
func (t *TermGrid) GetTextRange(blockMode bool, startRow, startCol, endRow, endCol int) string {
	var result []rune

	t.ForRange(blockMode, startRow, startCol, endRow, endCol, func(cell *widget.TextGridCell) {
		result = append(result, cell.Rune)
	}, func(row *widget.TextGridRow) {
		result = append(result, '\n')
	})

	return string(result)
}
