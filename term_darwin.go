package terminal

import (
	"bytes"
	"syscall"
	"unsafe"
)

const (
	procInfoCallPIDInfo  = 2 // PROC_INFO_CALL_PIDINFO
	procPIDVnodePathInfo = 9 // PROC_PIDVNODEPATHINFO

	maxPathLen    = 1024 // MAXPATHLEN
	vnodeInfoSize = 152  // sizeof(struct vnode_info), this comes before the path in a vnode_info_path

	// sizeof(struct proc_vnodepathinfo), which is the current directory then the root directory
	vnodePathInfoSize = 2 * (vnodeInfoSize + maxPathLen)
)

// processDir returns the current working directory of the process with the given ID,
// or an empty string if it could not be found.
func processDir(pid int) string {
	var info [vnodePathInfoSize]byte
	n, _, errno := syscall.Syscall6(syscall.SYS_PROC_INFO, procInfoCallPIDInfo, uintptr(pid),
		procPIDVnodePathInfo, 0, uintptr(unsafe.Pointer(&info[0])), vnodePathInfoSize)
	if errno != 0 || n != vnodePathInfoSize {
		return ""
	}

	path := info[vnodeInfoSize : vnodeInfoSize+maxPathLen]
	if end := bytes.IndexByte(path, 0); end >= 0 {
		path = path[:end]
	}
	return string(path)
}
