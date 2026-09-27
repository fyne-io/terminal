//go:build !windows
// +build !windows

package terminal

import (
	"os/exec"
	"path/filepath"
	"runtime"
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
