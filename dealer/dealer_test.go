package dealer

import (
	"github.com/thrasher-corp/gocryptotrader/engine"
	"testing"
)

func TestSetupExchanges(t *testing.T) {
	t.Log("Setup")

	var settings engine.Settings
	_, err := NewDealer(settings)
	if err != nil {
		t.Errorf("Failed to create dealer: %s\n", err)
	}

	t.Log("SetupExchanges succeeded.")
}