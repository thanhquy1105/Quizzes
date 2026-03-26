//go:build linux || freebsd || dragonfly
// +build linux freebsd dragonfly

package socket

import "golang.org/x/sys/unix"

func sysSocket(family, sotype, proto int) (int, error) {
	return unix.Socket(family, sotype|unix.SOCK_NONBLOCK|unix.SOCK_CLOEXEC, proto)
}
