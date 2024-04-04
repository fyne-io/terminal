package terminal

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"fyne.io/fyne/v2"
	_ "fyne.io/fyne/v2/test"

	"github.com/stretchr/testify/assert"
)

func TestNewTerminal(t *testing.T) {
	term := New()
	assert.NotNil(t, term)
	assert.NotNil(t, term.content)
}

func TestExitCode(t *testing.T) {
	term := New()
	assert.Equal(t, term.ExitCode(), int(-1))
}

func testExitCodeN(t *testing.T, n int) {
	term := New()
	term.Resize(fyne.NewSize(45, 45))
	go term.RunLocalShell()
	err := errors.New("NotYet")
	for err != nil {
		time.Sleep(50 * time.Millisecond)
		_, err = term.Write([]byte(fmt.Sprintf("exit %d\n", n)))
	}
	for term.ExitCode() == -1 {
		time.Sleep(50 * time.Millisecond)
	}

	assert.Equal(t, n, term.ExitCode())
}

func TestExitCode01(t *testing.T) {
	testExitCodeN(t, 0)
	testExitCodeN(t, 1)
}

func TestTerminal_Resize(t *testing.T) {
	term := New()
	term.Resize(fyne.NewSize(45, 45))

	assert.Equal(t, uint(5), term.config.Columns)
	assert.Equal(t, uint(2), term.config.Rows)
}

func TestTerminal_AddListener(t *testing.T) {
	term := New()
	listen := make(chan Config, 1)
	term.AddListener(listen)
	assert.Equal(t, 1, len(term.listeners))

	go term.onConfigure()
	select {
	case <-listen: // passed
	case <-time.After(time.Millisecond * 100):
		t.Error("Failed waiting for configure callback")
	}
	term.RemoveListener(listen)
	assert.Equal(t, 0, len(term.listeners))
}

func TestTerminal_SanitizePosition(t *testing.T) {
	tests := []struct {
		name   string
		pos    fyne.Position
		width  int
		height int
		want   fyne.Position
	}{
		{
			name:   "WithinBounds",
			pos:    fyne.NewPos(2, 3),
			width:  45,
			height: 45,
			want:   fyne.NewPos(2, 3),
		},
		{
			name:   "NegativeX",
			pos:    fyne.NewPos(-1, 2),
			width:  45,
			height: 45,
			want:   fyne.NewPos(0, 2),
		},
		{
			name:   "ExceedsWidth",
			pos:    fyne.NewPos(46, 2),
			width:  45,
			height: 45,
			want:   fyne.NewPos(45, 2),
		},
		{
			name:   "NegativeY",
			pos:    fyne.NewPos(3, -1),
			width:  45,
			height: 45,
			want:   fyne.NewPos(3, 0),
		},
		{
			name:   "ExceedsHeight",
			pos:    fyne.NewPos(3, 46),
			width:  45,
			height: 45,
			want:   fyne.NewPos(3, 45),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			term := New()
			term.Resize(fyne.NewSize(float32(tt.width), float32(tt.height)))

			got := term.sanitizePosition(tt.pos)

			if *got != tt.want {
				t.Errorf("got %v, want %v", *got, tt.want)
			}
		})
	}
}
