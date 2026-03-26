package server

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type Metrics struct {
	recvMsgCount    prometheus.Counter
	recvMsgBytes    prometheus.Counter
	sendMsgCount    prometheus.Counter
	sendMsgBytes    prometheus.Counter
	activeConnCount prometheus.Gauge
	authFailCount   prometheus.Counter
	joinQuizCount   prometheus.Counter
	answerQuizCount prometheus.Counter
	errCount        prometheus.Counter
}

func newMetrics() *Metrics {
	m := &Metrics{
		recvMsgCount: promauto.NewCounter(prometheus.CounterOpts{
			Name: "quiz_recv_msg_count_total",
			Help: "The total number of received messages",
		}),
		recvMsgBytes: promauto.NewCounter(prometheus.CounterOpts{
			Name: "quiz_recv_msg_bytes_total",
			Help: "The total number of received bytes",
		}),
		sendMsgCount: promauto.NewCounter(prometheus.CounterOpts{
			Name: "quiz_send_msg_count_total",
			Help: "The total number of sent messages",
		}),
		sendMsgBytes: promauto.NewCounter(prometheus.CounterOpts{
			Name: "quiz_send_msg_bytes_total",
			Help: "The total number of sent bytes",
		}),
		activeConnCount: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "quiz_active_conn_count",
			Help: "The current number of active connections",
		}),
		authFailCount: promauto.NewCounter(prometheus.CounterOpts{
			Name: "quiz_auth_fail_count_total",
			Help: "The total number of failed authentication attempts",
		}),
		joinQuizCount: promauto.NewCounter(prometheus.CounterOpts{
			Name: "quiz_join_count_total",
			Help: "The total number of users joining a quiz",
		}),
		answerQuizCount: promauto.NewCounter(prometheus.CounterOpts{
			Name: "quiz_answer_count_total",
			Help: "The total number of answers submitted",
		}),
		errCount: promauto.NewCounter(prometheus.CounterOpts{
			Name: "quiz_err_count_total",
			Help: "The total number of server errors",
		}),
	}
	return m
}

func (m *Metrics) RecvMsgCountAdd(v uint64) {
	m.recvMsgCount.Add(float64(v))
}

func (m *Metrics) RecvMsgBytesAdd(v uint64) {
	m.recvMsgBytes.Add(float64(v))
}

func (m *Metrics) SendMsgCountAdd(v uint64) {
	m.sendMsgCount.Add(float64(v))
}

func (m *Metrics) SendMsgBytesAdd(v uint64) {
	m.sendMsgBytes.Add(float64(v))
}

func (m *Metrics) ActiveConnInc() {
	m.activeConnCount.Inc()
}

func (m *Metrics) ActiveConnDec() {
	m.activeConnCount.Dec()
}

func (m *Metrics) AuthFailInc() {
	m.authFailCount.Inc()
}

func (m *Metrics) JoinQuizInc() {
	m.joinQuizCount.Inc()
}

func (m *Metrics) AnswerQuizInc() {
	m.answerQuizCount.Inc()
}

func (m *Metrics) ErrInc() {
	m.errCount.Inc()
}

func (m *Metrics) PrintMetrics(prefix string) {
	// Periodic logging if needed, but Prometheus handles visualization
}
