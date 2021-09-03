// Copyright (c) 2021 Romano
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php.

package main

import (
	"github.com/romanornr/autodealer/dealer"
	"github.com/romanornr/autodealer/flagparser"
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
	go func() {
		settings, _ := flagparser.DefaultEngineSettings()
		d, err := dealer.New(settings)
		if err != nil {
			logrus.Errorf("failed to load settings dealer: %s\n", err)
		}
		d.Run()
	}()
	go webserver.New()
	interrupt := signaler.WaitForInterrupt()
	gctlog.Infof(gctlog.Global, "Captured %v, shutdown requested.\n", interrupt)
}
