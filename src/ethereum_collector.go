package main

import (
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
)

type EthereumCollector struct {
	block_time_seconds                prometheus.Gauge
	block_reward_eth                  prometheus.Gauge
	last_block_number                 prometheus.Gauge
	difficulty_hashes                 prometheus.Gauge
	network_hashrate_hashes_per_sec   prometheus.Gauge
	eth_price_dollars                 prometheus.Gauge
	earnings_per_ghs_per_hour_dollars prometheus.Gauge
	address_balance                   *prometheus.GaugeVec
}

func newEthereumCollector() (*EthereumCollector, error) {
	namespace := "ethereum"

	return &EthereumCollector{
		block_time_seconds: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "block_time_seconds",
			Help:      "Time it took to find the last block, in seconds",
		}),
		block_reward_eth: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "block_reward_eth",
			Help:      "Reward for the last found block, in ETH",
		}),
		last_block_number: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "last_block_number",
			Help:      "Number of the last found block",
		}),
		difficulty_hashes: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "difficulty_hashes",
			Help:      "Last block difficulty in hashes",
		}),
		network_hashrate_hashes_per_sec: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "network_hashrate_hashes_per_sec",
			Help:      "Current network hasrate, in H/s",
		}),
		eth_price_dollars: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "eth_price_dollars",
			Help:      "Current ETH price, in USD",
		}),
		earnings_per_ghs_per_hour_dollars: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "earnings_per_gh_per_hour_dollars",
			Help:      "Mining earnings, dollars per GH/s per hour",
		}),
		address_balance: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "address_balance",
			Help:      "Balance on Ethereum address and unpaid on Ethermine pool",
		}, []string{"address", "currency", "location"}),
	}, nil
}

func (e *EthereumCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.block_time_seconds.Desc()
	ch <- e.block_reward_eth.Desc()
	ch <- e.last_block_number.Desc()
	ch <- e.difficulty_hashes.Desc()
	ch <- e.network_hashrate_hashes_per_sec.Desc()
	ch <- e.eth_price_dollars.Desc()
	ch <- e.earnings_per_ghs_per_hour_dollars.Desc()
	e.address_balance.Describe(ch)
}

func (e *EthereumCollector) Collect(ch chan<- prometheus.Metric) {
	ch <- e.block_time_seconds
	ch <- e.block_reward_eth
	ch <- e.last_block_number
	ch <- e.difficulty_hashes
	ch <- e.network_hashrate_hashes_per_sec
	ch <- e.eth_price_dollars
	ch <- e.earnings_per_ghs_per_hour_dollars
	e.address_balance.Collect(ch)
}

func (e *EthereumCollector) Update(info *EthereumInfo) {
	t, _ := strconv.ParseFloat(info.BlockTime, 64)
	e.block_time_seconds.Set(t)
	e.block_reward_eth.Set(info.BlockReward)
	e.last_block_number.Set(float64(info.LastBlockNumber))
	e.difficulty_hashes.Set(info.Difficulty)
	e.network_hashrate_hashes_per_sec.Set(float64(info.NetworkHashRate))
	e.eth_price_dollars.Set(info.ETHUSDPrice)
	e.earnings_per_ghs_per_hour_dollars.Set(1e9 * 3600 / info.Difficulty * info.BlockReward * info.ETHUSDPrice)

	for address, balances := range info.Balances {
		e.address_balance.With(prometheus.Labels{"address": address, "currency": "ETH", "location": "wallet"}).Set(balances.WalletETH)
		e.address_balance.With(prometheus.Labels{"address": address, "currency": "USD", "location": "wallet"}).Set(balances.WalletUSD)
		e.address_balance.With(prometheus.Labels{"address": address, "currency": "ETH", "location": "pool"}).Set(balances.PoolETH)
		e.address_balance.With(prometheus.Labels{"address": address, "currency": "USD", "location": "pool"}).Set(balances.PoolUSD)
	}
}
