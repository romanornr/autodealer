// Copyright (c) 2021 Romano
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php.

package main

import (
	"context"
	"github.com/romanornr/autodealer/dealer"
	"github.com/romanornr/autodealer/webserver"
	"github.com/sirupsen/logrus"
	"github.com/thrasher-corp/gocryptotrader/currency"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
	"github.com/thrasher-corp/gocryptotrader/exchanges/order"
	"github.com/thrasher-corp/gocryptotrader/gctscript"
	gctlog "github.com/thrasher-corp/gocryptotrader/log"
	"github.com/thrasher-corp/gocryptotrader/signaler"
	"time"
)

func init() {
	go gctscript.Setup()
}

func main() {
	d, err := dealer.NewBuilder().Build()
	if err != nil {
		logrus.Errorf("expected no error, got %v\n", err)
	}

	//var funding stream.FundingData
	//e, err := d.ExchangeManager.GetExchangeByName("ftx")
	//if err != nil {
	//	logrus.Errorf("expected error, got %s\n", err)
	//}

	//balancesStrategy := dealer.NewBalancesStrategy(time.Second * 5)
	//err = balancesStrategy.OnFunding(d, e, funding)
	//if err != nil {
	//	logrus.Errorf("balancing strategy failed for on funding: %s\n", err)
	//}

	//balances, err := d.Root.Get("balances")
	//if err != nil {
	//	logrus.Errorf("expected no error, got %s\n", err)
	//}

	go func() {
		d.Run(context.Background())
	}()

	///logrus.Info(balances)
	//
	var d2 = 10 * time.Second
	var t = time.Now().Add(d2)

	var orderReq order.GetOrdersRequest
	orderReq.AssetType = asset.Spot
	pairs := []string{"FTT/USD", "BTC/USD", "BTC/USDT"}
	p, err := currency.NewPairsFromStrings(pairs)
	if err != nil {
		logrus.Errorf("new pairs failed: %s\n", err)
	}
	orderReq.Pairs = p
	if orderReq.Validate() != nil {
		logrus.Errorf("failed to validate order: %s\n", orderReq)
	}

	go func() {
		for {
			logrus.Infof("getting active orders")
			o, err := d.GetActiveOrders(context.Background(), "FTX", orderReq)
			if err != nil {
				logrus.Errorf("error active orders: %s\n", err)
			}
			logrus.Infof("stream strategy: %v\n", o)
			if time.Now().Before(t) {
				time.Sleep(time.Second * 5)
				continue
			}
		}
	}()

	go webserver.New()

	interrupt := signaler.WaitForInterrupt()
	gctlog.Infof(gctlog.Global, "Captured %v, shutdown requested.\n", interrupt)
}
