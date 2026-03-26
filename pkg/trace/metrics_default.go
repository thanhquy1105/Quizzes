package trace

import (
	"net/http"
	"sort"
	"strconv"
	"time"

	"btaskee-quiz/pkg/wkhttp"
	"btaskee-quiz/pkg/wklog"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
)

var (
	meter = otel.Meter("metrics")
)

type metrics struct {
	cluster IClusterMetrics
	app     IAppMetrics
	system  ISystemMetrics
	db      IDBMetrics
	opts    *Options
	wklog.Log
}

func newMetrics(opts *Options) *metrics {
	return &metrics{
		cluster: newClusterMetrics(opts),
		app:     newAppMetrics(opts),
		system:  newSystemMetrics(opts),
		db:      NewDBMetrics(),
		opts:    opts,
		Log:     wklog.NewWKLog("Metrics"),
	}
}

func (d *metrics) System() ISystemMetrics {
	return d.system
}

func (d *metrics) App() IAppMetrics {
	return d.app
}

func (d *metrics) Cluster() IClusterMetrics {
	return d.cluster
}

func (d *metrics) DB() IDBMetrics {
	return d.db
}

func (d *metrics) Route(r *wkhttp.WKHttp) {

	r.GET("/metrics/app", d.appMetrics)
	r.GET("/metrics/cluster", d.clusterMetrics)
	r.GET("/metrics/system", d.systemMetrics)
}

func (d *metrics) appMetrics(c *wkhttp.Context) {
	latestStr := c.Query("latest")
	nodeIdStr := c.Query("node_id")
	var latest int64
	if latestStr != "" {
		latest, _ = strconv.ParseInt(latestStr, 10, 64)
	}
	var latestTime time.Time
	if latest == 0 {
		latestTime = time.Now().Add(-time.Minute * 5)
	} else {
		latestTime = time.Now().Add(-time.Second * time.Duration(latest))
	}
	var nodeId uint64
	if nodeIdStr != "" {
		nodeId, _ = strconv.ParseUint(nodeIdStr, 10, 64)
	}
	filterId := ""
	if nodeId != 0 {
		filterId = strconv.FormatUint(nodeId, 10)
	}

	sp := latest / 30
	if sp < 1 {
		sp = 1
	}

	rg := v1.Range{
		Start: latestTime,
		End:   time.Now(),
		Step:  time.Duration(sp) * time.Second,
	}

	var resps = make([]*appMetricsResp, 0)

	d.requestAndFillAppMetrics("app_conn_count_total", filterId, rg, false, &resps)

	d.requestAndFillAppMetrics("app_online_user_count_total", filterId, rg, false, &resps)

	d.requestAndFillAppMetrics("app_online_device_count_total", filterId, rg, false, &resps)

	d.requestAndFillAppMetrics("app_send_packet_count_total", filterId, rg, true, &resps)

	d.requestAndFillAppMetrics("app_send_packet_bytes_total", filterId, rg, true, &resps)

	d.requestAndFillAppMetrics("app_sendack_packet_count_total", filterId, rg, true, &resps)

	d.requestAndFillAppMetrics("app_sendack_packet_bytes_total", filterId, rg, true, &resps)

	d.requestAndFillAppMetrics("app_recv_packet_count_total", filterId, rg, true, &resps)

	d.requestAndFillAppMetrics("app_recv_packet_bytes_total", filterId, rg, true, &resps)

	d.requestAndFillAppMetrics("app_recvack_packet_count_total", filterId, rg, true, &resps)

	d.requestAndFillAppMetrics("app_recvack_packet_bytes_total", filterId, rg, true, &resps)

	d.requestAndFillAppMetrics("app_conn_packet_count_total", filterId, rg, true, &resps)

	d.requestAndFillAppMetrics("app_conn_packet_bytes_total", filterId, rg, true, &resps)

	d.requestAndFillAppMetrics("app_connack_packet_count_total", filterId, rg, true, &resps)

	d.requestAndFillAppMetrics("app_connack_packet_bytes_total", filterId, rg, true, &resps)

	d.requestAndFillAppMetrics("app_ping_count_total", filterId, rg, true, &resps)

	d.requestAndFillAppMetrics("app_ping_bytes_total", filterId, rg, true, &resps)

	d.requestAndFillAppMetrics("app_pong_count_total", filterId, rg, true, &resps)

	d.requestAndFillAppMetrics("app_pong_bytes_total", filterId, rg, true, &resps)

	sort.Slice(resps, func(i, j int) bool {
		return resps[i].Timestamp < resps[j].Timestamp
	})

	if resps == nil {
		resps = make([]*appMetricsResp, 0)
	}
	c.JSON(http.StatusOK, resps)
}

