//go:build !windows
// +build !windows

package terminal

import (
	"io"
	"os"
	"os/exec"
	"syscall"
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
		X: uint16(t.Size().Width * scale), Y: uint16(t.Size().Height * scale),
	})
}

// hangup tells the shell that its terminal has gone away.
func (t *Terminal) hangup() {
	if t.cmd == nil || t.cmd.Process == nil {
		return // SSH or other direct connection
	}
	_ = t.cmd.Process.Signal(syscall.SIGHUP)
}

func (t *Terminal) startPTY() (io.WriteCloser, io.Reader, io.Closer, error) {
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "bash"
	}

	env := os.Environ()
	env = append(env, "TERM=xterm-256color")
	env = append(env, "COLORTERM=truecolor")
	c := exec.Command(shell)
	c.Dir = t.startingDir()
	c.Env = env
	t.cmd = c
	t.config.PWD = c.Dir

	// Start the command with a pty.
	f, err := pty.Start(c)
	if err != nil {
		return nil, nil, nil, err
	}

	done := make(chan struct{})
	t.dirWatchDone = done
	go func() {
		tick := time.NewTicker(time.Millisecond * 250)
		defer tick.Stop()
		for {
			select {
			case <-done:
				return
			case <-tick.C:
			}
			if time.Since(lastKeyTime).Seconds() > 0.5 {
				continue
			}
			wd := processDir(c.Process.Pid)

			if wd != "" && wd != t.config.PWD {
				t.config.PWD = wd
				fyne.Do(t.onConfigure)
			}
		}
	}()

	return f, f, f, nil
}
