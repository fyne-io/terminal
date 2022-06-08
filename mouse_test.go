package terminal

import (
	"testing"

	"fyne.io/fyne/v2"
	"github.com/stretchr/testify/assert"
)

func TestEncodeMouse(t *testing.T) {
	term := New()
	assert.Equal(t, "\x1b[M !!", string(term.encodeMouse(1, 0, fyne.NewPos(4, 4))))
	assert.Equal(t, "\x1b[M!$#", string(term.encodeMouse(2, 0, fyne.NewPos(32, 36))))
	assert.Equal(t, "\x1b[M#!!", string(term.encodeMouse(0, 0, fyne.NewPos(4, 4))))
}

func TestEncodeMouse_Mods(t *testing.T) {
	term := New()
	assert.Equal(t, "\x1b[M$!!", string(term.encodeMouse(1,
		fyne.KeyModifierShift, fyne.NewPos(4, 4))))
	assert.Equal(t, "\x1b[M4!!", string(term.encodeMouse(1,
		fyne.KeyModifierShift|fyne.KeyModifierControl, fyne.NewPos(4, 4))))
	assert.Equal(t, "\x1b[M%!!", string(term.encodeMouse(2,
		fyne.KeyModifierShift, fyne.NewPos(4, 4))))
	assert.Equal(t, "\x1b[M5!!", string(term.encodeMouse(2,
		fyne.KeyModifierShift|fyne.KeyModifierControl, fyne.NewPos(4, 4))))
}
