//go:generate fyne bundle -o bundled.go Icon.png
//go:generate fyne bundle -o translation.go ../../translation/

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"image/color"
	"os"
	"time"

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

var (
	nextTab   = 1
	localizer *i18n.Localizer
)

func setupListener(t *terminal.Terminal, tab *container.TabItem, tabs *container.DocTabs, w fyne.Window) {
	listen := make(chan terminal.Config)
	go func() {
		for {
			config := <-listen

			if config.Title == "" {
				w.SetTitle(termTitle())
			} else {
				w.SetTitle(termTitle() + ": " + config.Title)
				tab.Text = config.Title
				tabs.Refresh()
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

func newTerm(debug bool, tabs *container.DocTabs, item *container.TabItem, w fyne.Window, a fyne.App) *fyne.Container {
	bg := canvas.NewRectangle(color.Gray{Y: 0x16})
	img := canvas.NewImageFromResource(data.FyneScene)
	img.FillMode = canvas.ImageFillContain
	img.Translucency = 0.95

	t := terminal.New()
	t.SetDebug(debug)
	setupListener(t, item, tabs, w)

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

		if len(tabs.Items) == 1 {
			w.Close()
		} else {
			tabs.Remove(tabs.Selected())
		}
	}()

	return container.NewMax(bg, img, t)
}

func newTerminalWindow(a fyne.App, debug bool) fyne.Window {
	w := a.NewWindow(termTitle())
	w.SetPadded(false)

	tabs := container.NewDocTabs()
	newTab := func() *container.TabItem {
		item := container.NewTabItemWithIcon(fmt.Sprintf("Tab %d", nextTab), resourceIconPng, nil)
		item.Content = newTerm(debug, tabs, item, w, a)

		nextTab++
		return item
	}
	tabs.Append(newTab())

	tabs.CreateTab = newTab
	tabs.OnSelected = func(tab *container.TabItem) {
		term := tab.Content.(*fyne.Container).Objects[2].(*terminal.Terminal)
		go func() {
			time.Sleep(time.Millisecond * 50)
			w.Canvas().Focus(term)
		}()
	}
	w.SetContent(tabs)

	cellSize := guessCellSize()
	w.Resize(fyne.NewSize(cellSize.Width*80, cellSize.Height*24))
	w.Canvas().Focus(tabs.Items[0].Content.(*fyne.Container).Objects[2].(*terminal.Terminal))

	return w
}
