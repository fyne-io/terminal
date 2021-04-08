package terminal

import (
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
