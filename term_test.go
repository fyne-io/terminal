package terminal

import (
	"testing"
	"time"

	"fyne.io/fyne"
	_ "fyne.io/fyne/test"

	"github.com/stretchr/testify/assert"
)

func TestNewTerminal(t *testing.T) {
	term := NewTerminal()
	assert.NotNil(t, term)
	assert.NotNil(t, term.content)
}

func TestTerminal_Resize(t *testing.T) {
	term := NewTerminal()
	term.Resize(fyne.NewSize(45, 45))

	assert.Equal(t, uint16(3), term.config.Columns)
	assert.Equal(t, uint16(2), term.config.Rows)
}

func TestTerminal_AddListener(t *testing.T) {
	term := NewTerminal()
	listen := make(chan Config)
	term.AddListener(listen)
	assert.Equal(t, 1, len(term.listeners))

	go term.onConfigure()
	select {
	case <-listen: // passed
	case <-time.After(time.Millisecond * 100):
		t.Error("Failed waiting for configure callback")
	}
}
