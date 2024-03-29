package main

import (
	"encoding/json"
	"net/http"
	"strings"
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
	monitorAddresses = kingpin.Flag(
		"monitor-addresses",
		"Monitor these Ethereum addresses for balance (comma-separated). Requires --etherscan-key, --ethermine-org or --2miners-org",
	).Default("").String()
	etherscanKey = kingpin.Flag(
		"etherscan-key",
		"Monitor wallet balances for --monitor-addresses via Etherscan.io API.",
	).Default("").String()
	monitorEthermine = kingpin.Flag(
		"ethermine-org",
		"Monitor unpaid balances for --monitor-addresses on Ethermine.org pool.",
	).Default("false").Bool()
	monitorTwoMiners = kingpin.Flag(
		"2miners-com",
		"Monitor unpaid balances for --monitor-addresses on 2miners.com pool.",
	).Default("false").Bool()
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

	monitorAddressesSplit := []string{}
	if *monitorAddresses != "" {
		monitorAddressesSplit = strings.Split(*monitorAddresses, ",")
	}

	ethereumCollectorGlobal := newEthereumCollector(false)
	ethereumCollector := newEthereumCollector(true)
	ethInfo, err := getEthereumInfoFromApis(monitorAddressesSplit, true)
	if err != nil {
		log.Fatalln("Error reading initial Ethereum info:", err)
	} else {
		t, _ := json.Marshal(ethInfo)
		log.Infoln("Read initial Ethereum info")
		log.Infoln(string(t))
		ethereumCollectorGlobal.UpdateFrom(ethInfo)
		ethereumCollector.UpdateFrom(ethInfo)
	}

	http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		metricsHandler(w, r, ethereumCollector)
	})
	http.HandleFunc("/metrics/global", func(w http.ResponseWriter, r *http.Request) {
		metricsHandler(w, r, ethereumCollectorGlobal)
	})
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
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
		<a href="metrics">Metrics</a><br>
		<a href="metrics/global">Without balances</a>
		</body>
		</html>`))
	})

	go func() {
		for {
			time.Sleep(*updateInterval)
			ethInfo, err := getEthereumInfoFromApis(monitorAddressesSplit, false)
			if err != nil {
				log.Errorln(err)
			} else {
				ethereumCollectorGlobal.UpdateFrom(ethInfo)
				ethereumCollector.UpdateFrom(ethInfo)
			}
		}
	}()

	log.Infoln("Listening on", *listenAddress)
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}
