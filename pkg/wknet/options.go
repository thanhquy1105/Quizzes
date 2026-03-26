package wknet

import (
	"runtime"
	"time"

	"github.com/WuKongIM/crypto/tls"
)

type Mode int

type Options struct {
	Addr string

	TCPTLSConfig *tls.Config
	WSTLSConfig  *tls.Config

	WsAddr  string
	WssAddr string

	MaxOpenFiles int

	SubReactorNum int

	ReadBufferSize int

	MaxWriteBufferSize int

	MaxReadBufferSize int

	SocketRecvBuffer int

	SocketSendBuffer int

	TCPKeepAlive time.Duration

	Event struct {
		OnReadBytes  func(n int)
		OnWirteBytes func(n int)
	}
}

func NewOptions() *Options {
	return &Options{
		Addr:               "tcp://127.0.0.1:5100",
		MaxOpenFiles:       GetMaxOpenFiles(),
		SubReactorNum:      runtime.NumCPU(),
		ReadBufferSize:     1024 * 32,
		MaxWriteBufferSize: 1024 * 1024 * 50,
		MaxReadBufferSize:  1024 * 1024 * 50,
	}
}

type Option func(opts *Options)

func WithAddr(v string) Option {
	return func(opts *Options) {
		opts.Addr = v
	}
}

func WithWSAddr(v string) Option {
	return func(opts *Options) {
		opts.WsAddr = v
	}
}

func WithWSSAddr(v string) Option {
	return func(opts *Options) {
		opts.WssAddr = v
	}
}

func WithTCPTLSConfig(v *tls.Config) Option {
	return func(opts *Options) {
		opts.TCPTLSConfig = v
	}
}

func WithWSTLSConfig(v *tls.Config) Option {
	return func(opts *Options) {
		opts.WSTLSConfig = v
	}
}

func WithMaxOpenFiles(v int) Option {
	return func(opts *Options) {
		opts.MaxOpenFiles = v
	}
}

func WithSubReactorNum(v int) Option {
	return func(opts *Options) {
		opts.SubReactorNum = v
	}
}

func WithSocketRecvBuffer(recvBuf int) Option {
	return func(opts *Options) {
		opts.SocketRecvBuffer = recvBuf
	}
}

func WithSocketSendBuffer(sendBuf int) Option {
	return func(opts *Options) {
		opts.SocketSendBuffer = sendBuf
	}
}

func WithTCPKeepAlive(v time.Duration) Option {
	return func(opts *Options) {
		opts.TCPKeepAlive = v
	}
}

func WithOnReadBytes(f func(n int)) Option {
	return func(opts *Options) {
		opts.Event.OnReadBytes = f
	}
}

func WithOnWirteBytes(f func(n int)) Option {

	return func(opts *Options) {
		opts.Event.OnWirteBytes = f
	}
}
