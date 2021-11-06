# ethereum_exporter

Prometheus exporter reporting:
- Overall Ethereum network stats (diffculty etc).
- ETH price.
- Balances of specified Ethereum addresses in ETH and USD (requires [Etherscan.io](https://etherscan.io/) API key).
- Unpaid balances on [Ethermine.org](https://ethermine.org) pool in ETH and USD. So far, only Ethermine pool is supported.

### Usage

```
docker run -d --restart always -p 8577:8577 500farm/ethereum-exporter ARGS...
```
To test, open http://localhost:8577. If does not work, check container output with `docker logs`.


### Optional args

```
--listen=IP:PORT
    Address to listen on (default 0.0.0.0:8577).

--update-interval=
    How often to query third-party APIs for updates (default 1m).
    
--monitor-addresses=0x....,0x....
    Addresses to watch on Ethereum network and Ethermine pool. Separate multiple addresses with comma.
    Use with --etherscan-key and/or --ethermine-org.
    
--etherscan-key=KEY
    Query balances for monitored addresses from Etherscan API. Specify the API key.
    
--ethermine-org
    Query unpaid balances for monitored addresses from Ethermine pool. No key required.
```

To get your Etherscan API key, register on https://etherscan.io/login, then go to API KEYS section of your account.

### Example output

```
# HELP ethereum_address_balance Balance on Ethereum address and unpaid on Ethermine pool
# TYPE ethereum_address_balance gauge
ethereum_address_balance{address="0xccB6CBEafc3b937db1Bbc837E7992a9dE9D34Aa4",currency="ETH",location="ethermine-org"} 0.09006578087788984
ethereum_address_balance{address="0xccB6CBEafc3b937db1Bbc837E7992a9dE9D34Aa4",currency="ETH",location="wallet"} 0.4816605943888761
ethereum_address_balance{address="0xccB6CBEafc3b937db1Bbc837E7992a9dE9D34Aa4",currency="USD",location="ethermine-org"} 306.98921412228754
ethereum_address_balance{address="0xccB6CBEafc3b937db1Bbc837E7992a9dE9D34Aa4",currency="USD",location="wallet"} 1641.740135974484

# HELP ethereum_block_reward_eth Reward for the last found block, in ETH
# TYPE ethereum_block_reward_eth gauge
ethereum_block_reward_eth 3.09754357871771

# HELP ethereum_block_time_seconds Time it took to find the last block, in seconds
# TYPE ethereum_block_time_seconds gauge
ethereum_block_time_seconds 13.526

# HELP ethereum_difficulty_hashes Last block difficulty in hashes
# TYPE ethereum_difficulty_hashes gauge
ethereum_difficulty_hashes 9.522379862834492e+15

# HELP ethereum_earnings_per_gh_per_hour_dollars Mining earnings, dollars per GH/s per hour
# TYPE ethereum_earnings_per_gh_per_hour_dollars gauge
ethereum_earnings_per_gh_per_hour_dollars 3.9915145987149914

# HELP ethereum_eth_price_dollars Current ETH price, in USD
# TYPE ethereum_eth_price_dollars gauge
ethereum_eth_price_dollars 3408.5

# HELP ethereum_last_block_number Number of the last found block
# TYPE ethereum_last_block_number gauge
ethereum_last_block_number 1.3342989e+07

# HELP ethereum_network_hashrate_hashes_per_sec Current network hasrate, in H/s
# TYPE ethereum_network_hashrate_hashes_per_sec gauge
ethereum_network_hashrate_hashes_per_sec 7.04005608667343e+14
```
