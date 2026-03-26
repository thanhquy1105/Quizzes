package log

import "go.uber.org/zap/zapcore"

type Options struct {
	Level    zapcore.Level
	LogDir   string
	LineNum  bool
	TraceOn  bool
	NoStdout bool
}

func NewOptions() *Options {

	return &Options{}
}
