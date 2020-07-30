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
	"github.com/sirupsen/logrus"
)

func startPrometheus(ctx context.Context, opt Option, ps *perfstat.Perfstat) {
	dangerGauge := promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "danger_score",
		Help: "Danger level by type and subsystem",
	}, []string{
		"type",
		"group",
	})

	issuesGauge := promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "issue_score",
		Help: "Issue score details",
	}, []string{
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
		genMetrics(dangerGauge, "", "")
		genMetrics(dangerGauge, "bottleneck", "cpu")
		genMetrics(dangerGauge, "bottleneck", "mem")
		genMetrics(dangerGauge, "bottleneck", "disk")
		genMetrics(dangerGauge, "bottleneck", "net")
		genMetrics(dangerGauge, "risk", "mem")
		genMetrics(dangerGauge, "risk", "disk")
		genMetrics(dangerGauge, "risk", "net")

		//DETECTIONS
		dr := ps.TopCriticity(-1, "", "", false)
		for _, d := range dr {
			relName := ""
			if len(d.Related) > 0 {
				relName = d.Related[0].Name
			}
			issuesGauge.WithLabelValues(d.Typ, groupFromID(d.ID), fmt.Sprintf("%s", d.ID), d.Res.Name, d.Res.PropertyName, relName).Set(d.Score)
			issueResourceGauge.WithLabelValues(d.Typ, groupFromID(d.ID), fmt.Sprintf("%s", d.ID), d.Res.Name, d.Res.PropertyName).Set(d.Res.PropertyValue)
		}

		return nil
	}, 0.5, 1.0, false)

	select {
	case <-ctx.Done():
		listenPort.Close()
	}
}

func genMetrics(g *prometheus.GaugeVec, typ string, group string) {
	g.WithLabelValues(typ, group).Set(ps.Score(typ, fmt.Sprintf("%s.*", group)))
}
