package terminal

import (
	"image/color"

	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
)

var (
	cursorColor     = color.RGBA{R: 255, G: 255, B: 0, A: 128}
	cursorBellColor = color.RGBA{R: 255, G: 0, B: 0, A: 128}
)

type render struct {
	term   *Terminal
	cursor *canvas.Rectangle
}

func (r *render) Layout(s fyne.Size) {
	r.term.content.Resize(s)
}

func (r *render) MinSize() fyne.Size {
	return fyne.NewSize(0, 0) // don't get propped open by the text cells
}

func (r *render) Refresh() {
	r.moveCursor()
	r.refreshCursor()

	r.term.content.Refresh()
}

func (r *render) BackgroundColor() color.Color {
	return color.Transparent
}

func (r *render) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.term.content, r.cursor}
}

func (r *render) Destroy() {
}

func (r *render) moveCursor() {
	cell := r.term.guessCellSize()
	r.cursor.Move(fyne.NewPos(cell.Width*r.term.cursorCol, cell.Height*r.term.cursorRow))
}

func (r *render) refreshCursor() {
	r.cursor.Hidden = !r.term.focused
	if r.term.bell {
		r.cursor.FillColor = cursorBellColor
	} else {
		r.cursor.FillColor = cursorColor
	}
	r.cursor.Refresh()
}

// CreateRenderer requests a new renderer for this terminal (just a wrapper around the TextGrid)
func (t *Terminal) CreateRenderer() fyne.WidgetRenderer {
	cur := canvas.NewRectangle(cursorColor)
	cur.Resize(fyne.NewSize(2, t.guessCellSize().Height))

	r := &render{term: t, cursor: cur}
	t.cursorMoved = r.moveCursor
	return r
}
