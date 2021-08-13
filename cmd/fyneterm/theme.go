package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type termTheme struct {
	fyne.Theme
}

func newTermTheme() fyne.Theme {
	return &termTheme{theme.DefaultTheme()}
}

// Color fixes a bug < 2.1 where theme.DarkTheme() would not override user preference.
func (t *termTheme) Color(n fyne.ThemeColorName, _ fyne.ThemeVariant) color.Color {
	return t.Theme.Color(n, theme.VariantDark)
}

func (t *termTheme) Size(n fyne.ThemeSizeName) float32 {
	if n == theme.SizeNameText {
		return 12
	}

	return t.Theme.Size(n)
}
