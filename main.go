// Copyright (c) 2021 Romano
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php.

package main

import (
	"github.com/hibiken/asynq"
	webserver2 "github.com/romanornr/autodealer/webserver"
	"github.com/thrasher-corp/gocryptotrader/gctscript"
	gctlog "github.com/thrasher-corp/gocryptotrader/log"
	"github.com/thrasher-corp/gocryptotrader/signaler"
)

const redisAddr = "127.0.0.1:6379"

func init() {
	go gctscript.Setup()
}

func main() {

	client := asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr})
	defer client.Close()

	// ------------------------------------------------------
	// Example 1: Enqueue task to be processed immediately.
	//            Use (*Client).Enqueue method.
	// ------------------------------------------------------

	//task, err := algo.NewTwapOrderTask("FTX")
	//if err != nil {
	//	log.Fatalf("could not create task: %v", err)
	//}
	//info, err := client.Enqueue(task)
	//if err != nil {
	//	log.Fatalf("could not enqueue task: %v", err)
	//}
	//logrus.Printf("enqueued task: id=%s queue=%s", info.ID, info.Queue)
	// d, err := dealer.NewBuilder().Build()
	// if err != nil {
	//	logrus.Errorf("expected no error, got %v\n", err)
	// }
	//
	// go func() {
	//	d.Run(context.Background())
	// }()
	// //
	// var d2 = 10 * time.Second
	// var t = time.Now().Add(d2)
	// //
	// var orderReq order.GetOrdersRequest
	// orderReq.AssetType = asset.Spot
	// pairs := []string{"FTT/USD", "BTC/USD", "BTC/USDT"}
	// p, err := currency.NewPairsFromStrings(pairs)
	// if err != nil {
	//	logrus.Errorf("new pairs failed: %s\n", err)
	// }
	// orderReq.Pairs = p
	// //if orderReq.Validate() != nil {
	// //	logrus.Errorf("failed to validate order: %s\n", orderReq)
	// //}
	// //
	// go func() {
	//	for {
	//		logrus.Infof("getting active orders")
	//		o, err := d.GetActiveOrders(context.Background(), "FTX", orderReq)
	//		if err != nil {
	//			logrus.Errorf("error active orders: %s\n", err)
	//		}
	//		logrus.Infof("stream strategy: %v\n", o[0])
	//		if time.Now().Before(t) {
	//			time.Sleep(time.Second * 5)
	//			continue
	//		}
	//	}
	// }()

	webserver2.New()

	interrupt := signaler.WaitForInterrupt()
	gctlog.Infof(gctlog.Global, "Captured %v, shutdown requested.\n", interrupt)
}
