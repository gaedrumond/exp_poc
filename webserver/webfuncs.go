package webserver

import (
	"fmt"
	"net/http"

	"github.com/prometheus-community/windows_exporter/log"
)

func WebHealthCheck(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "application/json")
	_, err := fmt.Fprintln(writer, `{"status":"ok"}`)
	if err != nil {
		log.Debugf("Failed to write to stream (web): %v", err)
	}
}

func WebConcurrencyLimit(limit int, next http.HandlerFunc) http.HandlerFunc {
	if limit <= 0 {
		return next
	}

	sem := make(chan struct{}, limit)
	return func(writer http.ResponseWriter, request *http.Request) {
		select {
		case sem <- struct{}{}:
			defer func() { <-sem }()
		default:
			writer.WriteHeader(http.StatusServiceUnavailable)
			_, _ = writer.Write([]byte("Too many concurrent requests"))
			return
		}
		next(writer, request)
	}
}
