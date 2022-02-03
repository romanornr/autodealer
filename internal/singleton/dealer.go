package singleton

import (
	"context"
	"github.com/romanornr/autodealer/internal/dealer"
	"github.com/sirupsen/logrus"
	"sync"
	"sync/atomic"
)

var (
	initialized uint32 // atomic flag
	instance    *dealer.Dealer
	once        sync.Once
	err         error
)

//GetDealerInstance can only create and return an initialized instance of Dealer.
//This means that GetDealerInstance will NOT create a new instance, if there is already an instance running.
func GetDealerInstance() *dealer.Dealer {
	once.Do(func() {
		instance, err = dealer.NewBuilder().Build(context.Background())
		if err != nil {
			logrus.Errorf("failed to create instance: %v", err)
		}
		go instance.Run(context.Background())
		atomic.StoreUint32(&initialized, 1)
		logrus.Infof("Created dealer instance\n")
	})
	return instance
}

// IsDealerInitialized function to check if dealer is initialized
func IsDealerInitialized() bool {
	return atomic.LoadUint32(&initialized) == 1
}
