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
	"fyne.io/fyne/v2/theme"
	"github.com/fyne-io/terminal"
	"github.com/fyne-io/terminal/cmd/fyneterm/data"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

const termOverlay = fyne.ThemeColorName("termOver")

var (
	localizer *i18n.Localizer
	sizer     *termTheme
)

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
	a.SetIcon(data.Icon)
	sizer = newTermTheme()
	a.Settings().SetTheme(sizer)
	w := newTerminalWindow(a, sizer, debug)
	w.ShowAndRun()
}

func newTerminalWindow(a fyne.App, th fyne.Theme, debug bool) fyne.Window {
	w := a.NewWindow(termTitle())
	w.SetPadded(false)

	bg := canvas.NewRectangle(theme.BackgroundColor())
	img := canvas.NewImageFromResource(data.FyneLogo)
	img.FillMode = canvas.ImageFillContain
	over := canvas.NewRectangle(th.Color(termOverlay, a.Settings().ThemeVariant()))

	ch := make(chan fyne.Settings)
	go func() {
		for {
			<-ch

			bg.FillColor = theme.BackgroundColor()
			bg.Refresh()
			over.FillColor = th.Color(termOverlay, a.Settings().ThemeVariant())
			over.Refresh()
		}
	}()
	a.Settings().AddChangeListener(ch)

	t := terminal.New()
	t.SetDebug(debug)
	setupListener(t, w)
	w.SetContent(container.NewStack(bg, img, over, t))

	cellSize := guessCellSize()
	w.Resize(fyne.NewSize(cellSize.Width*80, cellSize.Height*24))
	w.Canvas().Focus(t)

	t.AddShortcut(&desktop.CustomShortcut{KeyName: fyne.KeyN, Modifier: fyne.KeyModifierControl | fyne.KeyModifierShift},
		func(_ fyne.Shortcut) {
			w := newTerminalWindow(a, th, debug)
			w.Show()
		})
	t.AddShortcut(&desktop.CustomShortcut{KeyName: fyne.KeyEqual, Modifier: fyne.KeyModifierControl | fyne.KeyModifierShift},
		func(_ fyne.Shortcut) {
			sizer.fontSize++
			a.Settings().SetTheme(sizer)
			t.Refresh()
			t.Resize(t.Size())
		})
	t.AddShortcut(&desktop.CustomShortcut{KeyName: fyne.KeyMinus, Modifier: fyne.KeyModifierControl},
		func(_ fyne.Shortcut) {
			sizer.fontSize--
			a.Settings().SetTheme(sizer)
			t.Refresh()
			t.Resize(t.Size())
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
