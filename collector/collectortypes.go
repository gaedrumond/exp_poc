package collectorbase

import (
	"github.com/leoluk/perflib_exporter/perflib"
	"github.com/prometheus/client_golang/prometheus"
)

type ScrapeContext struct {
	perfObjects map[string]*perflib.PerfObject
}

type Collector interface {
	Collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) (err error)
}

type collectorBuilder func() (Collector, error)

var (
	builders                = make(map[string]collectorBuilder)
	perfCounterDependencies = make(map[string]string)
)

var nametable = perflib.QueryNameTable("Counter 009")
