package server

import (
	"time"

	"btaskee-quiz/internal/server/proto"
)

type Options struct {
	TCPAddr         string
	WSAddr          string
	RequestPoolSize int
	MessagePoolSize int
	MessagePoolOn   bool
	ConnPath        string
	ClosePath       string
	RequestTimeout  time.Duration
	OnMessage       func(conn Conn, msg *proto.Message)
	MaxIdle         time.Duration

	TimingWheelTick time.Duration
	TimingWheelSize int64
	OnRequest       func(conn Conn, req *proto.Request)
	OnResponse      func(conn Conn, resp *proto.Response)
	LogDetailOn     bool
}

func NewOptions() *Options {

	return &Options{
		TCPAddr:         "tcp://0.0.0.0:12000",
		RequestPoolSize: 20000,
		MessagePoolSize: 40000,
		MessagePoolOn:   true,
		ConnPath:        "/conn",
		ClosePath:       "/close",
		RequestTimeout:  10 * time.Second,
		MaxIdle:         120 * time.Second,
		TimingWheelTick: time.Millisecond * 10,
		TimingWheelSize: 100,
	}
}

type Option func(*Options)

func WithTCPAddr(addr string) Option {
	return func(o *Options) {
		o.TCPAddr = addr
	}
}

func WithWSAddr(addr string) Option {
	return func(o *Options) {
		o.WSAddr = addr
	}
}

func WithOnMessage(onMessage func(conn Conn, msg *proto.Message)) Option {
	return func(o *Options) {
		o.OnMessage = onMessage
	}
}

func WithRequestPoolSize(size int) Option {
	return func(o *Options) {
		o.RequestPoolSize = size
	}
}

func WithMessagePoolSize(size int) Option {
	return func(o *Options) {
		o.MessagePoolSize = size
	}
}

func WithMaxIdle(d time.Duration) Option {
	return func(o *Options) {
		o.MaxIdle = d
	}
}

func WithRequestTimeout(d time.Duration) Option {
	return func(o *Options) {
		o.RequestTimeout = d
	}
}

func WithLogDetail(on bool) Option {
	return func(o *Options) {
		o.LogDetailOn = on
	}
}
