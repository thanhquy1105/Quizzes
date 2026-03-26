package client

import "btaskee-quiz/pkg/wkserver/proto"

type Options struct {
	Addr  string
	Uid   string
	Token string

	HeartbeatTick        int
	HeartbeatTimeoutTick int
	OnMessage            func(msg *proto.Message)

	LogDetailOn bool
}

func NewOptions(opt ...Option) *Options {
	return &Options{
		HeartbeatTick:        10,
		HeartbeatTimeoutTick: 20,
	}
}

type Option func(*Options)

func WithUid(uid string) Option {
	return func(opts *Options) {
		opts.Uid = uid
	}
}

func WithToken(token string) Option {
	return func(opts *Options) {
		opts.Token = token
	}
}

func WithOnMessage(f func(msg *proto.Message)) Option {
	return func(opts *Options) {
		opts.OnMessage = f
	}
}
