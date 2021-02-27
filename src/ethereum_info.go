package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/prometheus/common/log"
)

type Balances struct {
	WalletUSD float64
	WalletETH float64
	PoolUSD   float64
	PoolETH   float64
}

type EthereumInfo struct {
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
	// from Ethplorer and pool APIs
	Balances map[string]Balances
}

type EthermineResponse struct {
	Status string `json:"status"`
	Error  string `json:"error"`
	Data   struct {
		Unpaid int64 `json:"unpaid"`
	} `json:"data"`
}

type EthplorerResponse struct {
	ETH struct {
		Balance float64 `json:"balance"`
	} `json:"ETH"`
	Error struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func getEthereumInfo(addresses []string, verbose bool) (*EthereumInfo, error) {
	result := new(EthereumInfo)
	result.Balances = make(map[string]Balances)

	{
		url := "https://whattomine.com/coins/151.json"
		if verbose {
			log.Infoln(url)
		}
		resp, err := http.Get(url)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		if verbose {
			log.Infoln(string(body))
		}
		if err := json.Unmarshal(body, result); err != nil {
			return nil, err
		}
	}
	{
		url := "https://min-api.cryptocompare.com/data/price?fsym=ETH&tsyms=USD"
		if verbose {
			log.Infoln(url)
		}
		resp, err := http.Get(url)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		if verbose {
			log.Infoln(string(body))
		}
		if err := json.Unmarshal(body, result); err != nil {
			return nil, err
		}
	}
	{
		for _, address := range addresses {
			balances, err := getBalances(address, result.ETHUSDPrice, verbose)
			if err != nil {
				return nil, err
			}
			result.Balances[address] = *balances
		}
	}

	return result, nil
}

func getBalances(address string, ethPrice float64, verbose bool) (*Balances, error) {
	balances := new(Balances)

	{
		url := "https://api.ethplorer.io/getAddressInfo/" + address + "?apiKey=freekey"
		if verbose {
			log.Infoln(url)
		}
		resp, err := http.Get(url)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		if verbose {
			log.Infoln(string(body))
		}
		var result EthplorerResponse
		if err := json.Unmarshal(body, &result); err != nil {
			return nil, err
		}
		if result.Error.Code > 0 {
			return nil, errors.New("Ethplorer API error: " + result.Error.Message)
		}
		balances.WalletETH = result.ETH.Balance
		balances.WalletUSD = result.ETH.Balance * ethPrice
	}
	{
		url := "https://api.ethermine.org/miner/" + address + "/currentStats"
		if verbose {
			log.Infoln(url)
		}
		resp, err := http.Get(url)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		if verbose {
			log.Infoln(string(body))
		}
		var result EthermineResponse
		if err := json.Unmarshal(body, &result); err != nil {
			return nil, err
		}
		if result.Status != "OK" {
			return nil, errors.New("Ethermine API error: " + result.Error)
		}
		balances.PoolETH = float64(result.Data.Unpaid) / 1e18
		balances.PoolUSD = balances.PoolETH * ethPrice
	}

	return balances, nil
}
