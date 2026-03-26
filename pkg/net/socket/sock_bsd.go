//go:build freebsd || dragonfly || darwin
// +build freebsd dragonfly darwin

package socket

import (
	"runtime"

	"golang.org/x/sys/unix"
)

func maxListenerBacklog() int {
	var (
		n   uint32
		err error
	)
	switch runtime.GOOS {
	case "darwin":
		n, err = unix.SysctlUint32("kern.ipc.somaxconn")
	case "freebsd":
		n, err = unix.SysctlUint32("kern.ipc.soacceptqueue")
	}
	if n == 0 || err != nil {
		return unix.SOMAXCONN
	}

	if n > 1<<16-1 {
		n = 1<<16 - 1
	}
	return int(n)
}
