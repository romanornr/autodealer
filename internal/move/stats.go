package move

import (
	"context"
	"github.com/romanornr/autodealer/internal/dealer"
	"github.com/sirupsen/logrus"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
	"github.com/thrasher-corp/gocryptotrader/exchanges/ftx"
	"strings"
)

// List struct for holding the list of all the moves stats
type List struct {
	Name      string
	Statistic []Statistic
}

// Statistic is a struct that holds the data for a statistic
type Statistic struct {
	Name  string
	Data  ftx.FuturesData
	Stats ftx.FutureStatsData
}

// GetStatistics returns a list of statistics for a given term structure
func GetStatistics(d *dealer.Dealer) (List, error) {
	var list List
	var s Statistic

	e, _ := d.ExchangeManager.GetExchangeByName("FTX")

	futures, err := e.GetAvailablePairs(asset.Futures)
	if err != nil {
		logrus.Errorf("Error getting available pairs: %s", err)
	}

	for _, future := range futures {
		// if futures strong contains "BTC-MOVE"
		if strings.Contains(future.String(), "BTC-MOVE") {
			f := ftx.FTX{Base: *e.GetBase()}
			data, err := f.GetFuture(context.Background(), future.String())
			if err != nil {
				return list, err
			}

			stats, err := f.GetFutureStats(context.Background(), future)
			if err != nil {
				return list, err
			}

			s.Name = future.String()
			s.Data = data
			s.Stats = stats

			list.Name = future.String()
			list.Statistic = append(list.Statistic, s)
		}
	}

	return list, nil
}
