package move

import (
	"github.com/romanornr/autodealer/internal/singleton"
	"github.com/sirupsen/logrus"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
)

// TermStructure is a term structure that is a combination of other FTX MOVE contracts.
func TermStructure() string {
	d := singleton.GetDealerInstance()

	e, _ := d.ExchangeManager.GetExchangeByName("FTX")
	futures, err := e.GetAvailablePairs(asset.Futures)
	if err != nil {
		logrus.Errorf("Error getting available pairs: %s", err)
	}

	for _, f := range futures {
		logrus.Printf("futures %s\n", f.Delimiter)
	}

	//e.FetchTicker(context.Background(), "MOVE", asset.Spot)

	return "MOVE"
}
