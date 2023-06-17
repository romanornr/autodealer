// Copyright (c) 2021 Romano
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php.

package main

import (
	"github.com/hibiken/asynq"
	"github.com/romanornr/autodealer/webserver"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/thrasher-corp/gocryptotrader/gctscript"
	gctlog "github.com/thrasher-corp/gocryptotrader/log"
	"github.com/thrasher-corp/gocryptotrader/signaler"
	"os"
)

const redisAddr = "127.0.0.1:6379"

func init() {
	go gctscript.Setup()
}

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout}).With().Timestamp().Logger()
	log.Logger = logger

	// Initialize the Asynq client with the Redis address
	client := asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr})
	defer client.Close()

	// Start the webserver
	log.Info().Msg("Starting server at 127.0.0.1:3333")
	webserver.New()

	// Wait for an interrupt signal to gracefully shutdown the server and log the shutdown request
	interrupt := signaler.WaitForInterrupt()
	gctlog.Infof(gctlog.Global, "Captured %v, shutdown requested.\n", interrupt)
}
