//go:build linux || freebsd || dragonfly || darwin
// +build linux freebsd dragonfly darwin

package socket

import (
	"net"
)

type Option struct {
	SetSockOpt func(int, int) error
	Opt        int
}

func TCPSocket(proto, addr string, passive bool, sockOpts ...Option) (int, net.Addr, error) {
	return tcpSocket(proto, addr, passive, sockOpts...)
}
