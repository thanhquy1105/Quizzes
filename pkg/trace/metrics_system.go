package trace

import (
	"context"
	"runtime"

	"btaskee-quiz/pkg/wklog"
	"go.opentelemetry.io/otel/metric"
	"go.uber.org/atomic"
)

type systemMetrics struct {
	opts *Options
	ctx  context.Context
	wklog.Log

	intranetIncomingBytes atomic.Int64
	intranetOutgoingBytes atomic.Int64
	extranetIncomingBytes atomic.Int64
	extranetOutgoingBytes atomic.Int64
}

func newSystemMetrics(opts *Options) *systemMetrics {
	s := &systemMetrics{
		ctx:  context.Background(),
		opts: opts,
		Log:  wklog.NewWKLog("systemMetrics"),
	}

	intranetIncomingBytes := NewInt64ObservableCounter("system_intranet_incoming_bytes")
	intranetOutgoingBytes := NewInt64ObservableCounter("system_intranet_outgoing_bytes")
	extranetIncomingBytes := NewInt64ObservableCounter("system_extranet_incoming_bytes")
	extranetOutgoingBytes := NewInt64ObservableCounter("system_extranet_outgoing_bytes")
	cpuUsage := NewFloat64ObservableCounter("system_cpu_percent")

	RegisterCallback(func(ctx context.Context, obs metric.Observer) error {
		obs.ObserveInt64(intranetIncomingBytes, s.intranetIncomingBytes.Load())
		obs.ObserveInt64(intranetOutgoingBytes, s.intranetOutgoingBytes.Load())
		obs.ObserveInt64(extranetIncomingBytes, s.extranetIncomingBytes.Load())
		obs.ObserveInt64(extranetOutgoingBytes, s.extranetOutgoingBytes.Load())
		cpuPercent := float64(runtime.NumCPU())
		obs.ObserveFloat64(cpuUsage, cpuPercent)

		return nil
	}, intranetIncomingBytes, intranetOutgoingBytes, extranetIncomingBytes, extranetOutgoingBytes, cpuUsage)

	return s
}

func (s *systemMetrics) IntranetIncomingAdd(v int64) {
	s.intranetIncomingBytes.Add(v)
}

func (s *systemMetrics) IntranetOutgoingAdd(v int64) {
	s.intranetOutgoingBytes.Add(v)
}

func (s *systemMetrics) ExtranetIncomingAdd(v int64) {
	s.extranetIncomingBytes.Add(v)
}

func (s *systemMetrics) ExtranetOutgoingAdd(v int64) {
	s.extranetOutgoingBytes.Add(v)
}

func (s *systemMetrics) CPUUsageAdd(v float64) {

}

func (s *systemMetrics) MemoryUsageAdd(v float64) {

}

func (s *systemMetrics) DiskIOReadCountAdd(v int64) {

}

func (s *systemMetrics) DiskIOWriteCountAdd(v int64) {

}
