package terminal

import (
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
