package report

import (
	"os"

	"golang.org/x/sys/windows"
)

func enableVirtualTerminal(f *os.File) bool {
	handle := windows.Handle(f.Fd())

	var mode uint32
	if err := windows.GetConsoleMode(handle, &mode); err != nil {
		return false
	}
	if mode&windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING != 0 {
		return true
	}

	return windows.SetConsoleMode(
		handle, mode|windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING) == nil
}
