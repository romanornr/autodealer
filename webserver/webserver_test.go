package webserver

import (
	"context"
	"errors"
	"github.com/romanornr/autodealer/config"
	"github.com/romanornr/autodealer/singleton"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"net/http"
	"os/signal"
	"syscall"
	"testing"
	"time"
)

func TestGracefulShutdown(t *testing.T) {

	// Create context that listns for the interrupt signal.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// load config
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

	// initializing the server in a goroutine so that
	// it won't block the graceful shutdown handling below
	go func() {
		if err := httpServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			logrus.Errorf("error starting http server: %s\n", err)
		}
		logrus.Printf("server stopped serving new connections")
	}()

	// Listen for the interrupt signal
	<-ctx.Done()

	// Restore default behavior on interrupt signal and notify user of shutdown.
	stop()
	logrus.Infof("shutting downserver gracefully, press Ctrl+C again to force")

	// The context is used to inform the server it has 5 seconds to finish
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		logrus.Errorf("error shutting down http server: %s\n", err)
	}

	logrus.Infof("server exiting")

	time.Sleep(11 * time.Second)

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
