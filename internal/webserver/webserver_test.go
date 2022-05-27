package webserver

import (
	"context"
	"errors"
	"github.com/romanornr/autodealer/internal/config"
	"github.com/romanornr/autodealer/internal/singleton"
	"github.com/spf13/viper"
	"net/http"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

func TestGracefulShutdown(t *testing.T) {

	config.AppConfig()
	logrus.Infof("API route mounted on port %s\n", viper.GetString("SERVER_PORT"))
	logrus.Infof("creating http Server")

	//go singleton.singleton.GetDealerInstance()
	go singleton.GetDealer()
	go asyncWebWorker()

	httpServer := &http.Server{
		// viper config .env get server address
		Addr:           viper.GetViper().GetString("SERVER_ADDR") + ":" + viper.GetViper().GetString("SERVER_PORT"),
		Handler:        service(),
		ReadTimeout:    viper.GetViper().GetDuration("SERVER_READ_TIMEOUT"),
		WriteTimeout:   viper.GetViper().GetDuration("SERVER_WRITE_TIMEOUT"),
		MaxHeaderBytes: 1 << 20,
	}

	go func() {
		if err := httpServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			logrus.Errorf("error starting http server: %s\n", err)
		}
		logrus.Printf("server stopped serving new connections")
	}()

	// Create a context to attempt a graceful 10-second shutdown.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	time.Sleep(9 * time.Second)

	// Attempt the graceful shutdown by closing the listener
	if err := httpServer.Shutdown(ctx); err != nil {
		logrus.Fatalf("failed to shutdown: %v", err)
	}

	// Check for closed server connection
	if err := httpServer.Shutdown(ctx); err != nil {
		logrus.Fatalf("failed to shutdown: %v", err)
		t.Fatalf("failed to shutdown: %v", err)
	}
	logrus.Printf("server gracefully stopped")
}
