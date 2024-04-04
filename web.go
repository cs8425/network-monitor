package main

import (
	"math"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	rttBucket = []float64{
		0.01,
		0.1,
		1.0,
		2.5,
		5.0,
		7.5,
		10.0,
		12.5,
		15.0,
		17.5,
		20.0,

		25.0,
		30.0,
		35.0,
		40.0,
		45.0,
		50.0,

		60.0,
		70.0,
		80.0,
		90.0,
		100.0,
		110.0,
		120.0,
		130.0,
		140.0,
		150.0,
		160.0,
		170.0,

		185.0,
		200.0,
		215.0,
		230.0,

		250.0,
		275.0,
		300.0,
		325.0,
		350.0,
		375.0,
		400.0,

		450.0,
		500.0,
		550.0,
		600.0,
		650.0,
		700.0,
		750.0,
		800.0,
		850.0,
		900.0,
		975.0,
	}

	// promBuildCollector = collectors.NewBuildInfoCollector()
	// promGoCollector    = collectors.NewGoCollector()
	promProcCollector = collectors.NewProcessCollector(collectors.ProcessCollectorOpts{})

	promPacketTransmitted = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "packet_transmitted",
		Help: "A counter of packet transmitted.",
	}, []string{
		"dst",
	})
	promPacketLoss = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "packet_loss",
		Help: "A counter of packet loss.",
	}, []string{
		"dst",
	})
	// promPacketReceived = prometheus.NewCounterVec(prometheus.CounterOpts{
	// 	Name: "packet_received",
	// 	Help: "A counter of packet received.",
	// }, []string{
	// 	"dst",
	// })
	promPacketRoundTripTime = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "tcp_rtt_milliseconds",
		Help: "A gauge of round trip time in milliseconds.",
	}, []string{
		"dst",
	})
	promRoundTripTimeHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "tcp_rtt_distribution",
		Help:    "A histogram of round trip time in milliseconds.",
		Buckets: rttBucket, //prometheus.ExponentialBucketsRange(1, SYN_INIT_TIMEOUT, 25),
	}, []string{
		"dst",
	})
)

func toMilliseconds(d time.Duration) float64 {
	ms := d / time.Millisecond
	nms := d % time.Millisecond
	return float64(ms) + float64(nms)/1e6
}

func appendResult(dst string, dt time.Duration, err error) {
	promPacketTransmitted.With(prometheus.Labels{"dst": dst}).Inc()
	promPacketLoss.With(prometheus.Labels{"dst": dst}).Add(0)

	ms := toMilliseconds(dt)
	if err != nil {
		// skip any other error
		if !os.IsTimeout(err) {
			return
		}
		// only count packet lost
		promPacketLoss.With(prometheus.Labels{"dst": dst}).Inc()
		// ms = math.NaN()
		promPacketRoundTripTime.With(prometheus.Labels{"dst": dst}).Set(math.NaN())
		// promRoundTripTimeHistogram.With(prometheus.Labels{"dst": dst}).Observe(math.NaN())
		// promRoundTripTimeHistogram.With(prometheus.Labels{"dst": dst}).Observe(math.Inf(1))
		return
	}
	promPacketRoundTripTime.With(prometheus.Labels{"dst": dst}).Set(ms)
	promRoundTripTimeHistogram.With(prometheus.Labels{"dst": dst}).Observe(ms)
}

func StatsHandler() http.Handler {
	// Create non-global registry.
	reg := prometheus.NewRegistry()

	// Add go runtime metrics and process collectors.
	reg.MustRegister(
		// promBuildCollector,
		// promGoCollector,
		promProcCollector,

		promPacketTransmitted,
		promPacketLoss,
		promPacketRoundTripTime,
		promRoundTripTimeHistogram,
	)
	return promhttp.HandlerFor(reg, promhttp.HandlerOpts{
		EnableOpenMetrics: true,
	})
}

func WebStart(addr string) {
	// Vln(0, prometheus.ExponentialBucketsRange(1, SYN_INIT_TIMEOUT, 25))
	// Vln(0, prometheus.LinearBuckets(0.0, 40, 25))

	if addr == "" {
		return
	}
	Vln(2, "[web][prometheus]Listening:", addr)

	mux := http.NewServeMux()
	//mux.Handle("/", http.FileServer(http.Dir("./www")))
	mux.Handle("/metrics", StatsHandler())

	srv := &http.Server{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 30 * time.Second,
		Addr:         addr,
		Handler:      mux,
	}
	err := srv.ListenAndServe()
	Vln(2, "[web][prometheus]:", err)
}
