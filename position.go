package terminal

import (
	"fmt"

	"fyne.io/fyne/v2"
)

type Position struct {
	Col, Row int
}

func (r Position) String() string {
	return fmt.Sprintf("row: %x, col: %x", r.Row, r.Col)
}

func (t *Terminal) GetTermPosition(pos fyne.Position) Position {
	cell := t.guessCellSize()
	col := int(pos.X/cell.Width) + 1
	row := int(pos.Y/cell.Height) + 1
	return Position{col, row}
}
