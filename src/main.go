package main

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
	"github.com/prometheus/common/version"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	listenAddress = kingpin.Flag(
		"listen",
		"Address to listen on.",
	).Default("0.0.0.0:8577").String()
	updateInterval = kingpin.Flag(
		"update-interval",
		"How often to query third-party APIs for updates.",
	).Default("1m").Duration()
)

func metricsHandler(w http.ResponseWriter, r *http.Request, ethereumCollector *EthereumCollector) {
	registry := prometheus.NewRegistry()
	registry.MustRegister(ethereumCollector)
	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)
}

func main() {
	kingpin.Version(version.Print("ethereum_exporter"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	log.Infoln("Starting ethereum exporter")

	ethereumCollector, _ := newEthereumCollector()
	ethInfo, err := getEthereumInfo()
	if err != nil {
		log.Errorln(err)
	} else {
		log.Infoln("Read initial Ethereum info: ", ethInfo)
		ethereumCollector.Update(ethInfo)
	}

	http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		metricsHandler(w, r, ethereumCollector)
	})
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
		<head>
		<title>Ethereum Exporter</title>
		<style>
		label{
		display:inline-block;
		width:75px;
		}
		form label {
		margin: 10px;
		}
		form input {
		margin: 10px;
		}
		</style>
		</head>
		<body>
		<h1>Ethereum Exporter</h1>
		<form action="/metrics">
		<label>Target:</label> <input type="text" name="target" placeholder="X.X.X.X" value="1.2.3.4"><br>
		<input type="submit" value="Submit">
		</form>
		</body>
		</html>`))
	})

	go func() {
		for {
			time.Sleep(*updateInterval)
			ethInfo, err := getEthereumInfo()
			if err != nil {
				log.Errorln(err)
			} else {
				ethereumCollector.Update(ethInfo)
			}
		}
	}()

	log.Infoln("Listening on", *listenAddress)
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}
