package terminal

import (
	"os"
	"strings"
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"github.com/stretchr/testify/assert"
)

func TestOSC_Title(t *testing.T) {
	term := New()
	assert.Equal(t, "", term.config.Title)

	term.handleOSC("0;Test")
	assert.Equal(t, "Test", term.config.Title)

	term.handleOSC("0;Testing;123")
	assert.Equal(t, "Testing;123", term.config.Title)
}

func TestOSC_LongStrings(t *testing.T) {
	term := New()
	term.Resize(fyne.NewSize(50, 50))
	handle := func(data string) {
		for len(data) > 0 { // in reads as the terminal gets them
			n := min(len(data), bufLen)
			term.handleOutput([]byte(data[:n]))
			data = data[n:]
		}
	}

	start := time.Now()
	long := strings.Repeat("a", maxCodeLen-2)
	handle("\x1b]0;" + long + "\x07")
	assert.Equal(t, long, term.config.Title)

	handle("\x1b]0;" + long + "too long\x07shown")
	assert.Equal(t, long, term.config.Title)
	assert.Equal(t, "shown", term.content.Text())
	assert.Less(t, time.Since(start), 10*time.Second)
}

func TestOSC_Directory(t *testing.T) {
	wd, err := os.Getwd()
	assert.NoError(t, err)
	dir := t.TempDir()
	term := New()
	configured := make(chan Config, 1)
	term.AddListener(configured)

	term.handleOSC("7;file://" + dir)
	assert.Equal(t, dir, term.config.PWD)
	assert.Equal(t, dir, (<-configured).PWD)
	now, err := os.Getwd()
	assert.NoError(t, err)
	assert.Equal(t, wd, now, "the process's working directory changed")

	term.handleOSC("7;file://some-host" + dir)
	assert.Equal(t, dir, term.config.PWD)
	term.handleOSC("7;file:///tmp/with%20space")
	assert.Equal(t, "/tmp/with space", term.config.PWD)
}
