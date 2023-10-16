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

func (t *Terminal) getTermPosition(pos fyne.Position) position {
	cell := t.guessCellSize()
	col := int(pos.X/cell.Width) + 1
	row := int(pos.Y/cell.Height) + 1
	return position{col, row}
}
