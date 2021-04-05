//go:generate fyne bundle -o bundled.go Icon.png

package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/layout"
	"github.com/fyne-io/terminal"
	"github.com/fyne-io/terminal/cmd/fyneterm/data"
)

const (
	termTitle = "Fyne Terminal"
)

func setupListener(t *terminal.Terminal, w fyne.Window) {
	listen := make(chan terminal.Config)
	go func() {
		for {
			config := <-listen

			if config.Title == "" {
				w.SetTitle(termTitle)
			} else {
				w.SetTitle(termTitle + ": " + config.Title)
			}
		}
	}()
	t.AddListener(listen)
}

func guessCellSize() fyne.Size {
	cell := canvas.NewText("M", color.White)
	cell.TextStyle.Monospace = true

	return cell.MinSize()
}

func main() {
	a := app.New()
	a.SetIcon(resourceIconPng)
	w := a.NewWindow(termTitle)
	w.SetPadded(false)

	bg := canvas.NewRectangle(color.Gray{Y: 0x16})
	img := canvas.NewImageFromResource(data.FyneScene)
	img.FillMode = canvas.ImageFillContain
	img.Translucency = 0.95

	t := terminal.NewTerminal()
	setupListener(t, w)
	w.SetContent(fyne.NewContainerWithLayout(layout.NewMaxLayout(), bg, img, t))

	cellSize := guessCellSize()
	w.Resize(fyne.NewSize(cellSize.Width*80, cellSize.Height*24))
	w.Canvas().Focus(t)

	go func() {
		err := t.Run()
		if err != nil {
			fyne.LogError("Failure in terminal", err)
		}
		a.Quit()
	}()
	w.ShowAndRun()
}