func (d *metrics) clusterMetrics(c *wkhttp.Context) {
	latestStr := c.Query("latest")
	var latest int64
	if latestStr != "" {
		latest, _ = strconv.ParseInt(latestStr, 10, 64)
	}
	var latestTime time.Time
	if latest == 0 {
		latestTime = time.Now().Add(-time.Minute * 5)
	} else {
		latestTime = time.Now().Add(-time.Second * time.Duration(latest))
	}

	sp := latest / 30
	if sp < 1 {
		sp = 1
	}

	rg := v1.Range{
		Start: latestTime,
		End:   time.Now(),
		Step:  time.Duration(sp) * time.Second,
	}

	var resps = make([]*clusterMetricsResp, 0)
	d.requestAndFillClusterMetrics("cluster_msg_incoming_count_total", rg, true, &resps)
	d.requestAndFillClusterMetrics("cluster_msg_outgoing_count_total", rg, true, &resps)
	d.requestAndFillClusterMetrics("cluster_msg_incoming_bytes_total", rg, true, &resps)
	d.requestAndFillClusterMetrics("cluster_msg_outgoing_bytes_total", rg, true, &resps)

	d.requestAndFillClusterMetrics("cluster_channel_msg_incoming_count_total", rg, true, &resps)
	d.requestAndFillClusterMetrics("cluster_channel_msg_outgoing_count_total", rg, true, &resps)
	d.requestAndFillClusterMetrics("cluster_channel_msg_incoming_bytes_total", rg, true, &resps)
	d.requestAndFillClusterMetrics("cluster_channel_msg_outgoing_bytes_total", rg, true, &resps)

	d.requestAndFillClusterMetrics("cluster_sendpacket_incoming_bytes_total", rg, true, &resps)
	d.requestAndFillClusterMetrics("cluster_sendpacket_incoming_count_total", rg, true, &resps)
	d.requestAndFillClusterMetrics("cluster_sendpacket_outgoing_bytes_total", rg, true, &resps)
	d.requestAndFillClusterMetrics("cluster_sendpacket_outgoing_count_total", rg, true, &resps)

	d.requestAndFillClusterMetrics("cluster_msg_sync_incoming_bytes_total", rg, true, &resps)
	d.requestAndFillClusterMetrics("cluster_msg_sync_outgoing_bytes_total", rg, true, &resps)
	d.requestAndFillClusterMetrics("cluster_msg_sync_incoming_count_total", rg, true, &resps)
	d.requestAndFillClusterMetrics("cluster_msg_sync_outgoing_count_total", rg, true, &resps)

	d.requestAndFillClusterMetrics("cluster_channel_msg_sync_incoming_count_total", rg, true, &resps)
	d.requestAndFillClusterMetrics("cluster_channel_msg_sync_outgoing_count_total", rg, true, &resps)
	d.requestAndFillClusterMetrics("cluster_channel_msg_sync_incoming_bytes_total", rg, true, &resps)
	d.requestAndFillClusterMetrics("cluster_channel_msg_sync_outgoing_bytes_total", rg, true, &resps)

	d.requestAndFillClusterMetrics("cluster_channel_active_count", rg, false, &resps)

	d.requestAndFillClusterMetrics("cluster_channel_propose_count_total", rg, true, &resps)
	d.requestAndFillClusterMetrics("cluster_channel_propose_failed_count_total", rg, true, &resps)
	d.requestAndFillClusterMetrics("cluster_channel_propose_latency_over_500ms_total", rg, true, &resps)
	d.requestAndFillClusterMetrics("cluster_channel_propose_latency_under_500ms_total", rg, true, &resps)

	d.requestAndFillClusterMetrics("cluster_msg_ping_incoming_bytes_total", rg, true, &resps)
	d.requestAndFillClusterMetrics("cluster_msg_ping_outgoing_bytes_total", rg, true, &resps)
	d.requestAndFillClusterMetrics("cluster_msg_ping_incoming_count_total", rg, true, &resps)
	d.requestAndFillClusterMetrics("cluster_msg_ping_outgoing_count_total", rg, true, &resps)

	sort.Slice(resps, func(i, j int) bool {
		return resps[i].Timestamp < resps[j].Timestamp
	})

	if resps == nil {
		resps = make([]*clusterMetricsResp, 0)
	}
	c.JSON(http.StatusOK, resps)
}

