package dealer

import (
	"github.com/romanornr/autodealer/flagparser"
	"testing"
)

func TestSetupExchanges(t *testing.T) {
	t.Log("Setup")

	settings, _ := flagparser.DefaultEngineSettings()
	_, err := NewDealer(settings)
	if err != nil {
		t.Errorf("Failed to create dealer: %s\n", err)
	}

	t.Log("SetupExchanges succeeded.")
}