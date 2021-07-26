package terminal

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
)

type render struct {
	term *Terminal
}

func (r *render) Layout(s fyne.Size) {
	r.term.content.Resize(s)
}

func (r *render) MinSize() fyne.Size {
	return fyne.NewSize(0, 0) // don't get propped open by the text cells
}

func (r *render) Refresh() {
	r.moveCursor()
	r.term.refreshCursor()

	r.term.content.Refresh()
}

func (r *render) BackgroundColor() color.Color {
	return color.Transparent
}

func (r *render) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.term.content, r.term.cursor}
}

func (r *render) Destroy() {
}

func (r *render) moveCursor() {
	cell := r.term.guessCellSize()
	r.term.cursor.Move(fyne.NewPos(cell.Width*float32(r.term.cursorCol), cell.Height*float32(r.term.cursorRow)))
}

func (t *Terminal) refreshCursor() {
	t.cursor.Hidden = !t.focused || t.cursorHidden
	if t.bell {
		t.cursor.FillColor = theme.ErrorColor()
	} else {
		t.cursor.FillColor = theme.PrimaryColor()
	}
	t.cursor.Refresh()
}

// CreateRenderer requests a new renderer for this terminal (just a wrapper around the TextGrid)
func (t *Terminal) CreateRenderer() fyne.WidgetRenderer {
	t.cursor = canvas.NewRectangle(theme.PrimaryColor())
	t.cursor.Hidden = true
	t.cursor.Resize(fyne.NewSize(2, t.guessCellSize().Height))

	r := &render{term: t}
	t.cursorMoved = r.moveCursor
	return r
}