func (d *metrics) systemMetrics(c *wkhttp.Context) {
	latestStr := c.Query("latest")
	var latest int64
	if latestStr != "" {
		latest, _ = strconv.ParseInt(latestStr, 10, 64)
	}
	var latestTime time.Time
	if latest == 0 {
		latestTime = time.Now().Add(-time.Minute * 5)
	} else {
		latestTime = time.Now().Add(-time.Second * time.Duration(latest))
	}

	sp := latest / 30
	if sp < 1 {
		sp = 1
	}

	rg := v1.Range{
		Start: latestTime,
		End:   time.Now(),
		Step:  time.Second * time.Duration(sp),
	}

	var resps = make([]*systemMetricsResp, 0)

	d.requestAndFillSystemMetrics("system_intranet_incoming_bytes_total", rg, true, &resps)

	d.requestAndFillSystemMetrics("system_intranet_outgoing_bytes_total", rg, true, &resps)

	d.requestAndFillSystemMetrics("system_extranet_incoming_bytes_total", rg, true, &resps)

	d.requestAndFillSystemMetrics("system_extranet_outgoing_bytes_total", rg, true, &resps)

	d.requestAndFillSystemMetrics("go_memstats_alloc_bytes", rg, false, &resps)

	d.requestAndFillSystemMetrics("go_goroutines", rg, false, &resps)

	d.requestAndFillSystemMetrics("go_gc_duration_seconds_count", rg, false, &resps)

	d.requestAndFillSystemMetrics("system_cpu_percent_total", rg, false, &resps)
	d.requestAndFillSystemMetrics("node_filefd_allocated", rg, false, &resps)

	sort.Slice(resps, func(i, j int) bool {
		return resps[i].Timestamp < resps[j].Timestamp
	})

	if resps == nil {
		resps = make([]*systemMetricsResp, 0)
	}
	c.JSON(http.StatusOK, resps)
}

func (d *metrics) requestAndFillAppMetrics(label string, filterId string, rg v1.Range, rate bool, resps *[]*appMetricsResp) {
	where := getLabelByFilterId(label, filterId)
	if rate {
		where = getRateByLabelFilterId(label, filterId)
	}
	countValue, err := d.opts.requestPrometheus(where, rg)
	if err != nil {
		d.Warn("request failed", zap.String("label", label), zap.Error(err))
	}
	if countValue != nil {
		valueMatrix, ok := countValue.(model.Matrix)
		if ok && len(valueMatrix) > 0 {
			values := valueMatrix[0].Values
			for _, v := range values {
				var newResp *appMetricsResp
				for _, resp := range *resps {
					if resp.Timestamp == v.Timestamp.Unix() {
						newResp = resp
						break
					}
				}
				if newResp == nil {
					newResp = &appMetricsResp{
						Timestamp: v.Timestamp.Time().Unix(),
					}
					*resps = append(*resps, newResp)
				}

				d.fillAppMetricsRespBySamplePair(newResp, label, v)
			}
		}
	}
}

func (d *metrics) requestAndFillClusterMetrics(label string, rg v1.Range, rate bool, resps *[]*clusterMetricsResp) {
	query := `rate(` + label + `[1m])`
	if !rate {
		query = label
	}
	countValue, err := d.opts.requestPrometheus(query, rg)
	if err != nil {
		d.Warn("request failed", zap.String("label", label), zap.Error(err))
	}
	if countValue != nil {
		valueMatrix, ok := countValue.(model.Matrix)
		if ok && len(valueMatrix) > 0 {
			d.fillClusterMetricsRespByMatrix(resps, label, valueMatrix)
		}
	}
}

func (d *metrics) requestAndFillSystemMetrics(label string, rg v1.Range, rate bool, resps *[]*systemMetricsResp) {
	query := `rate(` + label + `[1m])`
	if !rate {
		query = label
	}
	if label == "go_gc_duration_seconds_count" {
		query = `increase(` + label + `[1m])`
	}
	countValue, err := d.opts.requestPrometheus(query, rg)
	if err != nil {
		d.Warn("request failed", zap.String("label", label), zap.Error(err))
	}
	if countValue != nil {
		valueMatrix, ok := countValue.(model.Matrix)
		if ok && len(valueMatrix) > 0 {
			d.fillSystemMetricsRespByMatrix(resps, label, valueMatrix)
		}
	}
}

