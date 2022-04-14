package move

import (
	"context"
	"github.com/romanornr/autodealer/internal/dealer"
	"github.com/sirupsen/logrus"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
	"github.com/thrasher-corp/gocryptotrader/exchanges/ftx"
	"sort"
	"strings"
)

// TermStructure is a struct that holds the data for a term structure
type TermStructure struct {
	MOVE []ftx.FuturesData
}

// GetTermStructure is a term structure that is a combination of other FTX MOVE contracts
func GetTermStructure(d *dealer.Dealer) TermStructure {
	var termStructure TermStructure
	e, _ := d.ExchangeManager.GetExchangeByName("FTX")

	futures, err := e.GetAvailablePairs(asset.Futures)
	if err != nil {
		logrus.Errorf("Error getting available pairs: %s", err)
	}

	var quarterlyCount int
	for _, future := range futures {
		// if futures strong contains "BTC-MOVE"
		if strings.Contains(future.String(), "BTC-MOVE") {
			f := ftx.FTX{Base: *e.GetBase()}
			stat, err := f.GetFuture(context.Background(), future.String())
			if err != nil {
				logrus.Errorf("Error getting future: %s", err)
			}

			if stat.Group == "quarterly" {
				quarterlyCount++
			}

			// Avoid adding "Today", "Next Week" and "This Quarter" MOVE Contracts
			// Avoid first Quarter MOVE Contract by checking if quarterlyCount is 1
			if stat.ExpiryDescription == "Today" || stat.ExpiryDescription == "This Week" || stat.ExpiryDescription == "This Month" || quarterlyCount == 1 && stat.Group == "quarterly" {
				continue
			}

			termStructure.MOVE = append(termStructure.MOVE, stat)
		}
	}

	// sort MOVE by expiry
	sort.Slice(termStructure.MOVE, func(i, j int) bool {
		return termStructure.MOVE[i].Expiry.Before(termStructure.MOVE[j].Expiry)
	})

	return termStructure
}
