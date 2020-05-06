package main

import (
	"image/color"

	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/driver/desktop"
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
	w := a.NewWindow(termTitle)
	w.SetPadded(false)

	bg := canvas.NewRectangle(&color.RGBA{8, 8, 8, 255})
	img := canvas.NewImageFromResource(data.FyneScene)
	img.FillMode = canvas.ImageFillContain
	img.Translucency = 0.85

	t := terminal.NewTerminal()
	setupListener(t, w)
	w.SetContent(fyne.NewContainerWithLayout(layout.NewMaxLayout(), bg, img, t))
	w.Canvas().AddShortcut(&desktop.CustomShortcut{
		KeyName:  fyne.KeyD,
		Modifier: desktop.ControlModifier,
	}, func(_ fyne.Shortcut) {
		t.Exit()
	})

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
