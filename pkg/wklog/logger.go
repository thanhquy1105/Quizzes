package wklog

import (
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var logger *zap.Logger
var traceLogger *zap.Logger
var errorLogger *zap.Logger
var warnLogger *zap.Logger
var panicLogger *zap.Logger
var focusLogger *zap.Logger
var atom = zap.NewAtomicLevel()

var opts *Options

func Configure(op *Options) {
	atom.SetLevel(op.Level)
	opts = op

	loggerOpts := make([]zap.Option, 0)
	if opts.LineNum {
		loggerOpts = append(loggerOpts, zap.AddCaller(), zap.AddCallerSkip(2))
	}

	writers := make([]zapcore.WriteSyncer, 0)
	if !opts.NoStdout {
		writers = append(writers, zapcore.AddSync(os.Stdout))
	}

	infoWriter := zapcore.AddSync(&lumberjack.Logger{
		Filename:   path.Join(opts.LogDir, "info.log"),
		MaxSize:    500,
		MaxBackups: 3,
		MaxAge:     28,
	})
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(newEncoderConfig()),
		zapcore.NewMultiWriteSyncer(append(writers, zapcore.AddSync(infoWriter))...),
		atom,
	)
	logger = zap.New(core, loggerOpts...)

	traceWriter := zapcore.AddSync(&lumberjack.Logger{
		Filename:   path.Join(opts.LogDir, "trace.log"),
		MaxSize:    500,
		MaxBackups: 3,
		MaxAge:     28,
	})
	core = zapcore.NewCore(
		zapcore.NewJSONEncoder(newEncoderConfig()),
		zapcore.NewMultiWriteSyncer(append(writers, zapcore.AddSync(traceWriter))...),
		atom,
	)
	traceLogger = zap.New(core, loggerOpts...)

	errorWriter := zapcore.AddSync(&lumberjack.Logger{
		Filename:   path.Join(opts.LogDir, "error.log"),
		MaxSize:    500,
		MaxBackups: 3,
		MaxAge:     28,
	})
	core = zapcore.NewCore(
		zapcore.NewJSONEncoder(newEncoderConfig()),
		zapcore.NewMultiWriteSyncer(append(writers, zapcore.AddSync(errorWriter))...),
		zap.ErrorLevel,
	)
	errorLogger = zap.New(core, loggerOpts...)

	warnWriter := zapcore.AddSync(&lumberjack.Logger{
		Filename:   path.Join(opts.LogDir, "warn.log"),
		MaxSize:    500,
		MaxBackups: 3,
		MaxAge:     28,
	})
	core = zapcore.NewCore(
		zapcore.NewJSONEncoder(newEncoderConfig()),
		zapcore.NewMultiWriteSyncer(append(writers, zapcore.AddSync(warnWriter))...),
		zap.WarnLevel,
	)
	warnLogger = zap.New(core, loggerOpts...)

	panicWriter := zapcore.AddSync(&lumberjack.Logger{
		Filename:   path.Join(opts.LogDir, "panic.log"),
		MaxSize:    500,
		MaxBackups: 3,
		MaxAge:     28,
	})
	core = zapcore.NewCore(
		zapcore.NewJSONEncoder(newEncoderConfig()),
		zapcore.NewMultiWriteSyncer(append(writers, zapcore.AddSync(panicWriter))...),
		zap.PanicLevel,
	)
	panicLogger = zap.New(core, append(loggerOpts, zap.AddStacktrace(zapcore.PanicLevel))...)

	focusWriter := zapcore.AddSync(&lumberjack.Logger{
		Filename:   path.Join(opts.LogDir, "focus.log"),
		MaxSize:    500,
		MaxBackups: 3,
		MaxAge:     28,
	})
	core = zapcore.NewCore(
		zapcore.NewJSONEncoder(newEncoderConfig()),
		zapcore.NewMultiWriteSyncer(append(writers, zapcore.AddSync(focusWriter))...),
		zap.InfoLevel,
	)
	focusLogger = zap.New(core, loggerOpts...)

}

func Level() zapcore.Level {
	if opts == nil {
		return zapcore.InfoLevel
	}
	return opts.Level
}

func newEncoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{

		TimeKey:       "time",
		LevelKey:      "level",
		NameKey:       "logger",
		CallerKey:     "linenum",
		MessageKey:    "msg",
		StacktraceKey: "stacktrace",
		LineEnding:    zapcore.DefaultLineEnding,
		EncodeLevel:   zapcore.LowercaseLevelEncoder,
		EncodeCaller:  zapcore.FullCallerEncoder,
		EncodeName:    zapcore.FullNameEncoder,
		EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(t.Format("2006-01-02T15:04:05.999999999-07:00"))
		},
		EncodeDuration: func(d time.Duration, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendInt64(int64(d) / 1000000)
		},
	}
}

func Info(msg string, fields ...zap.Field) {

	if logger == nil {
		Configure(NewOptions())
	}
	logger.Info(msg, fields...)

}

