package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type termTheme struct {
	fyne.Theme
}

func newTermTheme() fyne.Theme {
	return &termTheme{theme.DefaultTheme()}
}

func (t *termTheme) Size(n fyne.ThemeSizeName) float32 {
	if n == theme.SizeNameText {
		return 12
	}

	return t.Theme.Size(n)
}
