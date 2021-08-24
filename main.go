// Copyright (c) 2021 Romano
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php.

package main

import (
	"github.com/romanornr/autodealer/engineManager"
	"github.com/romanornr/autodealer/webserver"
	"github.com/sirupsen/logrus"
	"github.com/thrasher-corp/gocryptotrader/engine"
	"github.com/thrasher-corp/gocryptotrader/gctscript"
	gctlog "github.com/thrasher-corp/gocryptotrader/log"
	"github.com/thrasher-corp/gocryptotrader/signaler"
)

func init() {
	go gctscript.Setup()
}

func main() {
	go func() {
		engineManager.StartMainEngine()
		if err := engine.Bot.WaitForInitialCurrencySync(); err != nil {
			logrus.Errorf("Failed to wait for initial currency sync: %s\n", err)
		}
		if engine.Bot.IsOnline() != true {
			logrus.Error("Connection manager is not online")
		}
	}()
	go webserver.New()
	interrupt := signaler.WaitForInterrupt()
	gctlog.Infof(gctlog.Global, "Captured %v, shutdown requested.\n", interrupt)
}
