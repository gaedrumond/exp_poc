package collectorbase

import (
	"strconv"
	"strings"
)

func registerCollector(name string, builder collectorBuilder, perfCounterNames ...string) {
	builders[name] = builder
	addPerfCounterDependencies(name, perfCounterNames)
}

func MapCounterToIndex(name string) string {
	return strconv.Itoa(int(nametable.LookupIndex(name)))
}

func addPerfCounterDependencies(name string, perfCounterNames []string) {
	perfIndicies := make([]string, 0, len(perfCounterNames))
	for _, cn := range perfCounterNames {
		perfIndicies = append(perfIndicies, MapCounterToIndex(cn))
	}
	perfCounterDependencies[name] = strings.Join(perfIndicies, " ")
}
