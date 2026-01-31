// +build windows

package security

import (
	"os"
	"syscall"
)

var (
	kernel32            = syscall.NewLazyDLL("kernel32.dll")
	procIsDebuggerPresent = kernel32.NewProc("IsDebuggerPresent")
)

// CheckDebuggerWin returns true if a debugger is attached (Windows specific).
func CheckDebuggerWin() bool {
	flag, _, _ := procIsDebuggerPresent.Call()
	return flag != 0
}

// CheckVMFilesWin checks for VM-specific files on Windows.
func CheckVMFilesWin() bool {
	vmFiles := []string{
		"C:\\windows\\System32\\Drivers\\VBoxMouse.sys",
		"C:\\windows\\System32\\Drivers\\vmmouse.sys",
		"C:\\windows\\System32\\Drivers\\vmhgfs.sys",
	}
	for _, f := range vmFiles {
		if _, err := os.Stat(f); err == nil {
			return true
		}
	}
	return false
}
