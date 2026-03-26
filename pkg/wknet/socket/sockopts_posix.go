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

func SetNoDelay(fd, noDelay int) error {
	return os.NewSyscallError("setsockopt", unix.SetsockoptInt(fd, unix.IPPROTO_TCP, unix.TCP_NODELAY, noDelay))
}

func SetRecvBuffer(fd, size int) error {
	return unix.SetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_RCVBUF, size)
}

func SetSendBuffer(fd, size int) error {
	return unix.SetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_SNDBUF, size)
}

func SetReuseport(fd, reusePort int) error {
	return os.NewSyscallError("setsockopt", unix.SetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_REUSEPORT, reusePort))
}

func SetReuseAddr(fd, reuseAddr int) error {
	return os.NewSyscallError("setsockopt", unix.SetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_REUSEADDR, reuseAddr))
}

func SetIPv6Only(fd, ipv6only int) error {
	return unix.SetsockoptInt(fd, unix.IPPROTO_IPV6, unix.IPV6_V6ONLY, ipv6only)
}

func SetLinger(fd, sec int) error {
	var l unix.Linger
	if sec >= 0 {
		l.Onoff = 1
		l.Linger = int32(sec)
	} else {
		l.Onoff = 0
		l.Linger = 0
	}
	return unix.SetsockoptLinger(fd, syscall.SOL_SOCKET, syscall.SO_LINGER, &l)
}

func SetMulticastMembership(proto string, udpAddr *net.UDPAddr) func(int, int) error {
	udpVersion, err := determineUDPProto(proto, udpAddr)
	if err != nil {
		return nil
	}

	switch udpVersion {
	case "udp4":
		return func(fd int, ifIndex int) error {
			return SetIPv4MulticastMembership(fd, udpAddr.IP, ifIndex)
		}
	case "udp6":
		return func(fd int, ifIndex int) error {
			return SetIPv6MulticastMembership(fd, udpAddr.IP, ifIndex)
		}
	default:
		return nil
	}
}

func SetIPv4MulticastMembership(fd int, mcast net.IP, ifIndex int) error {

	ip, err := interfaceFirstIPv4Addr(ifIndex)
	if err != nil {
		return err
	}

	mreq := &unix.IPMreq{}
	copy(mreq.Multiaddr[:], mcast.To4())
	copy(mreq.Interface[:], ip.To4())

	if ifIndex > 0 {
		if err := os.NewSyscallError("setsockopt", unix.SetsockoptInet4Addr(fd, syscall.IPPROTO_IP, syscall.IP_MULTICAST_IF, mreq.Interface)); err != nil {
			return err
		}
	}

	if err := os.NewSyscallError("setsockopt", unix.SetsockoptByte(fd, syscall.IPPROTO_IP, syscall.IP_MULTICAST_LOOP, 0)); err != nil {
		return err
	}
	return os.NewSyscallError("setsockopt", unix.SetsockoptIPMreq(fd, syscall.IPPROTO_IP, syscall.IP_ADD_MEMBERSHIP, mreq))
}

func SetIPv6MulticastMembership(fd int, mcast net.IP, ifIndex int) error {
	mreq := &unix.IPv6Mreq{}
	mreq.Interface = uint32(ifIndex)
	copy(mreq.Multiaddr[:], mcast.To16())

	if ifIndex > 0 {
		if err := os.NewSyscallError("setsockopt", unix.SetsockoptInt(fd, syscall.IPPROTO_IPV6, syscall.IPV6_MULTICAST_IF, ifIndex)); err != nil {
			return err
		}
	}

	if err := os.NewSyscallError("setsockopt", unix.SetsockoptInt(fd, syscall.IPPROTO_IPV6, syscall.IPV6_MULTICAST_LOOP, 0)); err != nil {
		return err
	}
	return os.NewSyscallError("setsockopt", unix.SetsockoptIPv6Mreq(fd, syscall.IPPROTO_IPV6, syscall.IPV6_JOIN_GROUP, mreq))
}

func interfaceFirstIPv4Addr(ifIndex int) (net.IP, error) {
	if ifIndex == 0 {
		return net.IP([]byte{0, 0, 0, 0}), nil
	}
	iface, err := net.InterfaceByIndex(ifIndex)
	if err != nil {
		return nil, err
	}
	addrs, err := iface.Addrs()
	if err != nil {
		return nil, err
	}
	for _, addr := range addrs {
		ip, _, err := net.ParseCIDR(addr.String())
		if err != nil {
			return nil, err
		}
		if ip.To4() != nil {
			return ip, nil
		}
	}
	return nil, errors.ErrNoIPv4AddressOnInterface
}