func (d *metrics) getAppMetricsRespBySamplePair(label string, pair model.SamplePair) (*appMetricsResp, error) {
	resp := &appMetricsResp{
		ConnCount: int64(pair.Value),
		Timestamp: pair.Timestamp.Time().Unix(),
	}

	return resp, nil
}

func (d *metrics) fillAppMetricsRespBySamplePair(resp *appMetricsResp, label string, pair model.SamplePair) {
	switch label {
	case "app_conn_count_total":
		resp.ConnCount = int64(pair.Value)
	case "app_online_user_count_total":
		resp.OnlineUserCount = int64(pair.Value)
	case "app_online_device_count_total":
		resp.OnlineDeviceCount = int64(pair.Value)
	case "app_send_packet_count_total":
		resp.SendPacketCountRate = float64(pair.Value)
	case "app_send_packet_bytes_total":
		resp.SendPacketBytesRate = float64(pair.Value)
	case "app_sendack_packet_count_total":
		resp.SendackPacketCountRate = float64(pair.Value)
	case "app_sendack_packet_bytes_total":
		resp.SendackPacketBytesRate = float64(pair.Value)
	case "app_recv_packet_count_total":
		resp.RecvPacketCountRate = float64(pair.Value)
	case "app_recv_packet_bytes_total":
		resp.RecvPacketBytesRate = float64(pair.Value)
	case "app_recvack_packet_count_total":
		resp.RecvackPacketCountRate = float64(pair.Value)
	case "app_recvack_packet_bytes_total":
		resp.RecvackPacketBytesRate = float64(pair.Value)
	case "app_conn_packet_count_total":
		resp.ConnPacketCountRate = float64(pair.Value)
	case "app_conn_packet_bytes_total":
		resp.ConnPacketBytesRate = float64(pair.Value)
	case "app_connack_packet_count_total":
		resp.ConnackPacketCountRate = float64(pair.Value)
	case "app_connack_packet_bytes_total":
		resp.ConnackPacketBytesRate = float64(pair.Value)
	case "app_ping_count_total":
		resp.PingPacketCountRate = float64(pair.Value)
	case "app_ping_bytes_total":
		resp.PingPacketBytesRate = float64(pair.Value)
	case "app_pong_count_total":
		resp.PongPacketCountRate = float64(pair.Value)
	case "app_pong_bytes_total":
		resp.PongPacketBytesRate = float64(pair.Value)

	}
}

func (d *metrics) fillClusterMetricsRespByMatrix(resps *[]*clusterMetricsResp, label string, matrix model.Matrix) {

	for _, v := range matrix {
		nodeId := v.Metric["id"]
		if nodeId == "" {
			continue
		}
		for _, pair := range v.Values {
			var newResp *clusterMetricsResp
			for _, resp := range *resps {
				if resp.Timestamp == pair.Timestamp.Time().Unix() {
					newResp = resp
					break
				}
			}
			if newResp == nil {
				newResp = &clusterMetricsResp{
					Timestamp: pair.Timestamp.Time().Unix(),
				}
				*resps = append(*resps, newResp)
			}
			d.fillClusterMetricsRespBySamplePair(newResp, label, string(nodeId), pair)
		}
	}
}

func (d *metrics) fillSystemMetricsRespByMatrix(resps *[]*systemMetricsResp, label string, matrix model.Matrix) {

	for _, v := range matrix {
		for _, pair := range v.Values {
			nodeId := v.Metric["id"]
			if label != "node_filefd_allocated" {
				if nodeId == "" {
					continue
				}
			}

			var newResp *systemMetricsResp
			for _, resp := range *resps {
				if resp.Timestamp == pair.Timestamp.Time().Unix() {
					newResp = resp
					break
				}
			}
			if newResp == nil {
				newResp = &systemMetricsResp{
					Timestamp: pair.Timestamp.Time().Unix(),
				}
				*resps = append(*resps, newResp)
			}
			d.fillSystemMetricsRespBySamplePair(newResp, label, string(nodeId), pair)
		}
	}
}

