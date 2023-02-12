package commandline

import (
	"fmt"
	_ "net/http/pprof"
	"os"
	"sort"

	"github.com/prometheus-community/windows_exporter/collector"
	"github.com/prometheus-community/windows_exporter/config"
	"github.com/prometheus-community/windows_exporter/log"
	"github.com/prometheus/common/version"
	webflag "github.com/prometheus/exporter-toolkit/web/kingpinflag"
	"gopkg.in/alecthomas/kingpin.v2"
)

const (
	DefaultCollectors            = "cpu"
	DefaultCollectorsPlaceholder = "[defaults]"
)

var (
	ConfigFile        = kingpin.Flag("config.file", "exporter configuration file path").Short('c').String()
	WebConfig         = webflag.AddFlags(kingpin.CommandLine, ":9182")
	MetricsPath       = kingpin.Flag("exp-endpoint", "endpoint where metrics will be exposed").Short('e').Default("/metrics").String()
	MaxRequest        = kingpin.Flag("max-request", "maximum number of concurrent request. 0 to desable").Short('m').Default("5").Int()
	EnabledCollectors = kingpin.Flag("collectors", "collectors comma separated list").Default(DefaultCollectors).String()
	PrintCollectors   = kingpin.Flag("collectors.print", "if true, print all available collectors").Short('p').Bool()
	TimeoutMargin     = kingpin.Flag("scrape.timeout-margin",
		"Seconds to subtract from the timeout allowed by the client. Tune to allow for overhead or high loads.",
	).Default("0.5").Float64()
)

func AddCommandLine() {
	log.AddFlags(kingpin.CommandLine)
	kingpin.Version(version.Print("custom_exporter"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	log.Debug("Logging has Started")

	checkConfigFile()
	checkPrintCollectors()
}

func checkConfigFile() {
	if *ConfigFile != "" {
		resolver, err := config.NewResolver(*ConfigFile)
		if err != nil {
			log.Fatalf("cant load config file: %v\n", err)
		}

		err = resolver.Bind(kingpin.CommandLine, os.Args[1:])
		if err != nil {
			log.Fatalf("%v\n", err)
		}

		*WebConfig.WebListenAddresses = (*WebConfig.WebListenAddresses)[1:]
		kingpin.Parse()
	}
}

func checkPrintCollectors() {
	if *PrintCollectors {
		collectors := collector.Available()
		collectorNames := make(sort.StringSlice, 0, len(collectors))

		for _, n := range collectors {
			collectorNames = append(collectorNames, n)
		}
		collectorNames.Sort()
		fmt.Printf("Available collectors:\n")
		for _, n := range collectorNames {
			fmt.Printf(" - %s\n", n)
		}
		return
	}
}
