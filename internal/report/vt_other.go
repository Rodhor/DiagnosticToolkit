//go:build !windows

package report

import "os"

func enableVirtualTerminal(*os.File) bool { return true }
