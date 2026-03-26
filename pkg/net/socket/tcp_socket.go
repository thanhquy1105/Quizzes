//go:build linux || freebsd || dragonfly || darwin
// +build linux freebsd dragonfly darwin

package socket

import (
	"net"
	"os"
	"syscall"

	"btaskee-quiz/pkg/errors"
)

var listenerBacklogMaxSize = maxListenerBacklog()

func GetTCPSockAddr(proto, addr string) (sa syscall.Sockaddr, family int, tcpAddr *net.TCPAddr, ipv6only bool, err error) {
	var tcpVersion string

	tcpAddr, err = net.ResolveTCPAddr(proto, addr)
	if err != nil {
		return
	}

	tcpVersion, err = determineTCPProto(proto, tcpAddr)
	if err != nil {
		return
	}

	switch tcpVersion {
	case "tcp4":
		family = syscall.AF_INET
		sa, err = ipToSockaddr(family, tcpAddr.IP, tcpAddr.Port, "")
	case "tcp6":
		ipv6only = true
		fallthrough
	case "tcp":
		family = syscall.AF_INET6
		sa, err = ipToSockaddr(family, tcpAddr.IP, tcpAddr.Port, tcpAddr.Zone)
	default:
		err = errors.ErrUnsupportedProtocol
	}

	return
}

func determineTCPProto(proto string, addr *net.TCPAddr) (string, error) {

	if addr.IP.To4() != nil {
		return "tcp4", nil
	}

	if addr.IP.To16() != nil {
		return "tcp6", nil
	}

	switch proto {
	case "tcp", "tcp4", "tcp6":
		return proto, nil
	}

	return "", errors.ErrUnsupportedTCPProtocol
}

func tcpSocket(proto, addr string, passive bool, sockOpts ...Option) (fd int, netAddr net.Addr, err error) {
	var (
		family   int
		ipv6only bool
		sa       syscall.Sockaddr
	)

	if sa, family, netAddr, ipv6only, err = GetTCPSockAddr(proto, addr); err != nil {
		return
	}

	if fd, err = sysSocket(family, syscall.SOCK_STREAM, syscall.IPPROTO_TCP); err != nil {
		err = os.NewSyscallError("socket", err)
		return
	}
	defer func() {

		if err != nil {
			if err, ok := err.(*os.SyscallError); ok && err.Err == syscall.EINPROGRESS {
				return
			}
			_ = syscall.Close(fd)
		}
	}()

	if family == syscall.AF_INET6 && ipv6only {
		if err = SetIPv6Only(fd, 1); err != nil {
			return
		}
	}

	for _, sockOpt := range sockOpts {
		if err = sockOpt.SetSockOpt(fd, sockOpt.Opt); err != nil {
			return
		}
	}

	if passive {
		if err = os.NewSyscallError("bind", syscall.Bind(fd, sa)); err != nil {
			return
		}

		err = os.NewSyscallError("listen", syscall.Listen(fd, listenerBacklogMaxSize))
	} else {
		err = os.NewSyscallError("connect", syscall.Connect(fd, sa))
	}

	return
}
