package main

import (
	"image/color"

	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/layout"

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

			w.SetTitle(termTitle + ": " + config.Title)
		}
	}()
	t.AddListener(listen)
}

func main() {
	a := app.New()
	w := a.NewWindow(termTitle)
	w.SetPadded(false)

	bg := canvas.NewRectangle(&color.RGBA{8, 8, 8, 255})
	img := canvas.NewImageFromResource(data.FyneScene)
	img.FillMode = canvas.ImageFillContain
	img.Translucency = 0.85

	t := terminal.NewTerminal()
	setupListener(t, w)
	w.SetContent(fyne.NewContainerWithLayout(layout.NewMaxLayout(), bg, img, t))
	w.Resize(fyne.NewSize(420, 260))
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
