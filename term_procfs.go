//go:build !windows && !darwin
// +build !windows,!darwin

package terminal

import (
	"os"
	"strconv"
)

// processDir returns the current working directory of the process with the given ID,
// or an empty string if it could not be found.
func processDir(pid int) string {
	wd, _ := os.Readlink("/proc/" + strconv.Itoa(pid) + "/cwd")
	return wd
}
