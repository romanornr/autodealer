// Copyright (c) 2021 Romano
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php.

package main

import (
	"github.com/romanornr/autodealer/dealer"
	"github.com/romanornr/autodealer/webserver"
	"github.com/sirupsen/logrus"
	"github.com/thrasher-corp/gocryptotrader/exchanges/account"
	"github.com/thrasher-corp/gocryptotrader/exchanges/ticker"
	"github.com/thrasher-corp/gocryptotrader/gctscript"
	gctlog "github.com/thrasher-corp/gocryptotrader/log"
	"github.com/thrasher-corp/gocryptotrader/signaler"
	"time"
)

func init() {
	go gctscript.Setup()
}

func main() {
	//go func() {
	//	d, err := dealer.NewBuilder().Build()
	//	if err != nil {
	//		logrus.Errorf("Failed to build builder: %v\n", err)
	//	}
	//	d.Run()
	//}()

	//ts := dealer.TickerStrategy{
	//	Interval: time.Second * 1,
	//	TickFunc: func(d *dealer.Dealer, e exchange.IBotExchange) {
	//		logrus.Errorf("TickFunc failed")
	//	},
	//}
	//
	//b := &dealer.BalancesStrategy{
	//	balances: sync.Map{},
	//	ticker:   ts,
	//}

	d, err := dealer.NewBuilder().Build()
	if err != nil {
		logrus.Errorf("expected no error, got %v\n", err)
	}
	balancesStrategy := dealer.NewBalancesStrategy(time.Second)

	e, err := d.ExchangeManager.GetExchangeByName("ftx")
	if err != nil {
		logrus.Errorf("expected error, got %s\n", err)
	}
	if err = balancesStrategy.Init(d, e); err != nil {
		logrus.Errorf("expected no error, got %v\n", err)
	}

	balancesStrategy.OnPrice(d, e, ticker.Price{})

	logrus.Infof("%v\n", balancesStrategy.OnBalanceChange(d, e, account.Change{}))



	//a, err := e.FetchAccountInfo(context.Background(), asset.Spot)
	//if err != nil {
	//	logrus.Errorf("fetch account info failed: %v\n", err)
	//}
	//balancesStrategy.
	//
	//holdings, ok := b.balances.Load(e.GetName())
	//if !ok {
	//	t.Errorf("expected no error, got %v\n", err)
	//}
	//
	//x := holdings.(account.Holdings)

	go webserver.New()
	interrupt := signaler.WaitForInterrupt()
	gctlog.Infof(gctlog.Global, "Captured %v, shutdown requested.\n", interrupt)
}
