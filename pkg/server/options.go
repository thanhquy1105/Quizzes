package server

import (
	"time"

	"btaskee-quiz/pkg/server/proto"

	"github.com/WuKongIM/crypto/tls"
)

type Options struct {
	Addr            string
	WSAddr          string
	WSSAddr         string
	WSTLSConfig     *tls.Config
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
		Addr:            "tcp://0.0.0.0:12000",
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

func WithAddr(addr string) Option {
	return func(o *Options) {
		o.Addr = addr
	}
}

func WithWSAddr(addr string) Option {
	return func(o *Options) {
		o.WSAddr = addr
	}
}

func WithWSSAddr(addr string) Option {
	return func(o *Options) {
		o.WSSAddr = addr
	}
}

func WithOnMessage(onMessage func(conn Conn, msg *proto.Message)) Option {
	return func(o *Options) {
		o.OnMessage = onMessage
	}
}
