package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type termTheme struct {
	fyne.Theme

	fontSize float32
}

func newTermTheme() *termTheme {
	return &termTheme{Theme: fyne.CurrentApp().Settings().Theme(), fontSize: 12}
}

// Color fixes a bug < 2.1 where theme.DarkTheme() would not override user preference.
func (t *termTheme) Color(n fyne.ThemeColorName, v fyne.ThemeVariant) color.Color {
	switch n {
	case theme.ColorNameBackground, theme.ColorNameForeground:
		return t.Theme.Color(n, v)
	}
	return t.Theme.Color(n, theme.VariantDark)
}

func (t *termTheme) Size(n fyne.ThemeSizeName) float32 {
	if n == theme.SizeNameText {
		return t.fontSize
	}

	return t.Theme.Size(n)
}