func Trace(msg string, fields ...zap.Field) {

	if traceLogger == nil {
		Configure(NewOptions())
	}
	traceLogger.Info(msg, fields...)

}

func Debug(msg string, fields ...zap.Field) {

	if logger == nil {
		Configure(NewOptions())
	}
	logger.Debug(msg, fields...)

}

func Error(msg string, fields ...zap.Field) {

	if errorLogger == nil {
		Configure(NewOptions())
	}
	errorLogger.Error(msg, fields...)

}

func Fatal(msg string, fields ...zap.Field) {

	if panicLogger == nil {
		Configure(NewOptions())
	}
	panicLogger.Fatal(msg, fields...)
}
func Panic(msg string, fields ...zap.Field) {

	if panicLogger == nil {
		Configure(NewOptions())
	}
	panicLogger.Panic(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {

	if warnLogger == nil {
		Configure(NewOptions())
	}
	warnLogger.Warn(msg, fields...)
}

func Foucs(msg string, fields ...zap.Field) {

	if focusLogger == nil {
		Configure(NewOptions())
	}
	focusLogger.Info(msg, fields...)
}

func Sync() error {
	err := panicLogger.Sync()
	if err != nil {
		fmt.Println("panicLogger sync error", err)
	}
	err = errorLogger.Sync()
	if err != nil {
		fmt.Println("errorLogger sync error", err)
	}
	err = warnLogger.Sync()
	if err != nil {
		fmt.Println("warnLogger sync error", err)
	}
	err = logger.Sync()
	if err != nil {
		fmt.Println("logger sync error", err)
	}
	return nil
}

type Log interface {
	Info(msg string, fields ...zap.Field)
	MessageTrace(msg string, clientMsgNo string, operationName string, fields ...zap.Field)
	Trace(msg string, action string, fields ...zap.Field)
	Debug(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
	Fatal(msg string, fields ...zap.Field)
	Panic(msg string, fields ...zap.Field)
	Foucs(msg string, fields ...zap.Field)
}

type WKLog struct {
	prefix string
}

func NewWKLog(prefix string) *WKLog {

	return &WKLog{prefix: prefix}
}

func (t *WKLog) Info(msg string, fields ...zap.Field) {
	var b strings.Builder
	b.WriteString("")
	b.WriteString(t.prefix)
	b.WriteString("")
	b.WriteString(msg)
	Info(b.String(), fields...)
}

func (t *WKLog) Trace(msg string, action string, fields ...zap.Field) {
	if !opts.TraceOn {
		return
	}

	var b strings.Builder
	b.WriteString("")
	b.WriteString(t.prefix)
	b.WriteString("")
	b.WriteString(msg)
	if len(fields) == 0 {
		Trace(b.String(), zap.Int("trace", 1), zap.String("action", action))
	} else {
		fields = append(fields, zap.Int("trace", 1), zap.String("action", action))
		Trace(b.String(), fields...)
	}
}

func (t *WKLog) MessageTrace(msg string, no string, action string, fields ...zap.Field) {

	if !opts.TraceOn {
		return
	}

	var b strings.Builder
	b.WriteString("")
	b.WriteString(t.prefix)
	b.WriteString("")
	b.WriteString(msg)
	if len(fields) == 0 {
		Trace(b.String(), zap.Int("trace", 1), zap.String("no", no), zap.String("action", action))
	} else {
		fields = append(fields, zap.Int("trace", 1), zap.String("no", no), zap.String("action", action))
		Trace(b.String(), fields...)
	}

}

func (t *WKLog) Debug(msg string, fields ...zap.Field) {
	var b strings.Builder
	b.WriteString("")
	b.WriteString(t.prefix)
	b.WriteString("")
	b.WriteString(msg)
	Debug(b.String(), fields...)
}

func (t *WKLog) Error(msg string, fields ...zap.Field) {
	var b strings.Builder
	b.WriteString("")
	b.WriteString(t.prefix)
	b.WriteString("")
	b.WriteString(msg)
	Error(b.String(), fields...)
}

func (t *WKLog) Warn(msg string, fields ...zap.Field) {
	var b strings.Builder
	b.WriteString("")
	b.WriteString(t.prefix)
	b.WriteString("")
	b.WriteString(msg)
	Warn(b.String(), fields...)
}

func (t *WKLog) Fatal(msg string, fields ...zap.Field) {
	var b strings.Builder
	b.WriteString("")
	b.WriteString(t.prefix)
	b.WriteString("")
	b.WriteString(msg)
	Fatal(b.String(), fields...)
}
func (t *WKLog) Panic(msg string, fields ...zap.Field) {
	var b strings.Builder
	b.WriteString("")
	b.WriteString(t.prefix)
	b.WriteString("")
	b.WriteString(msg)
	Panic(b.String(), fields...)
}

func (t *WKLog) Foucs(msg string, fields ...zap.Field) {
	var b strings.Builder
	b.WriteString("")
	b.WriteString(t.prefix)
	b.WriteString("")
	b.WriteString(msg)
	Foucs(b.String(), fields...)
}
