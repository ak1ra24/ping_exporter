package main

import (
	probing "github.com/ak1ra24/pro-bing"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	namespace = "ping"
)

var (
	labelNames = []string{"target", "ip", "broadcast", "name", "description"}

	sentCountVec = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "sent_count",
			Help:      "The number of send ping packet",
		},
		labelNames,
	)

	recvCountVec = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "recv_count",
			Help:      "The number of receive ping packet",
		},
		labelNames,
	)

	lastRttVec = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "last_rtt",
			Help:      "last rtt",
		},
		labelNames,
	)

	lossLabelNames = []string{"target", "ip", "broadcast", "name", "description", "loss_reason"}
	lossCountVec   = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "loss_count",
			Help:      "The number of loss packet",
		},
		lossLabelNames,
	)
)

type PingCollector struct {
	pingers *[]*PingerConfig
}

func GetBroadcastLabel(broadcast bool) string {
	if broadcast {
		return "true"
	} else {
		return "false"
	}
}

func newPingCollector(pingerConfigs *[]*PingerConfig) *PingCollector {
	for _, pingerConfig := range *pingerConfigs {
		pinger := pingerConfig.pinger
		host := pingerConfig.Host

		pinger.OnSend = func(pkt *probing.Packet) {
			sentCountVec.WithLabelValues(
				pkt.Target,
				pkt.IPAddr.String(),
				GetBroadcastLabel(host.Broadcast),
				host.Name,
				host.Description,
			).Inc()
		}

		pinger.OnRecv = func(pkt *probing.Packet) {
			if pkt.Loss && pkt.Target == pkt.Addr {
				lossCountVec.WithLabelValues(
					pkt.Target,
					pkt.Addr,
					GetBroadcastLabel(host.Broadcast),
					host.Name,
					host.Description,
					pkt.LossReason,
				).Inc()
			} else {
				if pkt.Addr == pkt.IPAddr.String() {
					recvCountVec.WithLabelValues(
						pkt.Target,
						pkt.IPAddr.String(),
						GetBroadcastLabel(host.Broadcast),
						host.Name,
						host.Description,
					).Inc()
					lastRttVec.WithLabelValues(
						pkt.Target,
						pkt.IPAddr.String(),
						GetBroadcastLabel(host.Broadcast),
						host.Name,
						host.Description,
					).Set(float64(pkt.Rtt.Seconds()))
				}
			}
		}

		pinger.OnDuplicateRecv = func(pkt *probing.Packet) {
			if pkt.Addr == pkt.IPAddr.String() {
				recvCountVec.WithLabelValues(
					pkt.Target,
					pkt.IPAddr.String(),
					GetBroadcastLabel(host.Broadcast),
					host.Name,
					host.Description,
				).Inc()
				lastRttVec.WithLabelValues(
					pkt.Target,
					pkt.IPAddr.String(),
					GetBroadcastLabel(host.Broadcast),
					host.Name,
					host.Description,
				).Set(float64(pkt.Rtt.Seconds()))
			}
		}
	}

	return &PingCollector{
		pingers: pingerConfigs,
	}
}

func (s *PingCollector) Describe(ch chan<- *prometheus.Desc) {
}

func (s *PingCollector) Collect(ch chan<- prometheus.Metric) {
}
