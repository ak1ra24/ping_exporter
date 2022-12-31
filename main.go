package main

import (
	"net/http"
	"os"
	"time"

	"github.com/ak1ra24/ping_exporter/config"
	probing "github.com/ak1ra24/pro-bing"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promlog"
	"github.com/prometheus/common/promlog/flag"
	"github.com/prometheus/common/version"
	"github.com/prometheus/exporter-toolkit/web"
	"github.com/prometheus/exporter-toolkit/web/kingpinflag"
	"golang.org/x/sync/errgroup"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	configFilePath = kingpin.Flag(
		"config",
		"ping exporter config yaml file path.",
	).Default("config.yaml").String()
	metricsPath = kingpin.Flag(
		"web.telemetry-path",
		"Path under which to expose metrics.",
	).Default("/metrics").String()
	interval  = kingpin.Flag("ping.interval", "Ping interval duration").Short('i').Default("1s").Duration()
	logger    log.Logger
	webConfig = kingpinflag.AddFlags(kingpin.CommandLine, ":9375")
	sc        = &config.SafeConfig{
		C: &config.Config{},
	}
)

func init() {
	prometheus.MustRegister(version.NewCollector("ping_exporter"))
}

type PingerConfig struct {
	pinger *probing.Pinger
	Host   config.Host
}

func main() {
	promlogConfig := &promlog.Config{}
	flag.AddFlags(kingpin.CommandLine, promlogConfig)
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	logger = promlog.New(promlogConfig)

	level.Info(logger).Log("msg", "Starting ping_exporter", "version", version.Info())
	pingerConfigs := []*PingerConfig{}

	if err := sc.ReloadConfig(*configFilePath); err != nil {
		level.Error(logger).Log("msg", "Failed to reload config file", "filename", *configFilePath)
		os.Exit(1)
	}

	sc.Lock()
	for _, target := range sc.C.Targets {
		for _, host := range target.Hosts {
			pinger := probing.New(host.IP)
			pinger.Size = target.Size
			pinger.RecordRtts = false
			pinger.SetNetwork(target.Network)
			pinger.Interval = target.Interval
			pinger.SetPrivileged(true)
			pingerConfig := PingerConfig{
				pinger: pinger,
				Host:   host,
			}
			pingerConfigs = append(pingerConfigs, &pingerConfig)
		}
	}
	sc.Unlock()

	if len(pingerConfigs) == 0 {
		level.Error(logger).Log("msg", "no targets specified in config file")
		os.Exit(1)
	}

	splay := time.Duration(interval.Nanoseconds() / int64(len(pingerConfigs)))
	g := new(errgroup.Group)
	for _, pingerConfig := range pingerConfigs {
		g.Go(pingerConfig.pinger.Run)
		time.Sleep(splay)
	}

	prometheus.MustRegister(newPingCollector(&pingerConfigs))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
			<head><title>Ping Exporter</title></head>
			<body>
			<h1>Ping Exporter</h1>
			<p><a href="` + *metricsPath + `">Metrics</a></p>
			</body>
			</html>`))
	})
	http.Handle(*metricsPath, promhttp.Handler())

	server := &http.Server{}
	if err := web.ListenAndServe(server, webConfig, logger); err != nil {
		level.Error(logger).Log("err", err)
		os.Exit(1)
	}

	for _, pingerConfig := range pingerConfigs {
		level.Debug(logger).Log("msg", "pinger stop")
		pingerConfig.pinger.Stop()
	}

	if err := g.Wait(); err != nil {
		level.Error(logger).Log("msg", "pingers failed", "error", err)
		os.Exit(1)
	}
}
