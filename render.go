package terminal

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	widget2 "github.com/fyne-io/terminal/internal/widget"
)

const cursorWidth = 2

// termContentLayout sizes the TextGrid to fill the container while
// reporting the TextGrid's full content MinSize for scroll bounds.
// The cursor is positioned manually and not affected by layout.
type termContentLayout struct {
	grid *widget2.TermGrid
}

func (l *termContentLayout) MinSize(_ []fyne.CanvasObject) fyne.Size {
	return l.grid.MinSize()
}

func (l *termContentLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	for _, o := range objects {
		if o == l.grid {
			o.Resize(size)
			o.Move(fyne.NewPos(0, 0))
		}
	}
}

type render struct {
	term *Terminal
}

func (r *render) Layout(s fyne.Size) {
	if r.term.scrollContainer != nil {
		r.term.scrollContainer.Resize(s)
	}
}

func (r *render) MinSize() fyne.Size {
	return fyne.NewSize(0, 0) // don't get propped open by the text cells
}

func (r *render) Refresh() {
	fyne.Do(func() { // TODO fix root cause in refresh on wrong thread
		r.moveCursor()
		r.term.refreshCursor()
	})

	r.term.content.Refresh()
}

func (r *render) BackgroundColor() color.Color {
	return color.Transparent
}

func (r *render) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.term.scrollContainer}
}

func (r *render) Destroy() {
}

func (r *render) moveCursor() {
	cell := r.term.guessCellSize()
	contentRow := r.term.rowOffset() + r.term.cursorRow

	// Position cursor at content-relative coordinates;
	// the scroll container handles making it scroll with the grid.
	r.term.cursor.Move(fyne.NewPos(cell.Width*float32(r.term.cursorCol), cell.Height*float32(contentRow)))
}

func (t *Terminal) refreshCursor() {
	t.cursor.Hidden = !t.focused || t.cursorHidden
	if t.bell {
		t.cursor.FillColor = theme.Color(theme.ColorNameError)
	} else {
		t.cursor.FillColor = theme.Color(theme.ColorNamePrimary)
	}
	t.cursor.Resize(fyne.NewSize(cursorWidth, t.guessCellSize().Height))
	t.cursor.Refresh()
}

// CreateRenderer requests a new renderer for this terminal (just a wrapper around the TextGrid)
func (t *Terminal) CreateRenderer() fyne.WidgetRenderer {
	t.ExtendBaseWidget(t)

	t.content = widget2.NewTermGrid()
	t.setupShortcuts()

	t.cursor = canvas.NewRectangle(theme.Color(theme.ColorNamePrimary))
	t.cursor.Hidden = true
	t.cursor.Resize(fyne.NewSize(cursorWidth, t.guessCellSize().Height))

	inner := container.New(&termContentLayout{grid: t.content}, t.content, t.cursor)
	t.scrollContainer = container.NewVScroll(inner)

	r := &render{term: t}
	t.cursorMoved = r.moveCursor
	return r
}
