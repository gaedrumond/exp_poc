package metrics

import (
	"exp_poc/metadata"
	"sync"
	"time"

	"github.com/prometheus-community/windows_exporter/collector"
	"github.com/prometheus-community/windows_exporter/log"
	"github.com/prometheus/client_golang/prometheus"
)

type CustomCollector struct {
	MaxScrapeDuration time.Duration
	Collectors        map[string]collector.Collector
}

type collectorOutcome int

const (
	pending collectorOutcome = iota
	success
	failed
)

func (coll CustomCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- scrapeDurationDesc
	ch <- scrapeSuccessDesc
}

func collectorsContext(coll CustomCollector) []string {
	cs := make([]string, 0, len(coll.Collectors))
	for name := range coll.Collectors {
		cs = append(cs, name)
	}
	return cs
}

func metricContext(coll CustomCollector, t time.Time, ch chan<- prometheus.Metric) *collector.ScrapeContext {
	cs := collectorsContext(coll)

	scrapeContext, err := collector.PrepareScrapeContext(cs)
	ch <- prometheus.MustNewConstMetric(
		snapshotDuration,
		prometheus.GaugeValue,
		time.Since(t).Seconds(),
		metadata.GetMetadata().MacAdd,
		metadata.GetMetadata().PCName,
	)
	if err != nil {
		log.Fatal(err)
	}

	return scrapeContext
}

func collectorOutcomeEvaluation(coll CustomCollector) map[string]collectorOutcome {
	collectorOutcome := make(map[string]collectorOutcome)
	for name := range coll.Collectors {
		collectorOutcome[name] = pending
	}

	return collectorOutcome
}

func collecting(name string, c collector.Collector, context *collector.ScrapeContext, bufferChannel chan<- prometheus.Metric) collectorOutcome {
	t := time.Now()
	err := c.Collect(context, bufferChannel)
	duration := time.Since(t).Seconds()
	bufferChannel <- prometheus.MustNewConstMetric(
		scrapeDurationDesc,
		prometheus.GaugeValue,
		duration,
		name,
		metadata.GetMetadata().MacAdd,
		metadata.GetMetadata().PCName,
	)

	if err != nil {
		log.Error(err)
	}

	return success
}

func metricBuffer(coll CustomCollector, ch chan<- prometheus.Metric, scrapeContext *collector.ScrapeContext) (map[string]collectorOutcome, sync.Mutex) {

	wg := sync.WaitGroup{}
	wg.Add(len(coll.Collectors))

	collectorOutcome := collectorOutcomeEvaluation(coll)
	metricBuffer := make(chan prometheus.Metric)

	l := sync.Mutex{}
	finished := false

	go func() {
		for m := range metricBuffer {
			l.Lock()
			if !finished {
				ch <- m
			}
			l.Unlock()
		}
	}()

	for name, c := range coll.Collectors {
		go func(name string, c collector.Collector) {
			defer wg.Done()
			outcome := collecting(name, c, scrapeContext, metricBuffer)
			l.Lock()
			if !finished {
				collectorOutcome[name] = outcome
			}
			l.Unlock()
		}(name, c)
	}

	allDone := make(chan struct{})
	go func() {
		wg.Wait()
		close(allDone)
		close(metricBuffer)
	}()

	select {
	case <-allDone:
	case <-time.After(coll.MaxScrapeDuration):
	}

	l.Lock()
	finished = true

	return collectorOutcome, l
}

func remainingCollectors(collectorOutcome map[string]collectorOutcome, ch chan<- prometheus.Metric) {
	remainingCollectorNames := make([]string, 0)
	for name, outcome := range collectorOutcome {
		var successValue, timeoutValue float64
		if outcome == pending {
			timeoutValue = 1.0
			remainingCollectorNames = append(remainingCollectorNames, name)
		}
		if outcome == success {
			successValue = 1.0
		}

		ch <- prometheus.MustNewConstMetric(
			scrapeSuccessDesc,
			prometheus.GaugeValue,
			successValue,
			name,
			metadata.GetMetadata().MacAdd,
			metadata.GetMetadata().PCName,
		)

		ch <- prometheus.MustNewConstMetric(
			scrapeTimeoutDesc,
			prometheus.GaugeValue,
			timeoutValue,
			name,
			metadata.GetMetadata().MacAdd,
			metadata.GetMetadata().PCName,
		)
	}
	if len(remainingCollectorNames) > 0 {
		log.Warn("tem collector timando out")
	}
}

func (coll CustomCollector) Collect(ch chan<- prometheus.Metric) {
	t := time.Now()

	scrapeContext := metricContext(coll, t, ch)

	collectorOutcome, l := metricBuffer(coll, ch, scrapeContext)

	remainingCollectors(collectorOutcome, ch)

	l.Unlock()
}
