package webserver

import (
	"context"
	"github.com/romanornr/autodealer/dealer"
	"github.com/sirupsen/logrus"
)

// TODO has to work with webserver routes to pass dealer
// Singleton perhaps?
func RunDealer() *dealer.Dealer {
	d, err := dealer.NewBuilder().Build()
	if err != nil {
		logrus.Errorf("expected no error, got %v\n", err)
	}

	go func() {
		d.Run(context.Background())
	}()

	return d
}
