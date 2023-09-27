package terminal

import (
	"fyne.io/fyne/v2"
	widget2 "github.com/fyne-io/terminal/internal/widget"
)

// getSelectedRange returns the current selection range, start row, start col, end row, end col
// It always returns a positive selection
func (t *Terminal) getSelectedRange() (int, int, int, int) {
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

func (t *Terminal) highlightSelectedText() {
	sr, sc, er, ec := t.getSelectedRange()
	tg := *t.content
	tg2 := widget2.TermGrid(tg)
	tg2.HighlightRange(t.blockMode, sr, sc, er, ec, t.highlightBitMask)
	t.Refresh()
}

func (t *Terminal) clearSelectedText() {
	sr, sc, er, ec := t.getSelectedRange()
	tg := *t.content
	tg2 := widget2.TermGrid(tg)
	tg2.ClearHighlightRange(t.blockMode, sr, sc, er, ec)
	t.Refresh()
	t.blockMode = false
	t.selecting = false
}

// SelectedText gets the text that is currently selected.
func (t *Terminal) SelectedText() string {
	sr, sc, er, ec := t.getSelectedRange()
	tg := *t.content
	tg2 := widget2.TermGrid(tg)
	return tg2.GetTextRange(t.blockMode, sr, sc, er, ec)
}

func (t *Terminal) copySelectedText(clipboard fyne.Clipboard) {
	// copy start and end sel to clipboard and clear the sel style
	text := t.SelectedText()
	fyne.CurrentApp()
	clipboard.SetContent(text)
	t.clearSelectedText()
}

func (t *Terminal) pasteText(clipboard fyne.Clipboard) {
	content := clipboard.Content()
	_, _ = t.in.Write([]byte(content))
}

func (t *Terminal) hasSelectedText() bool {
	return t.selStart != nil && t.selEnd != nil
}
