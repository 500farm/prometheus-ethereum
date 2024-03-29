package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/common/log"
)

type EthereumApiResults struct {
	// from WhatToMine API
	BlockTime       string  `json:"block_time"`
	BlockReward     float64 `json:"block_reward"`
	BlockReward24H  float64 `json:"block_reward24"`
	BlockReward3D   float64 `json:"block_reward3"`
	BlockReward7D   float64 `json:"block_reward7"`
	LastBlockNumber int64   `json:"last_block"`
	Difficulty      float64 `json:"difficulty"`
	Difficulty24H   float64 `json:"difficulty24"`
	Difficulty3D    float64 `json:"difficulty3"`
	Difficulty7D    float64 `json:"difficulty7"`
	NetworkHashRate int64   `json:"nethash"`
	// from CryptoCompare API
	ETHUSDPrice float64 `json:"USD"`
	// from wallet and pool APIs
	Balances []Balance
}

type Balance struct {
	Address  string
	Location string
	Balance  float64
}

type EthermineResponse struct {
	Status string `json:"status"`
	Error  string `json:"error"`
	Data   struct {
		Unpaid int64 `json:"unpaid"`
	} `json:"data"`
}

type EthermineDashboardResponse struct {
	Status string `json:"status"`
	Error  string `json:"error"`
	Data   struct {
		CurrentStatistics struct {
			Unpaid int64 `json:"unpaid"`
		} `json:"currentStatistics"`
	} `json:"data"`
}

type TwoMinersResponse struct {
	Stats struct {
		Balance int64 `json:"balance"`
	} `json:"stats"`
}

type EtherscanResponse struct {
	Status string `json:"status"`
	Result string `json:"result"`
}

func getEthereumInfoFromApis(addresses []string, verbose bool) (*EthereumApiResults, error) {
	result := new(EthereumApiResults)

	/*
		url := "https://whattomine.com/coins/151.json"
		body, err := apiCall(url, verbose)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(body, result); err != nil {
			return nil, err
		}
	*/
	{
		url := "https://min-api.cryptocompare.com/data/price?fsym=ETH&tsyms=USD"
		body, err := apiCall(url, verbose)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(body, result); err != nil {
			return nil, err
		}
	}
	{
		for _, address := range addresses {
			result.Balances = append(result.Balances, getBalances(address, verbose)...)
		}
	}

	return result, nil
}

func getBalances(address string, verbose bool) []Balance {
	balances := []Balance{}

	if *etherscanKey != "" {
		v, err := getWalletBalance(address, verbose, *etherscanKey)
		if err != nil {
			log.Errorln(err)
		} else {
			balances = append(balances, Balance{
				Address:  address,
				Location: "wallet",
				Balance:  v,
			})
		}
	}

	if *monitorEthermine {
		v, err := getEthermineBalance(address, verbose)
		if err != nil {
			log.Errorln(err)
		} else {
			balances = append(balances, Balance{
				Address:  address,
				Location: "ethermine-org",
				Balance:  v,
			})
		}
	}

	if *monitorTwoMiners {
		v, err := getTwoMinersBalance(address, verbose)
		if err != nil {
			log.Errorln(err)
		} else {
			balances = append(balances, Balance{
				Address:  address,
				Location: "2miners-com",
				Balance:  v,
			})
		}
	}

	return balances
}

func getWalletBalance(address string, verbose bool, apiKey string) (float64, error) {
	url := fmt.Sprintf("https://api.etherscan.io/api?module=account&action=balance&address=%s&tag=latest&apikey=%s", address, apiKey)
	body, err := apiCall(url, verbose)
	if err != nil {
		return 0, err
	}

	var result EtherscanResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return 0, err
	}
	if result.Status == "0" {
		return 0, errors.New("Etherscan API error: " + result.Result)
	}
	v, _ := strconv.ParseFloat(result.Result, 64)
	return v / 1e18, nil
}

func getEthermineBalance(address string, verbose bool) (float64, error) {
	url := "https://api.ethermine.org/miner/" + address + "/currentStats"
	body, err := apiCall(url, verbose)
	if err != nil {
		return 0, err
	}
	if bytes.Contains(body, []byte(`"data":"NO DATA"`)) {
		// When there are no workers, /currentStats may return "NO DATA", in this case we can get the unpaid balance from dashboard
		return getEthermineBalance2(address, verbose)
	}

	var result EthermineResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return 0, err
	}
	if result.Status != "OK" {
		return 0, errors.New("Ethermine API error: " + result.Error)
	}
	return float64(result.Data.Unpaid) / 1e18, nil
}

func getEthermineBalance2(address string, verbose bool) (float64, error) {
	url := "https://api.ethermine.org/miner/" + address + "/dashboard"
	body, err := apiCall(url, verbose)
	if err != nil {
		return 0, err
	}

	var result EthermineDashboardResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return 0, err
	}
	if result.Status != "OK" {
		return 0, errors.New("Ethermine API error: " + result.Error)
	}
	return float64(result.Data.CurrentStatistics.Unpaid) / 1e18, nil
}

func getTwoMinersBalance(address string, verbose bool) (float64, error) {
	url := "https://eth.2miners.com/api/accounts/" + address
	body, err := apiCall(url, verbose)
	if err != nil {
		return 0, err
	}

	var result TwoMinersResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return 0, err
	}
	return float64(result.Stats.Balance) / 1e9, nil
}

func apiCall(url string, verbose bool) ([]byte, error) {
	if verbose {
		log.Infoln("Fetching:", url)
	}
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return []byte{}, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, err
	}
	if verbose {
		log.Infoln(strings.TrimSpace(string(body)))
	}
	if resp.StatusCode != 200 {
		return nil, errors.New(resp.Status)
	}
	return body, nil
}
