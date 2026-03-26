package net

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

func WithWSTLSConfig(v *tls.Config) Option {
	return func(opts *Options) {
		opts.WSTLSConfig = v
	}
}
