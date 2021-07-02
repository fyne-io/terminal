//go:generate fyne bundle -o bundled.go Icon.png

package main

import (
	"flag"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"

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
	var debug bool
	flag.BoolVar(&debug, "debug", false, "Show terminal debug messages")
	flag.Parse()

	a := app.New()
	a.SetIcon(resourceIconPng)
	a.Settings().SetTheme(newTermTheme())

	w := newTerminalWindow(a, debug)
	w.ShowAndRun()
}

func newTerminalWindow(a fyne.App, debug bool) fyne.Window {
	w := a.NewWindow(termTitle)
	w.SetPadded(false)

	bg := canvas.NewRectangle(color.Gray{Y: 0x16})
	img := canvas.NewImageFromResource(data.FyneScene)
	img.FillMode = canvas.ImageFillContain
	img.Translucency = 0.95

	t := terminal.New()
	t.SetDebug(debug)
	setupListener(t, w)
	w.SetContent(container.NewMax(bg, img, t))

	cellSize := guessCellSize()
	w.Resize(fyne.NewSize(cellSize.Width*80, cellSize.Height*24))
	w.Canvas().Focus(t)

	t.AddShortcut(&desktop.CustomShortcut{KeyName: fyne.KeyN, Modifier: desktop.ControlModifier | desktop.ShiftModifier},
		func(_ fyne.Shortcut) {
			w := newTerminalWindow(a, debug)
			w.Show()
		})
	go func() {
		err := t.RunLocalShell()
		if err != nil {
			fyne.LogError("Failure in terminal", err)
		}
		w.Close()
	}()

	return w
}
