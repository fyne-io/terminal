package widget

import (
	"context"
	"image/color"
	"math"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type termGridRenderer struct {
	text *TermGrid

	cols, rows int

	cellSize     fyne.Size
	objects      []fyne.CanvasObject
	current      fyne.Canvas
	blink        bool
	shouldBlink  bool
	tickerCancel context.CancelFunc
}

func (t *termGridRenderer) appendTextCell(str rune) {
	text := canvas.NewText(string(str), theme.ForegroundColor())
	text.TextStyle.Monospace = true

	bg := canvas.NewRectangle(color.Transparent)
	t.objects = append(t.objects, bg, text)
}

func (t *termGridRenderer) setCellRune(str rune, pos int, style widget.TextGridStyle) {
	if str == 0 {
		str = ' '
	}
	fg := theme.ForegroundColor()
	if style != nil && style.TextColor() != nil {
		fg = style.TextColor()
	}
	bg := color.Color(color.Transparent)
	if style != nil && style.BackgroundColor() != nil {
		bg = style.BackgroundColor()
	}

	if s, ok := style.(*TermTextGridStyle); ok && s != nil && s.BlinkEnabled {
		t.shouldBlink = true
		if t.blink {
			fg = bg
		}
	}

	text := t.objects[pos*2+1].(*canvas.Text)
	text.TextSize = theme.TextSize()

	newStr := string(str)
	if text.Text != newStr || text.Color != fg {
		text.Text = newStr
		text.Color = fg
		t.refresh(text)
	}

	rect := t.objects[pos*2].(*canvas.Rectangle)
	if rect.FillColor != bg {
		rect.FillColor = bg
		t.refresh(rect)
	}
}

func (t *termGridRenderer) addCellsIfRequired() {
	cellCount := t.cols * t.rows
	if len(t.objects) == cellCount*2 {
		return
	}
	for i := len(t.objects); i < cellCount*2; i += 2 {
		t.appendTextCell(' ')
	}
}

func (t *termGridRenderer) refreshGrid() {
	line := 1
	x := 0
	// reset shouldBlink which can be set by setCellRune if a cell with BlinkEnabled is found
	t.shouldBlink = false

	for rowIndex, row := range t.text.Rows {
		i := 0
		if t.text.ShowLineNumbers {
			lineStr := []rune(strconv.Itoa(line))
			pad := t.lineNumberWidth() - len(lineStr)
			for ; i < pad; i++ {
				t.setCellRune(' ', x, widget.TextGridStyleWhitespace) // padding space
				x++
			}
			for c := 0; c < len(lineStr); c++ {
				t.setCellRune(lineStr[c], x, widget.TextGridStyleDefault) // line numbers
				i++
				x++
			}

			t.setCellRune('|', x, widget.TextGridStyleWhitespace) // last space
			i++
			x++
		}
		for _, r := range row.Cells {
			if i >= t.cols { // would be an overflow - bad
				continue
			}
			if t.text.ShowWhitespace && (r.Rune == ' ' || r.Rune == '\t') {
				sym := textAreaSpaceSymbol
				if r.Rune == '\t' {
					sym = textAreaTabSymbol
				}

				if r.Style != nil && r.Style.BackgroundColor() != nil {
					whitespaceBG := &widget.CustomTextGridStyle{FGColor: widget.TextGridStyleWhitespace.TextColor(),
						BGColor: r.Style.BackgroundColor()}
					t.setCellRune(sym, x, whitespaceBG) // whitespace char
				} else {
					t.setCellRune(sym, x, widget.TextGridStyleWhitespace) // whitespace char
				}
			} else {
				t.setCellRune(r.Rune, x, r.Style) // regular char
			}
			i++
			x++
		}
		if t.text.ShowWhitespace && i < t.cols && rowIndex < len(t.text.Rows)-1 {
			t.setCellRune(textAreaNewLineSymbol, x, widget.TextGridStyleWhitespace) // newline
			i++
			x++
		}
		for ; i < t.cols; i++ {
			t.setCellRune(' ', x, widget.TextGridStyleDefault) // blanks
			x++
		}

		line++
	}
	for ; x < len(t.objects)/2; x++ {
		t.setCellRune(' ', x, widget.TextGridStyleDefault) // trailing cells and blank lines
	}

	switch {
	case t.shouldBlink && t.tickerCancel == nil:
		t.runBlink()
	case !t.shouldBlink && t.tickerCancel != nil:
		t.tickerCancel()
		t.tickerCancel = nil
	}
}

func (t *termGridRenderer) runBlink() {
	if t.tickerCancel != nil {
		t.tickerCancel()
		t.tickerCancel = nil
	}
	var tickerContext context.Context
	tickerContext, t.tickerCancel = context.WithCancel(context.Background())
	ticker := time.NewTicker(blinkingInterval)
	blinking := false
	go func() {
		for {
			select {
			case <-tickerContext.Done():
				return
			case <-ticker.C:
				t.SetBlink(blinking)
				blinking = !blinking
				t.refreshGrid()
			}
		}
	}()
}

func (t *termGridRenderer) lineNumberWidth() int {
	return len(strconv.Itoa(t.rows + 1))
}

func (t *termGridRenderer) updateGridSize(size fyne.Size) {
	bufRows := len(t.text.Rows)
	bufCols := 0
	for _, row := range t.text.Rows {
		bufCols = int(math.Max(float64(bufCols), float64(len(row.Cells))))
	}
	sizeCols := math.Floor(float64(size.Width) / float64(t.cellSize.Width))
	sizeRows := math.Floor(float64(size.Height) / float64(t.cellSize.Height))

	if t.text.ShowWhitespace {
		bufCols++
	}
	if t.text.ShowLineNumbers {
		bufCols += t.lineNumberWidth()
	}

	t.cols = int(math.Max(sizeCols, float64(bufCols)))
	t.rows = int(math.Max(sizeRows, float64(bufRows)))
	t.addCellsIfRequired()
}

func (t *termGridRenderer) Layout(size fyne.Size) {
	t.updateGridSize(size)

	i := 0
	cellPos := fyne.NewPos(0, 0)
	for y := 0; y < t.rows; y++ {
		for x := 0; x < t.cols; x++ {
			t.objects[i*2+1].Move(cellPos)

			t.objects[i*2].Resize(t.cellSize)
			t.objects[i*2].Move(cellPos)
			cellPos.X += t.cellSize.Width
			i++
		}

		cellPos.X = 0
		cellPos.Y += t.cellSize.Height
	}
}

func (t *termGridRenderer) MinSize() fyne.Size {
	longestRow := float32(0)
	for _, row := range t.text.Rows {
		longestRow = fyne.Max(longestRow, float32(len(row.Cells)))
	}
	return fyne.NewSize(t.cellSize.Width*longestRow,
		t.cellSize.Height*float32(len(t.text.Rows)))
}

func (t *termGridRenderer) Refresh() {
	// we may be on a new canvas, so just update it to be sure
	if fyne.CurrentApp() != nil && fyne.CurrentApp().Driver() != nil {
		t.current = fyne.CurrentApp().Driver().CanvasForObject(t.text)
	}

	// theme could change text size
	t.updateCellSize()

	widget.TextGridStyleWhitespace = &widget.CustomTextGridStyle{FGColor: theme.DisabledColor()}
	t.updateGridSize(t.text.Size())
	t.refreshGrid()
}

func (t *termGridRenderer) ApplyTheme() {
}

func (t *termGridRenderer) Objects() []fyne.CanvasObject {
	return t.objects
}

func (t *termGridRenderer) Destroy() {
}

func (t *termGridRenderer) refresh(obj fyne.CanvasObject) {
	if t.current == nil {
		if fyne.CurrentApp() != nil && fyne.CurrentApp().Driver() != nil {
			// cache canvas for this widget, so we don't look it up many times for every cell/row refresh!
			t.current = fyne.CurrentApp().Driver().CanvasForObject(t.text)
		}

		if t.current == nil {
			return // not yet set up perhaps?
		}
	}

	t.current.Refresh(obj)
}

func (t *termGridRenderer) updateCellSize() {
	size := fyne.MeasureText("M", theme.TextSize(), fyne.TextStyle{Monospace: true})

	// round it for seamless background
	size.Width = float32(math.Round(float64((size.Width))))
	size.Height = float32(math.Round(float64((size.Height))))

	t.cellSize = size
}

func (t *termGridRenderer) SetBlink(b bool) {
	t.blink = b
}
