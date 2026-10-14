package terminal

import (
	"bytes"
	"testing"

	"fyne.io/fyne/v2"
	"github.com/stretchr/testify/assert"
)

func newQueryTerminal() (*Terminal, *bytes.Buffer) {
	buf := &bytes.Buffer{}
	term := New()
	term.in = NopCloser(buf)
	term.Resize(fyne.NewSize(400, 200))
	return term, buf
}

func TestAPC_StringTerminator(t *testing.T) {
	term, buf := newQueryTerminal()
	// kitty graphics probe as sent by opencode, then DA1 and some text
	term.handleOutput([]byte("\x1b_Gi=31337,s=1,v=1,a=q,t=d,f=24;AAAA\x1b\\\x1b[cHello"))

	assert.False(t, term.state.apc)
	assert.Equal(t, "Hello", term.Text())
	assert.Equal(t, "\x1b[?2;22c", buf.String())
}

func TestDCS_StringTerminator(t *testing.T) {
	term, buf := newQueryTerminal()
	term.handleOutput([]byte("\x1bP+q4d73\x1b\\Hi"))

	assert.False(t, term.state.dcs)
	assert.Equal(t, "Hi", term.Text())
	assert.Equal(t, "\x1bP0+r\x1b\\", buf.String())
}

func TestDeviceStatusReport(t *testing.T) {
	term, buf := newQueryTerminal()
	term.handleOutput([]byte("\x1b[3;5H\x1b[6n"))
	assert.Equal(t, "\x1b[3;5R", buf.String())

	buf.Reset()
	term.handleOutput([]byte("\x1b[?6n"))
	assert.Equal(t, "\x1b[?3;5R", buf.String())

	buf.Reset()
	term.handleOutput([]byte("\x1b[5n"))
	assert.Equal(t, "\x1b[0n", buf.String())
}

func TestRequestMode(t *testing.T) {
	term, buf := newQueryTerminal()
	term.handleOutput([]byte("\x1b[?2026$p\x1b[?25$p\x1b[?2004h\x1b[?2004$p\x1b[20$p"))

	assert.Equal(t, "", term.Text(), "no stray characters should be printed")
	assert.Equal(t, "\x1b[?2026;0$y\x1b[?25;1$y\x1b[?2004;1$y\x1b[20;2$y", buf.String())
}

func TestQueryVersion(t *testing.T) {
	term, buf := newQueryTerminal()
	term.handleOutput([]byte("\x1b[>0q\x1b[2 q"))

	assert.Equal(t, "", term.Text())
	assert.Equal(t, "\x1bP>|fyneterm\x1b\\", buf.String())
}

func TestKittyKeyboardQuery_Ignored(t *testing.T) {
	term, buf := newQueryTerminal()
	term.handleOutput([]byte("\x1b[?u\x1b[>1u\x1b[<u"))

	assert.Equal(t, "", term.Text())
	assert.Equal(t, "", buf.String())
}

func TestWindowOps_ReportSize(t *testing.T) {
	term, buf := newQueryTerminal()
	term.config.Rows = 24
	term.config.Columns = 80
	term.handleOutput([]byte("\x1b[18t"))

	assert.Equal(t, "\x1b[8;24;80t", buf.String())

	buf.Reset()
	term.handleOutput([]byte("\x1b[14t\x1b[16t"))
	assert.Regexp(t, `^\x1b\[4;\d+;\d+t\x1b\[6;\d+;\d+t$`, buf.String())
}

func TestOSC_ColorQuery(t *testing.T) {
	term, buf := newQueryTerminal()
	term.handleOutput([]byte("\x1b]11;?\a"))
	assert.Regexp(t, `^\x1b\]11;rgb:[0-9a-f]{4}/[0-9a-f]{4}/[0-9a-f]{4}\a$`, buf.String())

	buf.Reset()
	term.handleOutput([]byte("\x1b]10;?\x1b\\"))
	assert.Regexp(t, `^\x1b\]10;rgb:[0-9a-f]{4}/[0-9a-f]{4}/[0-9a-f]{4}\x1b\\$`, buf.String())

	buf.Reset()
	term.handleOutput([]byte("\x1b]2;Title\x1b\\"))
	assert.Equal(t, "Title", term.config.Title)
	assert.Equal(t, "", buf.String())
}
