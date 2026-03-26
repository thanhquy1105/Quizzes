package errors

import "errors"

var (
	ErrEmptyEngine = errors.New("the internal engine is empty")

	ErrEngineShutdown = errors.New("server is going to be shutdown")

	ErrEngineInShutdown = errors.New("server is already in shutdown")

	ErrAcceptSocket = errors.New("accept a new connection error")

	ErrTooManyEventLoopThreads = errors.New("too many event-loops under LockOSThread mode")

	ErrUnsupportedProtocol = errors.New("only unix, tcp/tcp4/tcp6, udp/udp4/udp6 are supported")

	ErrUnsupportedTCPProtocol = errors.New("only tcp/tcp4/tcp6 are supported")

	ErrUnsupportedUDPProtocol = errors.New("only udp/udp4/udp6 are supported")

	ErrUnsupportedUDSProtocol = errors.New("only unix is supported")

	ErrUnsupportedPlatform = errors.New("unsupported platform in gnet")

	ErrUnsupportedOp = errors.New("unsupported operation")

	ErrNegativeSize = errors.New("negative size is invalid")

	ErrNoIPv4AddressOnInterface = errors.New("no IPv4 address on interface")
)
