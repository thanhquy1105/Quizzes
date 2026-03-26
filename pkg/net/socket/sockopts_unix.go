//go:build linux || freebsd || dragonfly
// +build linux freebsd dragonfly

package socket

import (
	"errors"
	"os"

	"golang.org/x/sys/unix"
)

func SetKeepAlivePeriod(fd, secs int) error {
	if secs <= 0 {
		return errors.New("invalid time duration")
	}
	if err := os.NewSyscallError("setsockopt", unix.SetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_KEEPALIVE, 1)); err != nil {
		return err
	}
	if err := os.NewSyscallError("setsockopt", unix.SetsockoptInt(fd, unix.IPPROTO_TCP, unix.TCP_KEEPINTVL, secs)); err != nil {
		return err
	}
	return os.NewSyscallError("setsockopt", unix.SetsockoptInt(fd, unix.IPPROTO_TCP, unix.TCP_KEEPIDLE, secs))
}
