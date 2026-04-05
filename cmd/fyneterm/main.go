package main

import (
	"embed"
	"flag"
	"image/color"
	"math"
	"os"
	"path/filepath"
	"runtime"

	"github.com/fyne-io/terminal"
	"github.com/fyne-io/terminal/cmd/fyneterm/data"
	"github.com/fyshos/fancyfs"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
)

//go:embed translation
var translations embed.FS

func termTitle() string {
	return lang.L("Title")
}

func guessCellSize(th *termTheme) fyne.Size {
	cell := canvas.NewText("M", color.White)
	cell.TextStyle.Monospace = true

	min := cell.MinSize()
	scale := th.fontSize / theme.TextSize()
	return fyne.NewSize(float32(math.Round(float64(min.Width*scale))), float32(math.Round(float64(min.Height*scale))))
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

func findTerminal(item *container.TabItem) *terminal.Terminal {
	override := item.Content.(*container.ThemeOverride)
	stack := override.Content.(*fyne.Container)
	for _, obj := range stack.Objects {
		if t, ok := obj.(*terminal.Terminal); ok {
			return t
		}
	}
	return nil
}

func newTerminalWindow(a fyne.App, debug bool) fyne.Window {
	w := a.NewWindow(termTitle())
	w.SetPadded(false)
	th := newTermTheme()

	var wrapper *fyne.Container
	tabs := container.NewDocTabs()

	// updateView swaps between showing DocTabs (2+ tabs) or the
	// single tab's content directly (1 tab, no tab bar visible).
	// It adjusts the window height to compensate for the tab bar.
	updateView := func() {
		barHeight := tabs.MinSize().Height - tabs.Selected().Content.MinSize().Height
		size := w.Canvas().Size()
		if len(tabs.Items) == 1 {
			wrapper.Objects = []fyne.CanvasObject{tabs.Items[0].Content}
			w.Resize(fyne.NewSize(size.Width, size.Height-barHeight))
		} else {
			wrapper.Objects = []fyne.CanvasObject{tabs}
			w.Resize(fyne.NewSize(size.Width, size.Height+barHeight))
		}
		wrapper.Refresh()
	}

	firstTab := newTab(tabs, updateView, debug, th, w, a)
	tabs.Append(firstTab)
	tabs.CreateTab = func() *container.TabItem {
		tab := newTab(tabs, updateView, debug, th, w, a)
		// updateView after DocTabs finishes appending the new tab
		defer func() {
			updateView()
			w.Canvas().Focus(findTerminal(tab))
		}()
		return tab
	}

	tabs.OnSelected = func(item *container.TabItem) {
		if item.Text == "" || item.Text == termTitle() {
			w.SetTitle(termTitle())
		} else {
			w.SetTitle(termTitle() + ": " + item.Text)
		}
		w.Canvas().Focus(findTerminal(item))
	}

	// Start in single-tab mode (no tab bar visible)
	wrapper = container.NewStack(firstTab.Content)
	w.SetContent(wrapper)

	cellSize := guessCellSize(th)
	w.Resize(fyne.NewSize(cellSize.Width*80, cellSize.Height*24))

	w.Canvas().Focus(findTerminal(firstTab))

	return w
}

func newTab(tabs *container.DocTabs, refresh func(), debug bool, th *termTheme, w fyne.Window, a fyne.App) *container.TabItem {
	bg := canvas.NewRectangle(theme.Color(theme.ColorNameBackground))
	img := canvas.NewImageFromResource(data.FyneLogo)
	img.FillMode = canvas.ImageFillContain
	img.Translucency = 0.8

	setDir := func(pwd string) {
		ff, err := fancyfs.DetailsForFolder(storage.NewFileURI(pwd))
		if ff == nil || err != nil {
			if err != nil && err != fancyfs.ErrNoMetadata {
				fyne.LogError("Could not read dir metadata", err)
			}

			img.File = ""
			img.Resource = data.FyneLogo
			img.Image = nil
			img.FillMode = canvas.ImageFillContain
			img.Refresh()
			return
		}

		if ff.BackgroundURI != nil {
			img.File = ff.BackgroundURI.Path()
		} else {
			img.File = ""
		}

		homeStr, _ := os.UserHomeDir()
		if pwd == "" || pwd == homeStr {
			img.Resource = data.FyneLogo
		} else {
			img.Resource = ff.BackgroundResource
		}
		img.FillMode = ff.BackgroundFill
		img.Refresh()
	}

	t := terminal.New()
	t.SetDebug(debug)
	if len(os.Args) >= 2 {
		s, err := filepath.Abs(os.Args[1])
		if err == nil {
			t.SetStartDir(s)
			setDir(s)
		}
	} else {
		wd, err := os.Getwd()
		if err == nil {
			setDir(wd)
		}
	}

	a.Settings().AddListener(func(s fyne.Settings) {
		bg.FillColor = theme.Color(theme.ColorNameBackground)
		bg.Refresh()
	})

	sizeOverride := container.NewThemeOverride(container.NewStack(bg, img, t), th)
	tabItem := container.NewTabItem(termTitle(), sizeOverride)

	listen := make(chan terminal.Config)
	go func() {
		for config := range listen {
			fyne.Do(func() {
				title := config.Title
				if title == "" {
					tabItem.Text = termTitle()
				} else {
					tabItem.Text = title
				}
				if len(tabs.Items) > 1 {
					tabs.Refresh()
				}

				if tabs.Selected() == tabItem {
					if title == "" {
						w.SetTitle(termTitle())
					} else {
						w.SetTitle(termTitle() + ": " + title)
					}
				}

				setDir(config.PWD)
			})
		}
	}()
	t.AddListener(listen)

	// New window shortcut
	newWin := func(_ fyne.Shortcut) {
		w := newTerminalWindow(a, debug)
		w.Show()
	}
	t.AddShortcut(&desktop.CustomShortcut{KeyName: fyne.KeyN, Modifier: fyne.KeyModifierControl | fyne.KeyModifierShift}, newWin)
	if runtime.GOOS == "darwin" {
		t.AddShortcut(&desktop.CustomShortcut{KeyName: fyne.KeyN, Modifier: fyne.KeyModifierSuper}, newWin)
	}

	// New tab shortcut
	newTabShortcut := func(_ fyne.Shortcut) {
		item := newTab(tabs, refresh, debug, th, w, a)
		tabs.Append(item)
		tabs.Select(item)
		refresh()
		w.Canvas().Focus(findTerminal(item))
	}
	t.AddShortcut(&desktop.CustomShortcut{KeyName: fyne.KeyT, Modifier: fyne.KeyModifierControl | fyne.KeyModifierShift}, newTabShortcut)
	if runtime.GOOS == "darwin" {
		t.AddShortcut(&desktop.CustomShortcut{KeyName: fyne.KeyT, Modifier: fyne.KeyModifierSuper}, newTabShortcut)
	}

	// Font size shortcuts
	refreshAllTabs := func() {
		for _, item := range tabs.Items {
			item.Content.Refresh()
		}
	}
	t.AddShortcut(&desktop.CustomShortcut{KeyName: fyne.KeyEqual, Modifier: fyne.KeyModifierShortcutDefault | fyne.KeyModifierShift},
		func(_ fyne.Shortcut) {
			th.fontSize++
			refreshAllTabs()
		})
	t.AddShortcut(&desktop.CustomShortcut{KeyName: fyne.KeyMinus, Modifier: fyne.KeyModifierShortcutDefault},
		func(_ fyne.Shortcut) {
			th.fontSize--
			refreshAllTabs()
		})

	go func() {
		err := t.RunLocalShell()
		if err != nil {
			fyne.LogError("Failure in terminal", err)
		}
		fyne.Do(func() {
			tabs.Remove(tabItem)
			if len(tabs.Items) == 0 {
				w.Close()
				return
			}
			refresh()
			w.Canvas().Focus(findTerminal(tabs.Selected()))
		})
	}()

	return tabItem
}
