package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	listeningAddress = flag.String("telemetry.address", ":8080", "Address on which to expose metrics.")
	metricsEndpoint  = flag.String("telemetry.endpoint", "/metrics", "Path under which to expose metrics.")
	addr             = flag.String("addr", "unix:///var/run/docker.sock", "Docker address to connect to")
)

func main() {
	flag.Parse()

	manager := newDockerManager(*addr)
	exporter := NewExporter(manager)
	prometheus.MustRegister(exporter)

	log.Printf("Starting Server: %s", *listeningAddress)
	http.Handle(*metricsEndpoint, prometheus.Handler())
	log.Fatal(http.ListenAndServe(*listeningAddress, nil))
}
