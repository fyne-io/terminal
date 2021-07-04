package terminal

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
)

func (t *Terminal) handleMouseDownV200(btn int, mods desktop.Modifier, pos fyne.Position) {
	_, _ = t.Write(t.encodeMouse(btn, mods, pos))
}

func (t *Terminal) handleMouseDownX10(btn int, _ desktop.Modifier, pos fyne.Position) {
	_, _ = t.Write(t.encodeMouse(btn, 0, pos))
}

func (t *Terminal) handleMouseUpV200(btn int, mods desktop.Modifier, pos fyne.Position) {
	_, _ = t.Write(t.encodeMouse(0, mods, pos))
}

func (t *Terminal) handleMouseUpX10(_ int, _ desktop.Modifier, _ fyne.Position) {
	// no-op for X10 mode
}

func (t *Terminal) encodeMouse(button int, mods desktop.Modifier, pos fyne.Position) []byte {
	cell := t.guessCellSize()
	col := byte(pos.X/cell.Width) + 1
	row := byte(pos.Y/cell.Height) + 1
	var btn byte
	if button == 0 {
		btn = 3
	} else {
		btn = byte(button) - 1
	}

	if mods&desktop.ShiftModifier != 0 {
		btn += 4
	}
	if mods&desktop.AltModifier != 0 {
		btn += 8
	}
	if mods&desktop.ControlModifier != 0 {
		btn += 16
	}

	return []byte{asciiEscape, '[', 'M', 32 + btn, 32 + col, 32 + row}
}
