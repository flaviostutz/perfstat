package main

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/flaviostutz/perfstat"
	"github.com/flaviostutz/signalutils"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/shirou/gopsutil/host"
	"github.com/sirupsen/logrus"
)

func startPrometheus(ctx context.Context, opt Option, ps *perfstat.Perfstat) {

	info, _ := host.Info()

	dangerGauge := promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "danger_score",
		Help: "Danger level by type and subsystem",
	}, []string{
		"host",
		"type",
		"group",
	})

	issuesGauge := promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "issue_score",
		Help: "Issue score details",
	}, []string{
		"host",
		"type",
		"group",
		"id",
		"resource_name",
		"resource_property_name",
		"related_resource_name",
	})

	issueResourceGauge := promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "issue_resource_value",
		Help: "Issue resource value",
	}, []string{
		"host",
		"type",
		"group",
		"id",
		"resource_name",
		"resource_property_name",
	})

	//setup prometheus metrics http server
	router := mux.NewRouter()
	router.Handle(opt.promPath, promhttp.Handler())

	listen := fmt.Sprintf("%s:%d", opt.promBindHost, opt.promBindPort)
	listenPort, err := net.Listen("tcp", listen)
	if err != nil {
		panic(err)
	}

	//HTTP SERVER
	go func() {
		fmt.Printf("Starting Prometheus Exporter at http://%s:%d%s\n", opt.promBindHost, opt.promBindPort, opt.promPath)
		http.Serve(listenPort, router)
		defer listenPort.Close()
	}()

	//GENERATE METRICS
	signalutils.StartWorker(ctx, "prom-metrics", func() error {
		logrus.Debugf("Generating Prometheus metrics")

		//DANGER LEVELS
		genMetrics(dangerGauge, info, "", "")
		genMetrics(dangerGauge, info, "bottleneck", "cpu")
		genMetrics(dangerGauge, info, "bottleneck", "mem")
		genMetrics(dangerGauge, info, "bottleneck", "disk")
		genMetrics(dangerGauge, info, "bottleneck", "net")
		genMetrics(dangerGauge, info, "risk", "mem")
		genMetrics(dangerGauge, info, "risk", "disk")
		genMetrics(dangerGauge, info, "risk", "net")

		//DETECTIONS
		dr := ps.TopCriticity(-1, "", "", false)
		for _, d := range dr {
			relName := ""
			if len(d.Related) > 0 {
				relName = d.Related[0].Name
			}
			issuesGauge.WithLabelValues(info.Hostname, d.Typ, groupFromID(d.ID), fmt.Sprintf("%s", d.ID), d.Res.Name, d.Res.PropertyName, relName).Set(d.Score)
			issueResourceGauge.WithLabelValues(info.Hostname, d.Typ, groupFromID(d.ID), fmt.Sprintf("%s", d.ID), d.Res.Name, d.Res.PropertyName).Set(d.Res.PropertyValue)
		}

		return nil
	}, 0.5, 1.0, false)

	select {
	case <-ctx.Done():
		listenPort.Close()
	}
}

func genMetrics(g *prometheus.GaugeVec, info *host.InfoStat, typ string, group string) {
	g.WithLabelValues(info.Hostname, typ, group).Set(ps.Score(typ, fmt.Sprintf("%s.*", group)))
}
