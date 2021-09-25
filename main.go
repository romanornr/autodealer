// Copyright (c) 2021 Romano
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php.

package main

import (
	"github.com/romanornr/autodealer/dealer"
	"github.com/romanornr/autodealer/webserver"
	"github.com/sirupsen/logrus"
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
	balancesStrategy := dealer.NewBalancesStrategy(time.Second)

	e, err := d.ExchangeManager.GetExchangeByName("ftx")
	if err != nil {
		logrus.Errorf("expected error, got %s\n", err)
	}
	if err = balancesStrategy.Init(d, e); err != nil {
		logrus.Errorf("expected no error, got %v\n", err)
	}

	go func() {
		dealer.Stream(d, e, balancesStrategy)
	}()

	var d2 = 200 * time.Second
	var t = time.Now().Add(d2)

	go func() {
		for {
			logrus.Infof("stream strategy: %s\n", balancesStrategy)
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
