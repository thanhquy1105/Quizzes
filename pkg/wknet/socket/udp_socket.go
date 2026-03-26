//go:build linux || freebsd || dragonfly || darwin
// +build linux freebsd dragonfly darwin

package socket

import (
	"net"
	"os"
	"syscall"

	"golang.org/x/sys/unix"

	"btaskee-quiz/pkg/errors"
)

func GetUDPSockAddr(proto, addr string) (sa syscall.Sockaddr, family int, udpAddr *net.UDPAddr, ipv6only bool, err error) {
	var udpVersion string

	udpAddr, err = net.ResolveUDPAddr(proto, addr)
	if err != nil {
		return
	}

	udpVersion, err = determineUDPProto(proto, udpAddr)
	if err != nil {
		return
	}

	switch udpVersion {
	case "udp4":
		family = unix.AF_INET
		sa, err = ipToSockaddr(family, udpAddr.IP, udpAddr.Port, "")
	case "udp6":
		ipv6only = true
		fallthrough
	case "udp":
		family = unix.AF_INET6
		sa, err = ipToSockaddr(family, udpAddr.IP, udpAddr.Port, udpAddr.Zone)
	default:
		err = errors.ErrUnsupportedProtocol
	}

	return
}

func determineUDPProto(proto string, addr *net.UDPAddr) (string, error) {

	if addr.IP.To4() != nil {
		return "udp4", nil
	}

	if addr.IP.To16() != nil {
		return "udp6", nil
	}

	switch proto {
	case "udp", "udp4", "udp6":
		return proto, nil
	}

	return "", errors.ErrUnsupportedUDPProtocol
}

func udpSocket(proto, addr string, connect bool, sockOpts ...Option) (fd int, netAddr net.Addr, err error) {
	var (
		family   int
		ipv6only bool
		sa       syscall.Sockaddr
	)

	if sa, family, netAddr, ipv6only, err = GetUDPSockAddr(proto, addr); err != nil {
		return
	}

	if fd, err = sysSocket(family, syscall.SOCK_DGRAM, syscall.IPPROTO_UDP); err != nil {
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

	if err = os.NewSyscallError("setsockopt", syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_BROADCAST, 1)); err != nil {
		return
	}

	for _, sockOpt := range sockOpts {
		if err = sockOpt.SetSockOpt(fd, sockOpt.Opt); err != nil {
			return
		}
	}

	if connect {
		err = os.NewSyscallError("connect", syscall.Connect(fd, sa))
	} else {
		err = os.NewSyscallError("bind", syscall.Bind(fd, sa))
	}

	return
}
