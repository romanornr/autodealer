// Copyright (c) 2021 Romano
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php.

package main

import (
	"context"
	"github.com/romanornr/autodealer/dealer"
	"github.com/romanornr/autodealer/webserver"
	"github.com/sirupsen/logrus"
	"github.com/thrasher-corp/gocryptotrader/gctscript"
	gctlog "github.com/thrasher-corp/gocryptotrader/log"
	"github.com/thrasher-corp/gocryptotrader/signaler"
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
	//var d2 = 200 * time.Second
	//var t = time.Now().Add(d2)
	//
	//go func() {
	//	for {
	//		logrus.Infof("stream strategy: %v\n", balances)
	//		if time.Now().Before(t) {
	//			time.Sleep(time.Second * 5)
	//			continue
	//		}
	//	}
	//}()

	go webserver.New()

	interrupt := signaler.WaitForInterrupt()
	gctlog.Infof(gctlog.Global, "Captured %v, shutdown requested.\n", interrupt)
}
