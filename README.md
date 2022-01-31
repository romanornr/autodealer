# autodealer

[![Go CI](https://github.com/romanornr/autodealer/actions/workflows/go.yml/badge.svg)](https://github.com/romanornr/autodealer/actions/workflows/go.yml)


# autodealer 
#### autodealer is a simple application that allows you to connect to cryptocurrency exchanges to trade & transfer assets
###### This project relies on: https://github.com/thrasher-corp/gocryptotrader

##### Stack
- Golang
- Vuejs
- Redis

##### Features
- [x] Deposit
- [x] Withdraw
- [x] Transfer assets between exchanges
- [x] Buy/Sell
- [x] FTX Move Contracts term structure
- [ ] Tradingview library
- [ ] TWAP
- [ ] Rebalance portfolio
- [ ] Rebalance with TWAP
- [ ] Grid trading
- [ ] FTX Move Contracts dashboard panel
- [ ] Hedge with FTX Move Contracts

##### Supported exchange
- [x] FTX
- [x] Binance
- [x] Bitfinex
- [x] Huobi
- [x] Kraken
- [x] Okex
- [x] Bittrex

###### Disclaimer
This project is highly experimental. <br>
The author takes no responsibility
<br><br>

![alt text](https://github.com/romanornr/autodealer/blob/master/screenshots/1.png?raw=true)
<br><br>
![alt text](https://github.com/romanornr/autodealer/blob/master/screenshots/2.png?raw=true)
<br><br>
![alt text](https://github.com/romanornr/autodealer/blob/master/screenshots/3.png?raw=true)


###### API

http://127.0.0.1:3333/api

API documentation will be added later.


###### Minimum Recommended Specifications
- Go 1.17.6

##### libraries used
- gocryptotrader https://github.com/thrasher-corp/gocryptotrader
- asynq Golang https://github.com/hibiken/asynq

###### setup
``git clone https://github.com/romanornr/autodealer.git``

``cd autodealer``

###### config file
``cd ~/.gocryptotrader``

``wget https://raw.githubusercontent.com/thrasher-corp/gocryptotrader/master/config_example.json``

``mv config_example.json config.json``


###### async monitoring
``cd autodealer``

``./asynqmon``

![alt text](https://github.com/romanornr/autodealer/blob/master/screenshots/4.png?raw=true)