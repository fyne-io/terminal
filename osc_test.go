package terminal

import (
	"os"
	"testing"

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
