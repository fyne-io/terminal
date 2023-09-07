package terminal

import (
	"fyne.io/fyne/v2"
	widget2 "github.com/fyne-io/terminal/widget"
)

// GetSelectedRange returns the current selection range, start row, start col, end row, end col
// It always returns a positive selection
func (t *Terminal) GetSelectedRange() (int, int, int, int) {
	if t.selStart == nil || t.selEnd == nil {
		return 0, 0, 0, 0
	}

	startRow := t.selStart.Row
	startCol := t.selStart.Col
	endRow := t.selEnd.Row
	endCol := t.selEnd.Col

	// Check if the user has selected in reverse
	if startRow > endRow || (startRow == endRow && startCol > endCol) {
		// Swap the start and end rows and columns
		startRow, endRow = endRow, startRow
		startCol, endCol = endCol, startCol
	}

	return startRow - 1, startCol - 1, endRow - 1, endCol - 1
}

func (t *Terminal) HighlightSelectedText() {
	sr, sc, er, ec := t.GetSelectedRange()
	t.content.HighlightRange(t.blockMode, sr, sc, er, ec, widget2.WithInvert(t.highlightBitMask))
	t.Refresh()
}

func (t *Terminal) ClearSelectedText() {

	sr, sc, er, ec := t.GetSelectedRange()
	t.content.ClearHighlightRange(t.blockMode, sr, sc, er, ec)
	t.Refresh()
	t.blockMode = false
	t.selecting = false
}

func (t *Terminal) SelectedText() string {
	sr, sc, er, ec := t.GetSelectedRange()
	return t.content.GetTextRange(t.blockMode, sr, sc, er, ec)
}

func (t *Terminal) CopySelectedText(clipboard fyne.Clipboard) {
	// copy start and end sel to clipboard and clear the sel style
	text := t.SelectedText()
	fyne.CurrentApp()
	clipboard.SetContent(text)
	t.ClearSelectedText()
}

func (t *Terminal) PasteText(clipboard fyne.Clipboard) {
	content := clipboard.Content()
	_, _ = t.in.Write([]byte(content))
}

func (t *Terminal) HasSelectedText() bool {
	return t.selStart != nil && t.selEnd != nil
}
