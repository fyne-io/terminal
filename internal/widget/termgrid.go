package widget

import (
	"time"

	"fyne.io/fyne/v2/widget"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

const (
	textAreaSpaceSymbol   = '·'
	textAreaTabSymbol     = '→'
	textAreaNewLineSymbol = '↵'
	blinkingInterval      = 500 * time.Millisecond
)

// TermGrid is a monospaced grid of characters.
// This is designed to be used by our terminal emulator.
type TermGrid struct {
	widget.TextGrid
}

// CreateRenderer is a private method to Fyne which links this widget to it's renderer
func (t *TermGrid) CreateRenderer() fyne.WidgetRenderer {
	t.ExtendBaseWidget(t)
	render := &termGridRenderer{text: t}
	render.updateCellSize()
	// N.B these global variables are not a good idea.
	widget.TextGridStyleDefault = &widget.CustomTextGridStyle{}
	widget.TextGridStyleWhitespace = &widget.CustomTextGridStyle{FGColor: theme.Color(theme.ColorNameDisabled)}

	return render
}

// NewTermGrid creates a new empty TextGrid widget.
func NewTermGrid() *TermGrid {
	grid := &TermGrid{}
	grid.ExtendBaseWidget(grid)
	return grid
}
