package terminal

import (
	"image/color"

	"fyne.io/fyne"
)

type render struct {
	term *Terminal
}

func (r *render) Layout(s fyne.Size) {
	r.term.content.Resize(s)
}

func (r *render) MinSize() fyne.Size {
	return r.term.content.MinSize()
}

func (r *render) Refresh() {
	r.term.content.Refresh()
}

func (r *render) BackgroundColor() color.Color {
	return color.Transparent
}

func (r *render) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.term.content}
}

func (r *render) Destroy() {
}

// CreateRenderer requests a new renderer for this terminal (just a wrapper around the TextGrid)
func (t *Terminal) CreateRenderer() fyne.WidgetRenderer {
	return &render{term: t}
}
