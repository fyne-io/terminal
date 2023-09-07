package widget

import (
	"image/color"

	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type TermGrid struct {
	*widget.TextGrid
}

func NewTermGrid() *TermGrid {
	return &TermGrid{
		widget.NewTextGrid(),
	}
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

	t.ForRangeFn(blockMode, startRow, startCol, endRow, endCol, applyHighlight, nil)
}

// ClearHighlightRange disables the highlight style for the given range
func (t *TermGrid) ClearHighlightRange(blockMode bool, startRow, startCol, endRow, endCol int) {
	clearHighlight := func(cell *widget.TextGridCell) {
		// Check if already highlighted
		if h, ok := cell.Style.(*HighlightedTextGridStyle); ok {
			h.Highlighted = false
		}
	}
	t.ForRangeFn(blockMode, startRow, startCol, endRow, endCol, clearHighlight, nil)
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
