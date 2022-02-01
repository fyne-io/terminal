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
func (t *termTheme) Color(n fyne.ThemeColorName, v fyne.ThemeVariant) color.Color {
	switch n {
	case termBackground:
		if v == theme.VariantLight {
			return &color.Gray{Y: 0xff}
		}
		return &color.NRGBA{R: 0x05, G: 0x08, B: 0x6b, A: 0xff}
	case termOverlay:
		if v == theme.VariantLight {
			return color.NRGBA{R: 0xaa, G: 0xaa, B: 0xaa, A: 0xf6}
		}
		return color.NRGBA{R: 0x16, G: 0x16, B: 0x16, A: 0xf6}
	}
	return t.Theme.Color(n, theme.VariantDark)
}

func (t *termTheme) Size(n fyne.ThemeSizeName) float32 {
	if n == theme.SizeNameText {
		return 12
	}

	return t.Theme.Size(n)
}
