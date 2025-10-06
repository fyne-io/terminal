//go:build !windows
// +build !windows

package terminal

import (
	"io"
	"os"
	"os/exec"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"github.com/creack/pty"
)

func (t *Terminal) updatePTYSize() {
	if t.pty == nil { // SSH or other direct connection?
		return
	}
	scale := float32(1.0)
	c := fyne.CurrentApp().Driver().CanvasForObject(t)
	if c != nil {
		scale = c.Scale()
	}
	_ = pty.Setsize(t.pty.(*os.File), &pty.Winsize{
		Rows: uint16(t.config.Rows), Cols: uint16(t.config.Columns),
		X: uint16(t.Size().Width * scale), Y: uint16(t.Size().Height * scale)})
}

func (t *Terminal) startPTY() (io.WriteCloser, io.Reader, io.Closer, error) {
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "bash"
	}

	env := os.Environ()
	env = append(env, "TERM=xterm-256color")
	c := exec.Command(shell)
	c.Dir = t.startingDir()
	c.Env = env
	t.cmd = c

	go func() {
		for {
			time.Sleep(time.Millisecond * 250)
			if time.Since(lastKeyTime).Seconds() > 0.5 {
				continue
			}
			wd, _ := os.Readlink("/proc/" + strconv.Itoa(c.Process.Pid) + "/cwd")

			if wd != t.config.PWD {
				t.config.PWD = wd
				fyne.Do(t.onConfigure)
			}
		}
	}()

	// Start the command with a pty.
	f, err := pty.Start(c)
	return f, f, f, err
}
