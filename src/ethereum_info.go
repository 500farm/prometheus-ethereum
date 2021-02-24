package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

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
}

func getEthereumInfo() (*EthereumInfo, error) {
	result := new(EthereumInfo)

	{
		resp, err := http.Get("https://whattomine.com/coins/151.json")
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(body, result); err != nil {
			return nil, err
		}
	}
	{
		resp, err := http.Get("https://min-api.cryptocompare.com/data/price?fsym=ETH&tsyms=USD")
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(body, result); err != nil {
			return nil, err
		}
	}

	return result, nil
}