func (d *metrics) fillClusterMetricsRespBySamplePair(resp *clusterMetricsResp, label string, nodeId string, pair model.SamplePair) {
	switch label {
	case "cluster_msg_incoming_count_total":
		resp.MsgIncomingCountRate = append(resp.MsgIncomingCountRate, &labelValue{Label: nodeId, Value: float64(pair.Value)})
	case "cluster_msg_outgoing_count_total":
		resp.MsgOutgoingCountRate = append(resp.MsgOutgoingCountRate, &labelValue{Label: nodeId, Value: float64(pair.Value)})
	case "cluster_msg_incoming_bytes_total":
		resp.MsgIncomingBytesRate = append(resp.MsgIncomingBytesRate, &labelValue{Label: nodeId, Value: float64(pair.Value)})
	case "cluster_msg_outgoing_bytes_total":
		resp.MsgOutgoingBytesRate = append(resp.MsgOutgoingBytesRate, &labelValue{Label: nodeId, Value: float64(pair.Value)})
	case "cluster_channel_msg_incoming_count_total":
		resp.ChannelMsgIncomingCountRate = append(resp.ChannelMsgIncomingCountRate, &labelValue{Label: nodeId, Value: float64(pair.Value)})
	case "cluster_channel_msg_outgoing_count_total":
		resp.ChannelMsgOutgoingCountRate = append(resp.ChannelMsgOutgoingCountRate, &labelValue{Label: nodeId, Value: float64(pair.Value)})
	case "cluster_channel_msg_incoming_bytes_total":
		resp.ChannelMsgIncomingBytesRate = append(resp.ChannelMsgIncomingBytesRate, &labelValue{Label: nodeId, Value: float64(pair.Value)})
	case "cluster_channel_msg_outgoing_bytes_total":
		resp.ChannelMsgOutgoingBytesRate = append(resp.ChannelMsgOutgoingBytesRate, &labelValue{Label: nodeId, Value: float64(pair.Value)})
	case "cluster_sendpacket_incoming_bytes_total":
		resp.SendPacketIncomingBytesRate = append(resp.SendPacketIncomingBytesRate, &labelValue{Label: nodeId, Value: float64(pair.Value)})
	case "cluster_sendpacket_incoming_count_total":
		resp.SendPacketIncomingCountRate = append(resp.SendPacketIncomingCountRate, &labelValue{Label: nodeId, Value: float64(pair.Value)})
	case "cluster_sendpacket_outgoing_bytes_total":
		resp.SendPacketOutgoingBytesRate = append(resp.SendPacketOutgoingBytesRate, &labelValue{Label: nodeId, Value: float64(pair.Value)})
	case "cluster_sendpacket_outgoing_count_total":
		resp.SendPacketOutgoingCountRate = append(resp.SendPacketOutgoingCountRate, &labelValue{Label: nodeId, Value: float64(pair.Value)})
	case "cluster_msg_sync_incoming_bytes_total":
		resp.MsgSyncIncomingBytesRate = append(resp.MsgSyncIncomingBytesRate, &labelValue{Label: nodeId, Value: float64(pair.Value)})
	case "cluster_msg_sync_outgoing_bytes_total":
		resp.MsgSyncOutgoingBytesRate = append(resp.MsgSyncOutgoingBytesRate, &labelValue{Label: nodeId, Value: float64(pair.Value)})
	case "cluster_msg_sync_incoming_count_total":
		resp.MsgSyncIncomingCountRate = append(resp.MsgSyncIncomingCountRate, &labelValue{Label: nodeId, Value: float64(pair.Value)})
	case "cluster_msg_sync_outgoing_count_total":
		resp.MsgSyncOutgoingCountRate = append(resp.MsgSyncOutgoingCountRate, &labelValue{Label: nodeId, Value: float64(pair.Value)})
	case "cluster_channel_active_count":
		resp.ChannelActiveCount = append(resp.ChannelActiveCount, &labelValue{Label: nodeId, Value: float64(pair.Value)})
	case "cluster_channel_propose_count_total":
		resp.ChannelProposeCountRate = append(resp.ChannelProposeCountRate, &labelValue{Label: nodeId, Value: float64(pair.Value)})
	case "cluster_channel_propose_failed_count_total":
		resp.ChannelProposeFailedCountRate = append(resp.ChannelProposeFailedCountRate, &labelValue{Label: nodeId, Value: float64(pair.Value)})
	case "cluster_channel_propose_latency_over_500ms_total":
		resp.ChannelProposeLatencyOver500msRate = append(resp.ChannelProposeLatencyOver500msRate, &labelValue{Label: nodeId, Value: float64(pair.Value)})
	case "cluster_channel_propose_latency_under_500ms_total":
		resp.ChannelProposeLatencyUnder500msRate = append(resp.ChannelProposeLatencyUnder500msRate, &labelValue{Label: nodeId, Value: float64(pair.Value)})
	case "cluster_msg_ping_incoming_bytes_total":
		resp.MsgPingIncomingBytesRate = append(resp.MsgPingIncomingBytesRate, &labelValue{Label: nodeId, Value: float64(pair.Value)})
	case "cluster_msg_ping_outgoing_bytes_total":
		resp.MsgPingOutgoingBytesRate = append(resp.MsgPingOutgoingBytesRate, &labelValue{Label: nodeId, Value: float64(pair.Value)})
	case "cluster_msg_ping_incoming_count_total":
		resp.MsgPingIncomingCountRate = append(resp.MsgPingIncomingCountRate, &labelValue{Label: nodeId, Value: float64(pair.Value)})
	case "cluster_msg_ping_outgoing_count_total":
		resp.MsgPingOutgoingCountRate = append(resp.MsgPingOutgoingCountRate, &labelValue{Label: nodeId, Value: float64(pair.Value)})

	}
}

