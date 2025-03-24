package main

import (
	"embed"
	"flag"
	"image/color"
	"runtime"

	"github.com/fyne-io/terminal"
	"github.com/fyne-io/terminal/cmd/fyneterm/data"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/theme"
)

const termOverlay = fyne.ThemeColorName("termOver")

//go:embed translation
var translations embed.FS

func setupListener(t *terminal.Terminal, w fyne.Window) {
	listen := make(chan terminal.Config)
	go func() {
		for {
			config := <-listen

			fyne.Do(func() {
				if config.Title == "" {
					w.SetTitle(termTitle())
				} else {
					w.SetTitle(termTitle() + ": " + config.Title)
				}
			})
		}
	}()
	t.AddListener(listen)
}

func termTitle() string {
	return lang.L("Title")
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

	lang.AddTranslationsFS(translations, "translation")

	a := app.New()
	a.SetIcon(data.Icon)
	w := newTerminalWindow(a, debug)
	w.ShowAndRun()
}

func newTerminalWindow(a fyne.App, debug bool) fyne.Window {
	w := a.NewWindow(termTitle())
	w.SetPadded(false)
	th := newTermTheme()

	bg := canvas.NewRectangle(theme.Color(theme.ColorNameBackground))
	img := canvas.NewImageFromResource(data.FyneLogo)
	img.FillMode = canvas.ImageFillContain
	over := canvas.NewRectangle(th.Color(termOverlay, a.Settings().ThemeVariant()))

	a.Settings().AddListener(func(s fyne.Settings) {
		bg.FillColor = theme.Color(theme.ColorNameBackground)
		bg.Refresh()
		over.FillColor = th.Color(termOverlay, s.ThemeVariant())
		over.Refresh()
	})

	t := terminal.New()
	t.SetDebug(debug)
	setupListener(t, w)
	sizeOverride := container.NewThemeOverride(container.NewStack(bg, img, over, t), th)
	w.SetContent(sizeOverride)

	cellSize := guessCellSize()
	w.Resize(fyne.NewSize(cellSize.Width*80, cellSize.Height*24))
	w.Canvas().Focus(t)

	newTerm := func(_ fyne.Shortcut) {
		w := newTerminalWindow(a, debug)
		w.Show()
	}
	t.AddShortcut(&desktop.CustomShortcut{KeyName: fyne.KeyN, Modifier: fyne.KeyModifierControl | fyne.KeyModifierShift}, newTerm)
	if runtime.GOOS == "darwin" {
		t.AddShortcut(&desktop.CustomShortcut{KeyName: fyne.KeyN, Modifier: fyne.KeyModifierSuper}, newTerm)
	}
	t.AddShortcut(&desktop.CustomShortcut{KeyName: fyne.KeyEqual, Modifier: fyne.KeyModifierShortcutDefault | fyne.KeyModifierShift},
		func(_ fyne.Shortcut) {
			th.fontSize++
			sizeOverride.Theme = th
			sizeOverride.Refresh()
		})
	t.AddShortcut(&desktop.CustomShortcut{KeyName: fyne.KeyMinus, Modifier: fyne.KeyModifierShortcutDefault},
		func(_ fyne.Shortcut) {
			th.fontSize--
			sizeOverride.Theme = th
			sizeOverride.Refresh()
		})

	go func() {
		err := t.RunLocalShell()
		if err != nil {
			fyne.LogError("Failure in terminal", err)
		}
		fyne.Do(w.Close)
	}()

	return w
}
