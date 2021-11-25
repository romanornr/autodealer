package singleton

import (
	"github.com/sirupsen/logrus"
	"testing"
)

// TestGetDealerInstance tests the GetDealerInstance function
func TestGetDealerInstance(t *testing.T) {
	// create loop to test dealer
	//var d *dealer.Dealer

	for i := 0; i < 3; i++ {
		if IsDealerInitialized() == true {
			break
		}
		GetDealerInstance()
	}

	logrus.Printf("Dealer initialization: %v", IsDealerInitialized())

}
