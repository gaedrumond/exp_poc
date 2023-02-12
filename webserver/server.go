package webserver

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus-community/windows_exporter/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/version"
)

const defaultTimeout = 10.0

var TimeoutSeconds float64

var err error

func (mh *MetricsHandler) ServerHTTP(writer http.ResponseWriter, request *http.Request) {
	getPrometheusTimeout(request)

	TimeoutSeconds := setTimeout(mh)

	reg := prometheusRegistry(mh, TimeoutSeconds, writer, request)

	h := promhttp.HandlerFor(reg, promhttp.HandlerOpts{})
	h.ServeHTTP(writer, request)

}

func prometheusRegistry(mh *MetricsHandler, TimeoutSeconds float64, writer http.ResponseWriter, request *http.Request) *prometheus.Registry {
	reg := prometheus.NewRegistry()
	err, wc := mh.CollectorFactory(time.Duration(TimeoutSeconds*float64(time.Second)), request.URL.Query()["collect[]"])
	if err != nil {
		log.Warnln("Couldn't create filtered metrics handler: ", err)
		writer.WriteHeader(http.StatusBadRequest)
		writer.Write([]byte(fmt.Sprintf("Couldn't create filtered metrics handler: %s", err))) //nolint:errcheck
		return nil
	}
	reg.MustRegister(wc)
	reg.MustRegister(
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		collectors.NewGoCollector(),
		version.NewCollector("windows_exporter"),
	)
	return reg
}

func getPrometheusTimeout(request *http.Request) {
	if v := request.Header.Get("X-Prometheus-Scrape-Timeout-Seconds"); v != "" {
		TimeoutSeconds, err = strconv.ParseFloat(v, 64)
		if err != nil {
			log.Warnf("Couldn't parse X-Prometheus-Scrape-Timeout-Seconds: %q. Defaulting timeout to %f", v, defaultTimeout)
		}
	}
}

func setTimeout(mh *MetricsHandler) float64 {
	if TimeoutSeconds == 0 {
		TimeoutSeconds = defaultTimeout
	}
	TimeoutSeconds = TimeoutSeconds - mh.TimeoutMargin
	return TimeoutSeconds
}
