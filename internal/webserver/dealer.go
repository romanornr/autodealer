package webserver

import (
	"github.com/romanornr/autodealer/internal/dealer"
	"github.com/sirupsen/logrus"
	"sync"
	"sync/atomic"
)

// TODO has to work with webserver routes to pass dealer
// Singleton perhaps?

var initialized uint32

var instance *dealer.Dealer
var mu sync.Mutex

// GetDealerInstance function takes no arguments and returns the singleton defined in `package dealer` object.
// So this function figures out we have initialized the dealer or not, it takes a very long time to do this introspection, so we use `atomic uint32`.
// Next, in this case, as this Go code is part of a singleton configuration, we early return `nil` if the dealer has been initialized before.
// This gives us a single point of entry to initialize the object and is great for dependency injection. Each time we are trying to access the dealer
// instance, we make sure we have one, otherwise we create the dealer.
// Next, dealing with concurrent accesses. If a process wanted to access the dealer object, we use sync.Mutex to avoid a race condition.
func GetDealerInstance() *dealer.Dealer {
	var err error
	if atomic.LoadUint32(&initialized) == 1 {
		return instance
	}
	mu.Lock()
	defer mu.Unlock()
	if initialized == 0 {
		instance, err = dealer.NewBuilder().Build()
		if err != nil {
			logrus.Errorf("failed to create instance: %v", err)
		}
		atomic.StoreUint32(&initialized, 1)
		logrus.Infof("Created dealer instance\n")
	}
	return instance
}
