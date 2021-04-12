package terminal

import (
	"io"
	"log"
	"os"
	"syscall"

	"fyne.io/fyne/v2"
	"github.com/ActiveState/termtest/conpty"
)

func (t *Terminal) updatePTYSize() {
	// TODO windows pty resize
}

func (t *Terminal) startPTY() (io.WriteCloser, io.Reader, *os.File, error) {
	cpty, err := conpty.New(80, 25)
	if err != nil {
		return nil, nil, nil, err
	}

	pid, _, err := cpty.Spawn(
		"C:\\WINDOWS\\System32\\WindowsPowerShell\\v1.0\\powershell.exe",
		[]string{},
		&syscall.ProcAttr{
			Env: os.Environ(),
		},
	)
	if err != nil {
		return nil, nil, nil, err
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return nil, nil, nil, err
	}
	go func() {
		_, err := process.Wait()
		if err != nil {
			log.Fatalf("Error waiting for process: %v", err)
		}
		cpty.Close()
		fyne.CurrentApp().Quit()
	}()

	return cpty.InPipe(), cpty.OutPipe(), nil, nil
}
