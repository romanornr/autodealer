package singleton

import (
	"github.com/romanornr/autodealer/webserver"
	"sync"
	"testing"
)

// TestGetDealerInstance tests the GetDealerInstance function
func TestGetDealerInstance(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		webserver.GetDealerInstance()
	}()
	wg.Wait()
}
