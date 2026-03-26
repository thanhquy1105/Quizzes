package wkserver

import (
	"time"

	"btaskee-quiz/pkg/wkserver/proto"
	"github.com/WuKongIM/crypto/tls"
)

type Options struct {
	Addr            string
	WSAddr          string // websocket listening address
	WSSAddr         string // websocket secure listening address
	WSTLSConfig     *tls.Config
	RequestPoolSize int // The size of the request pool, the default is 10000
	MessagePoolSize int  // The size of the message pool, the default is 10000
	MessagePoolOn   bool // Whether to open the message pool, the default is true
	ConnPath        string
	ClosePath       string
	RequestTimeout  time.Duration
	OnMessage       func(conn Conn, msg *proto.Message)
	MaxIdle         time.Duration

	TimingWheelTick time.Duration // The time-round training interval must be 1ms or more
	TimingWheelSize int64         // Time wheel size
	OnRequest       func(conn Conn, req *proto.Request)
	OnResponse      func(conn Conn, resp *proto.Response)
	LogDetailOn     bool // 是否开启详细日志

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

func WithRequestTimeout(timeout time.Duration) Option {
	return func(o *Options) {
		o.RequestTimeout = timeout
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

func WithMessagePoolOn(on bool) Option {
	return func(o *Options) {
		o.MessagePoolOn = on
	}
}

func WithAddr(addr string) Option {
	return func(o *Options) {
		o.Addr = addr
	}
}

func WithConnPath(path string) Option {
	return func(o *Options) {
		o.ConnPath = path
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

func WithWSTLSConfig(tlsConfig *tls.Config) Option {
	return func(o *Options) {
		o.WSTLSConfig = tlsConfig
	}
}

func WithClosePath(path string) Option {
	return func(o *Options) {
		o.ClosePath = path
	}
}

func WithMaxIdle(maxIdle time.Duration) Option {
	return func(o *Options) {
		o.MaxIdle = maxIdle
	}
}

func WithTimingWheelTick(tick time.Duration) Option {
	return func(o *Options) {
		o.TimingWheelTick = tick
	}
}

func WithTimingWheelSize(size int64) Option {
	return func(o *Options) {
		o.TimingWheelSize = size
	}
}

func WithOnMessage(onMessage func(conn Conn, msg *proto.Message)) Option {
	return func(o *Options) {
		o.OnMessage = onMessage
	}
}

func WithOnRequest(onRequest func(conn Conn, req *proto.Request)) Option {
	return func(o *Options) {
		o.OnRequest = onRequest
	}
}

func WithOnResponse(onResponse func(conn Conn, resp *proto.Response)) Option {
	return func(o *Options) {
		o.OnResponse = onResponse
	}
}
func WithLogDetailOn(on bool) Option {
	return func(opts *Options) {
		opts.LogDetailOn = on
	}
}
