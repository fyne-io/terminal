package widget

import (
	"context"
	"image/color"
	"math"
	"strconv"
	"time"

	"fyne.io/fyne/v2/widget"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
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
	th := t.text.Theme()
	v := fyne.CurrentApp().Settings().ThemeVariant()

	text := canvas.NewText(string(str), th.Color(theme.ColorNameForeground, v))
	text.TextStyle.Monospace = true

	bg := canvas.NewRectangle(color.Transparent)

	ul := canvas.NewLine(color.Transparent)

	t.objects = append(t.objects, bg, text, ul)
}

func (t *termGridRenderer) setCellRune(str rune, pos int, style widget.TextGridStyle) {
	if str == 0 {
		str = ' '
	}
	rect := t.objects[pos*3].(*canvas.Rectangle)
	text := t.objects[pos*3+1].(*canvas.Text)
	underline := t.objects[pos*3+2].(*canvas.Line)

	th := t.text.Theme()
	v := fyne.CurrentApp().Settings().ThemeVariant()
	fg := th.Color(theme.ColorNameForeground, v)
	bg := color.Color(color.Transparent)
	text.TextSize = th.Size(theme.SizeNameText)
	textStyle := fyne.TextStyle{}
	var underlineStrokeWidth float32 = 1
	var underlineStrokeColor color.Color = color.Transparent

	if style != nil && style.TextColor() != nil {
		fg = style.TextColor()
	}

	if style != nil {
		if style.BackgroundColor() != nil {
			bg = style.BackgroundColor()
		}
		if style.Style().Bold {
			underlineStrokeWidth = 2
			textStyle = fyne.TextStyle{
				Bold: true,
			}
		}
		if style.Style().Underline {
			underlineStrokeColor = fg
		}
	}

	if s, ok := style.(*TermTextGridStyle); ok && s != nil && s.BlinkEnabled {
		t.shouldBlink = true
		if t.blink {
			fg = bg
			underlineStrokeColor = bg
		}
	}

	newStr := string(str)
	if text.Text != newStr || text.Color != fg || textStyle != text.TextStyle {
		text.Text = newStr
		text.Color = fg
		text.TextStyle = textStyle
		t.refresh(text)
	}

	if underlineStrokeWidth != underline.StrokeWidth || underlineStrokeColor != underline.StrokeColor {
		underline.StrokeWidth, underline.StrokeColor = underlineStrokeWidth, underlineStrokeColor
		t.refresh(underline)
	}

	if rect.FillColor != bg {
		rect.FillColor = bg
		t.refresh(rect)
	}
}

func (t *termGridRenderer) addCellsIfRequired() {
	cellCount := t.cols * t.rows
	if len(t.objects) == cellCount*3 {
		return
	}
	for i := len(t.objects); i < cellCount*3; i += 3 {
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
	for ; x < len(t.objects)/3; x++ {
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
			// rect
			t.objects[i*3].Resize(t.cellSize)
			t.objects[i*3].Move(cellPos)

			// text
			t.objects[i*3+1].Move(cellPos)

			// underline
			t.objects[i*3+2].Move(cellPos.Add(fyne.Position{X: 0, Y: t.cellSize.Height}))
			t.objects[i*3+2].Resize(fyne.Size{Width: t.cellSize.Width})

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

	th := t.text.Theme()
	v := fyne.CurrentApp().Settings().ThemeVariant()
	widget.TextGridStyleWhitespace = &widget.CustomTextGridStyle{FGColor: th.Color(theme.ColorNameDisabled, v)}
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
	th := t.text.Theme()
	size := fyne.MeasureText("M", th.Size(theme.SizeNameText), fyne.TextStyle{Monospace: true})

	// round it for seamless background
	size.Width = float32(math.Round(float64(size.Width)))
	size.Height = float32(math.Round(float64(size.Height)))

	t.cellSize = size
}

func (t *termGridRenderer) SetBlink(b bool) {
	t.blink = b
}
