package singleton

import (
	"context"
	"errors"
	"github.com/romanornr/autodealer/dealer"
	"github.com/rs/zerolog/log"
	"sync"
)

var Ds = &DealerSingleton{}

type DealerSingleton struct {
	initialized bool
	instance    *dealer.Dealer
	mutex       sync.Mutex
	err         error
	cancel      context.CancelFunc
}

func GetDealer(ctx context.Context) (*dealer.Dealer, error) {
	return Ds.InitAndGetDealer(ctx)
}

func (ds *DealerSingleton) InitAndGetDealer(ctx context.Context) (*dealer.Dealer, error) {
	if ctx.Err() != nil {
		return nil, errors.New("context is canceled or deadline exceeded")
	}

	ctx, ds.cancel = context.WithCancel(ctx)

	ds.mutex.Lock()
	defer ds.mutex.Unlock()

	// Only initialize if not already initialized
	if !ds.initialized {
		ds.instance, ds.err = dealer.NewBuilder().Build(ctx)
		if ds.err != nil {
			log.Error().Err(ds.err).Msg("failed to create instance")
			return nil, ds.err
		}
		// As run does not return an error, we just run it in a goroutine
		go ds.instance.Run(ctx)
		ds.initialized = true
		log.Info().Msg("Created dealer instance")
	}
	return ds.instance, nil
}

func (ds *DealerSingleton) isDealerInitialized() bool {
	return ds.initialized
}
