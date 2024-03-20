package widget

import (
	"fyne.io/fyne/v2"
)

var (
	m = map[fyne.CanvasObject]fyne.WidgetRenderer{}
)

func GetRenderer(w fyne.CanvasObject) fyne.WidgetRenderer {
	return m[w]
}

func RegisterRenderer(w fyne.CanvasObject, r fyne.WidgetRenderer) {
	m[w] = r
}
