//go:build !windows
// +build !windows

package terminal

import (
	"errors"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"fyne.io/fyne/v2"

	"github.com/stretchr/testify/assert"
)

func TestRunLocalShell_MissingStartDir(t *testing.T) {
	term := New()
	term.Resize(fyne.NewSize(45, 45))
	term.SetStartDir(filepath.Join(t.TempDir(), "missing"))

	lastKeyTime = time.Now() // the working directory is only checked after recent key presses
	assert.Error(t, term.RunLocalShell())

	// a failed start must not leave the directory check running against a missing process
	time.Sleep(time.Millisecond * 300)
}

func TestProcessDir(t *testing.T) {
	if runtime.GOOS != "linux" && runtime.GOOS != "darwin" {
		t.Skip("working directory lookup is not supported on " + runtime.GOOS)
	}

	dir, err := filepath.EvalSymlinks(t.TempDir())
	assert.NoError(t, err)

	cmd := exec.Command("sleep", "10")
	cmd.Dir = dir
	assert.NoError(t, cmd.Start())
	defer func() {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
	}()

	assert.Equal(t, dir, processDir(cmd.Process.Pid))
}

func TestProcessDir_MissingProcess(t *testing.T) {
	cmd := exec.Command("true")
	assert.NoError(t, cmd.Run())

	assert.Equal(t, "", processDir(cmd.Process.Pid))
}

func TestTerminal_Close(t *testing.T) {
	term := New()
	term.Resize(fyne.NewSize(45, 45))
	done := make(chan error)
	go func() {
		done <- term.RunLocalShell()
	}()

	// a program that will not exit by itself, or respond to input
	err := errors.New("NotYet")
	for err != nil {
		time.Sleep(50 * time.Millisecond)
		_, err = term.Write([]byte("sleep 600\n"))
	}
	child := 0
	assert.Eventually(t, func() bool {
		out, _ := exec.Command("pgrep", "-P", strconv.Itoa(term.cmd.Process.Pid), "sleep").Output()
		child, _ = strconv.Atoi(strings.TrimSpace(string(out)))
		return child != 0
	}, time.Second*5, time.Millisecond*20)
	running := func() bool {
		return syscall.Kill(child, 0) == nil
	}
	assert.True(t, running())

	term.Close()
	select {
	case err := <-done:
		assert.NoError(t, err)
	case <-time.After(time.Second * 5):
		t.Fatal("Terminal did not stop after Close")
	}
	assert.Eventually(t, func() bool { return !running() }, time.Second*2, time.Millisecond*20)
	assert.Nil(t, term.dirWatchDone)

	term.Close() // closing again is harmless
}

func TestTerminal_Close_BeforeOpen(t *testing.T) {
	term := New() // without a size the shell will not start
	done := make(chan error)
	go func() {
		done <- term.RunLocalShell()
	}()

	term.Close()
	select {
	case err := <-done:
		assert.NoError(t, err)
	case <-time.After(time.Second * 5):
		t.Fatal("Terminal did not stop after Close")
	}
	assert.Nil(t, term.cmd)
}
