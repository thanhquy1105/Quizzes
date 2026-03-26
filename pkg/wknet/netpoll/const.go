package netpoll

type PollEvent int

const (
	PollEventUnknown PollEvent = 1 << iota
	PollEventRead
	PollEventWrite
	PollEventClose
)
