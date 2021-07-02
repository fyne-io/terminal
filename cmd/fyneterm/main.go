//go:generate fyne bundle -o bundled.go Icon.png
//go:generate fyne bundle -o translation.go ../../translation/

package main

import (
	"encoding/json"
	"flag"
	"image/color"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"

	"github.com/fyne-io/terminal"
	"github.com/fyne-io/terminal/cmd/fyneterm/data"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

var localizer *i18n.Localizer

func setupListener(t *terminal.Terminal, w fyne.Window) {
	listen := make(chan terminal.Config)
	go func() {
		for {
			config := <-listen

			if config.Title == "" {
				w.SetTitle(termTitle())
			} else {
				w.SetTitle(termTitle() + ": " + config.Title)
			}
		}
	}()
	t.AddListener(listen)
}

func termTitle() string {
	return localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "Title",
			Other: "Fyne Terminal",
		},
	})
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

	bundle := i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)
	bundle.MustParseMessageFileBytes(resourceActiveFrJson.Content(), resourceActiveFrJson.Name())
	bundle.MustParseMessageFileBytes(resourceActiveRuJson.Content(), resourceActiveRuJson.Name())
	localizer = i18n.NewLocalizer(bundle, os.Getenv("LANG"))

	a := app.New()
	a.SetIcon(resourceIconPng)
	a.Settings().SetTheme(newTermTheme())

	w := newTerminalWindow(a, debug)
	w.ShowAndRun()
}

func newTerminalWindow(a fyne.App, debug bool) fyne.Window {
	w := a.NewWindow(termTitle())
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
