package collectorsloading

import (
	"exp_poc/commandline"
	"strings"

	"github.com/prometheus-community/windows_exporter/collector"
	"github.com/prometheus-community/windows_exporter/log"
)

func GettingCollectors(list string) (map[string]collector.Collector, error) {
	collectors := map[string]collector.Collector{}
	enabled := func() []string {
		expanded := strings.Replace(list, commandline.DefaultCollectorsPlaceholder, commandline.DefaultCollectors, -1)
		separated := strings.Split(expanded, ",")
		unique := map[string]bool{}
		for _, s := range separated {
			if s != "" {
				unique[s] = true
			}
		}
		result := make([]string, 0, len(unique))
		for s := range unique {
			result = append(result, s)
		}
		return result
	}

	for _, name := range enabled() {
		c, err := collector.Build(name)
		if err != nil {
			log.Error(err)
		}
		collectors[name] = c
	}

	return collectors, nil
}

func EnabledCollectors(collectors map[string]collector.Collector) {
	ret := make([]string, 0, len(collectors))
	for key := range collectors {
		ret = append(ret, key)
	}
	log.Infof("Enabled collectors: %v",
		strings.Join(ret, ", "))
}