func (d *metrics) fillSystemMetricsRespBySamplePair(resp *systemMetricsResp, label string, nodeId string, pair model.SamplePair) {
	switch label {
	case "system_intranet_incoming_bytes_total":
		resp.IntranetIncomingBytesRate = append(resp.IntranetIncomingBytesRate, &labelValue{Label: nodeId, Value: float64(pair.Value)})
	case "system_intranet_outgoing_bytes_total":
		resp.IntranetOutgoingBytesRate = append(resp.IntranetOutgoingBytesRate, &labelValue{Label: nodeId, Value: float64(pair.Value)})
	case "system_extranet_incoming_bytes_total":
		resp.ExtranetIncomingBytesRate = append(resp.ExtranetIncomingBytesRate, &labelValue{Label: nodeId, Value: float64(pair.Value)})
	case "system_extranet_outgoing_bytes_total":
		resp.ExtranetOutgoingBytesRate = append(resp.ExtranetOutgoingBytesRate, &labelValue{Label: nodeId, Value: float64(pair.Value)})
	case "go_memstats_alloc_bytes":
		resp.MemstatsAllocBytes = append(resp.MemstatsAllocBytes, &labelValue{Label: nodeId, Value: float64(pair.Value)})
	case "go_goroutines":
		resp.Goroutines = append(resp.Goroutines, &labelValue{Label: nodeId, Value: float64(pair.Value)})
	case "go_gc_duration_seconds_count":
		resp.GCDurationSecondsCount = append(resp.GCDurationSecondsCount, &labelValue{Label: nodeId, Value: float64(pair.Value)})
	case "system_cpu_percent_total":
		resp.CpuPercent = append(resp.CpuPercent, &labelValue{Label: nodeId, Value: float64(pair.Value)})
	case "node_filefd_allocated":
		resp.FilefdAllocated = int64(pair.Value)
	}

}

func getLabelByFilterId(label string, filterId string) string {
	if filterId != "" {
		return label + `{id="` + filterId + `"}`
	}
	return `sum(` + label + `)`
}

func getRateByLabelFilterId(label string, filterId string) string {
	if filterId != "" {
		return `rate(` + label + `{id="` + filterId + `"}[1m])`
	}
	return `sum(rate(` + label + `[1m]))`
}

type appMetricsResp struct {
	ConnCount         int64 `json:"conn_count"`
	OnlineUserCount   int64 `json:"online_user_count"`
	OnlineDeviceCount int64 `json:"online_device_count"`

	SendPacketCountRate    float64 `json:"send_packet_count_rate"`
	SendPacketBytesRate    float64 `json:"send_packet_bytes_rate"`
	SendackPacketCountRate float64 `json:"sendack_packet_count_rate"`
	SendackPacketBytesRate float64 `json:"sendack_packet_bytes_rate"`

	RecvPacketCountRate    float64 `json:"recv_packet_count_rate"`
	RecvPacketBytesRate    float64 `json:"recv_packet_bytes_rate"`
	RecvackPacketCountRate float64 `json:"recvack_packet_count_rate"`
	RecvackPacketBytesRate float64 `json:"recvack_packet_bytes_rate"`

	ConnPacketCountRate    float64 `json:"conn_packet_count_rate"`
	ConnPacketBytesRate    float64 `json:"conn_packet_bytes_rate"`
	ConnackPacketCountRate float64 `json:"connack_packet_count_rate"`
	ConnackPacketBytesRate float64 `json:"connack_packet_bytes_rate"`

	PingPacketCountRate float64 `json:"ping_packet_count_rate"`
	PingPacketBytesRate float64 `json:"ping_packet_bytes_rate"`
	PongPacketCountRate float64 `json:"pong_packet_count_rate"`
	PongPacketBytesRate float64 `json:"pong_packet_bytes_rate"`

	Timestamp int64 `json:"timestamp"`
}

