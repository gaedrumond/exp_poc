package main

import (
	"encoding/json"
	"exp_poc/collectorsloading"
	"exp_poc/commandline"
	"exp_poc/metadata"
	"exp_poc/metrics"
	"exp_poc/systemuser"
	"exp_poc/webserver"
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus-community/windows_exporter/collector"
	"github.com/prometheus-community/windows_exporter/initiate"
	"github.com/prometheus-community/windows_exporter/log"
	"github.com/prometheus/common/version"
	"github.com/prometheus/exporter-toolkit/web"
)

type prometheusVersion struct {
	Version   string `json:"version"`
	Revision  string `json:"revision"`
	Branch    string `json:"branch"`
	BuildUser string `json:"buildUser"`
	BuildDate string `json:"buildDate"`
	GoVersion string `json:"goVersion"`
}

func main() {
	fmt.Println(metadata.GetMetadata())
	commandline.AddCommandLine()
	webserver.InitWbem()

	collectors, err := collectorsloading.GettingCollectors(*commandline.EnabledCollectors)
	if err != nil {
		log.Fatalf("Couldn't load collectors: %s", err)
	}

	user := systemuser.GetCurrentUser()

	log.Infof("Running as %v", user.Username)

	systemuser.ValidateUser(user)

	collectorsloading.EnabledCollectors(collectors)

	h := &webserver.MetricsHandler{
		TimeoutMargin: *commandline.TimeoutMargin,
		CollectorFactory: func(timeout time.Duration, requestedCollectors []string) (error, *metrics.CustomCollector) {
			filteredCollectors := make(map[string]collector.Collector)
			// scrape all enabled collectors if no collector is requested
			if len(requestedCollectors) == 0 {
				filteredCollectors = collectors
			}
			for _, name := range requestedCollectors {
				col, exists := collectors[name]
				if !exists {
					return fmt.Errorf("unavailable collector: %s", name), nil
				}
				filteredCollectors[name] = col
			}
			return nil, &metrics.CustomCollector{
				Collectors:        filteredCollectors,
				MaxScrapeDuration: timeout,
			}
		},
	}

	http.HandleFunc(*commandline.MetricsPath, webserver.WebConcurrencyLimit(*commandline.MaxRequest, h.ServerHTTP))
	http.HandleFunc("/health", webserver.WebHealthCheck)
	getPrometheusVersion()
	webDisplay()

	log.Infoln("Starting custom_exporter", version.Info())
	log.Infoln("Build context", version.BuildContext())

	startExporter()
}

func getPrometheusVersion() {
	http.HandleFunc("/version", func(w http.ResponseWriter, r *http.Request) {
		// we can't use "version" directly as it is a package, and not an object that
		// can be serialized.
		err := json.NewEncoder(w).Encode(prometheusVersion{
			Version:   version.Version,
			Revision:  version.Revision,
			Branch:    version.Branch,
			BuildUser: version.BuildUser,
			BuildDate: version.BuildDate,
			GoVersion: version.GoVersion,
		})
		if err != nil {
			http.Error(w, fmt.Sprintf("error encoding JSON: %s", err), http.StatusInternalServerError)
		}
	})
}

func webDisplay() {
	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		_, _ = writer.Write([]byte(`<html>
<head><title>custom_exporter</title></head>
<body>
<h1>windows_exporter</h1>
<p><a href="` + *commandline.MetricsPath + `">Metrics</a></p>
<p><i>` + version.Info() + `</i></p>
</body>
</html>`))
	})
}

func startExporter() {
	go func() {
		server := &http.Server{}
		if err := web.ListenAndServe(server, commandline.WebConfig, log.NewToolkitAdapter()); err != nil {
			log.Fatalf("cannot start windows_exporter: %s", err)
		}
	}()

	for {
		if <-initiate.StopCh {
			log.Info("Shutting down windows_exporter")
			break
		}
	}
}
