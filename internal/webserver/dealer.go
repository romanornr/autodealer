package webserver

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/romanornr/autodealer/internal/dealer"
	"github.com/sirupsen/logrus"
)

// TODO has to work with webserver routes to pass dealer
// Singleton perhaps?

var initialized uint32
var instance *dealer.Dealer
var mu sync.Mutex

// GetDealerInstance checks if the dealer has been initialized. If it has, we return the instance.
// Next, we lock the mutex. This is to avoid a race condition.
// If the dealer has not been initialized, we create the dealer. We store the dealer in the instance variable.
// We set the initialized flag to 1. We unlock the mutex. We return the instance.
func GetDealerInstance() *dealer.Dealer {
	var err error
	if atomic.LoadUint32(&initialized) == 1 {
		logrus.Print("Dealer already loaded")
		return instance
	}
	mu.Lock()
	defer mu.Unlock()
	if initialized == 0 {
		instance, err = dealer.NewBuilder().Build(context.Background())
		if err != nil {
			logrus.Errorf("failed to create instance: %v", err)
		}
		atomic.StoreUint32(&initialized, 1)

		go instance.Run(context.Background())

		logrus.Infof("Created dealer instance\n")
	}
	return instance
}