type clusterMetricsResp struct {
	ChannelProposeCountRate             []*labelValue `json:"channel_propose_count_rate"`
	ChannelProposeFailedCountRate       []*labelValue `json:"channel_propose_failed_count_rate"`
	ChannelProposeLatencyOver500msRate  []*labelValue `json:"channel_propose_latency_over_500ms_rate"`
	ChannelProposeLatencyUnder500msRate []*labelValue `json:"channel_propose_latency_under_500ms_rate"`

	MsgPingIncomingBytesRate []*labelValue `json:"msg_ping_incoming_bytes_rate"`
	MsgPingOutgoingBytesRate []*labelValue `json:"msg_ping_outgoing_bytes_rate"`
	MsgPingIncomingCountRate []*labelValue `json:"msg_ping_incoming_count_rate"`
	MsgPingOutgoingCountRate []*labelValue `json:"msg_ping_outgoing_count_rate"`

	MsgIncomingCountRate []*labelValue `json:"msg_incoming_count_rate"`
	MsgOutgoingCountRate []*labelValue `json:"msg_outgoing_count_rate"`
	MsgIncomingBytesRate []*labelValue `json:"msg_incoming_bytes_rate"`
	MsgOutgoingBytesRate []*labelValue `json:"msg_outgoing_bytes_rate"`

	ChannelMsgIncomingCountRate []*labelValue `json:"channel_msg_incoming_count_rate"`
	ChannelMsgOutgoingCountRate []*labelValue `json:"channel_msg_outgoing_count_rate"`
	ChannelMsgIncomingBytesRate []*labelValue `json:"channel_msg_incoming_bytes_rate"`
	ChannelMsgOutgoingBytesRate []*labelValue `json:"channel_msg_outgoing_bytes_rate"`

	SendPacketIncomingCountRate []*labelValue `json:"sendpacket_incoming_count_rate"`
	SendPacketIncomingBytesRate []*labelValue `json:"sendpacket_incoming_bytes_rate"`
	SendPacketOutgoingBytesRate []*labelValue `json:"sendpacket_outgoing_bytes_rate"`
	SendPacketOutgoingCountRate []*labelValue `json:"sendpacket_outgoing_count_rate"`

	MsgSyncIncomingBytesRate []*labelValue `json:"msg_sync_incoming_bytes_rate"`
	MsgSyncOutgoingBytesRate []*labelValue `json:"msg_sync_outgoing_bytes_rate"`
	MsgSyncIncomingCountRate []*labelValue `json:"msg_sync_incoming_count_rate"`
	MsgSyncOutgoingCountRate []*labelValue `json:"msg_sync_outgoing_count_rate"`

	ChannelActiveCount []*labelValue `json:"channel_active_count"`

	Timestamp int64 `json:"timestamp"`
}

type systemMetricsResp struct {
	IntranetIncomingBytesRate []*labelValue `json:"intranet_incoming_bytes_rate"`
	IntranetOutgoingBytesRate []*labelValue `json:"intranet_outgoing_bytes_rate"`

	ExtranetIncomingBytesRate []*labelValue `json:"extranet_incoming_bytes_rate"`
	ExtranetOutgoingBytesRate []*labelValue `json:"extranet_outgoing_bytes_rate"`

	MemstatsAllocBytes     []*labelValue `json:"memstats_alloc_bytes"`
	Goroutines             []*labelValue `json:"goroutines"`
	GCDurationSecondsCount []*labelValue `json:"gc_duration_seconds_count"`

	CpuPercent      []*labelValue `json:"cpu_percent"`
	FilefdAllocated int64         `json:"filefd_allocated"`
	Timestamp       int64         `json:"timestamp"`
}

type labelValue struct {
	Label string  `json:"label"`
	Value float64 `json:"value"`
}
