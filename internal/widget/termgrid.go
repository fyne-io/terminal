package widget

import (
	"context"
	"time"

	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"fyne.io/fyne/v2"
)

const blinkingInterval = 500 * time.Millisecond

// TermGrid is a monospaced grid of characters.
// This is designed to be used by our terminal emulator.
type TermGrid struct {
	widget.TextGrid

	// OnScrolled is called when the grid content moves, with the new offset of the content from the top left.
	OnScrolled func(fyne.Position)

	tickerCancel context.CancelFunc
	scroll       *container.Scroll
}

// CreateRenderer is a private method to Fyne which links this widget to it's renderer
func (t *TermGrid) CreateRenderer() fyne.WidgetRenderer {
	t.ExtendBaseWidget(t)

	r := t.TextGrid.CreateRenderer()
	if scroll, ok := r.Objects()[0].(*container.Scroll); ok {
		t.scroll = scroll
		t.hookScroll()
	}
	return r
}

// NewTermGrid creates a new empty TextGrid widget that scrolls vertically through its rows.
func NewTermGrid() *TermGrid {
	grid := &TermGrid{}
	grid.ExtendBaseWidget(grid)

	grid.Scroll = container.ScrollVerticalOnly
	return grid
}

// ScrollOffset returns how far the grid content has been scrolled from its top left, in pixels.
// Before the grid has been rendered there is no scroll state so a zero offset is returned.
func (t *TermGrid) ScrollOffset() fyne.Position {
	if t.scroll == nil {
		return fyne.Position{}
	}

	return t.scroll.Offset
}

// ScrollToBottom scrolls the grid so that the most recent rows are visible.
// It is safe to call before the grid has been rendered, in which case it does nothing.
func (t *TermGrid) ScrollToBottom() {
	if t.scroll == nil {
		return
	}

	t.TextGrid.ScrollToBottom()
}

// hookScroll wraps the scroll listener that TextGrid uses to recycle its rows so that we are told about scrolling too.
func (t *TermGrid) hookScroll() {
	// TextGrid installs its listener when the content is first rendered, so make sure that has happened
	t.scroll.Content.Refresh()

	inner := t.scroll.OnScrolled
	t.scroll.OnScrolled = func(offset fyne.Position) {
		if inner != nil {
			inner(offset)
		}
		if f := t.OnScrolled; f != nil {
			f(offset)
		}
	}
}

// Refresh will be called when this grid should update.
// We update our blinking status and then call the TextGrid we extended to refresh too.
func (t *TermGrid) Refresh() {
	t.refreshBlink(false)
}

func (t *TermGrid) refreshBlink(blink bool) {
	// reset shouldBlink which can be set by setCellRune if a cell with BlinkEnabled is found
	shouldBlink := false

	for _, row := range t.Rows {
		for _, r := range row.Cells {
			if s, ok := r.Style.(*TermTextGridStyle); ok && s != nil && s.BlinkEnabled {
				shouldBlink = true

				s.blink(blink)
			}
		}
	}
	fyne.Do(t.TextGrid.Refresh) // TODO fix root cause in refresh on wrong thread

	switch {
	case shouldBlink && t.tickerCancel == nil:
		t.runBlink()
	case !shouldBlink && t.tickerCancel != nil:
		t.tickerCancel()
		t.tickerCancel = nil
	}
}

func (t *TermGrid) runBlink() {
	if t.tickerCancel != nil {
		t.tickerCancel()
		t.tickerCancel = nil
	}
	var tickerContext context.Context
	tickerContext, t.tickerCancel = context.WithCancel(context.Background())
	ticker := time.NewTicker(blinkingInterval)
	blinking := false
	go func() {
		for {
			select {
			case <-tickerContext.Done():
				return
			case <-ticker.C:
				blinking = !blinking
				fyne.Do(func() {
					t.refreshBlink(blinking)
				})
			}
		}
	}()
}
