package widget

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"github.com/stretchr/testify/assert"
)

func TestTermGrid_OnScrolled(t *testing.T) {
	grid := NewTermGrid()
	grid.SetText("1\n2\n3\n4\n5\n6\n7\n8\n9\n10")
	w := test.NewTempWindow(t, grid)
	w.Resize(fyne.NewSize(100, 40))

	var scrolled []fyne.Position
	grid.OnScrolled = func(offset fyne.Position) {
		scrolled = append(scrolled, offset)
	}
	assert.Zero(t, grid.ScrollOffset())

	grid.ScrollToBottom()
	assert.Len(t, scrolled, 1)
	assert.Greater(t, scrolled[0].Y, float32(0))
	assert.Equal(t, scrolled[0], grid.ScrollOffset())

	grid.ScrollToTop()
	assert.Len(t, scrolled, 2)
	assert.Zero(t, scrolled[1].Y)
	assert.Zero(t, grid.ScrollOffset())
}
