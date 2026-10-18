package terminal

import (
	"fmt"

	"fyne.io/fyne/v2"
)

type position struct {
	Col, Row int
}

func (r position) String() string {
	return fmt.Sprintf("row: %x, col: %x", r.Row, r.Col)
}

// getTermPosition returns the position within the terminal content (including scrollback)
// for the given widget-relative coordinates, taking any scroll offset into account.
func (t *Terminal) getTermPosition(pos fyne.Position) position {
	if t.content != nil {
		pos = pos.Add(t.content.ScrollOffset())
	}
	return t.getScreenPosition(pos)
}

// getScreenPosition returns the screen-relative position (ignoring scroll offset).
// Used for mouse reporting to terminal applications.
func (t *Terminal) getScreenPosition(pos fyne.Position) position {
	cell := t.guessCellSize()
	col := int(pos.X/cell.Width) + 1
	row := int(pos.Y/cell.Height) + 1
	return position{col, row}
}

// getTextPosition converts a terminal position (row and col) to fyne coordinates.
func (t *Terminal) getTextPosition(pos position) fyne.Position {
	cell := t.guessCellSize()
	x := (pos.Col - 1) * int(cell.Width)  // Convert column to pixel position (1-based to 0-based)
	y := (pos.Row - 1) * int(cell.Height) // Convert row to pixel position (1-based to 0-based)
	return fyne.NewPos(float32(x), float32(y))
}
