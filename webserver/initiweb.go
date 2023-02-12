package webserver

import (
	_ "net/http/pprof"

	"github.com/StackExchange/wmi"
	"github.com/prometheus-community/windows_exporter/log"
)

func InitWbem() {
	log.Debugf("Initializing SWbemServices")
	s, err := wmi.InitializeSWbemServices(wmi.DefaultClient)
	if err != nil {
		log.Fatal(err)
	}
	wmi.DefaultClient.AllowMissingFields = true
	wmi.DefaultClient.SWbemServicesClient = s
}
