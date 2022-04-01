package shortestPath

import (
	"errors"
	"github.com/sirupsen/logrus"
	"github.com/thrasher-corp/gocryptotrader/currency"
	exchange "github.com/thrasher-corp/gocryptotrader/exchanges"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
)

var (
	ErrNoPairsFound = errors.New("no pairs found")
	ErrNoPathFound  = errors.New("no path found")
)

// GetPrice returns the price the base currency is worth in the quote currency
func GetPrice(e exchange.IBotExchange, base, target currency.Code, a asset.Item) (price float64, err error) {

	pairs := MatchPairsForCurrency(e, target, a)

	if len(pairs) == 0 {
		logrus.Error("No pairs found for currency: ", target)
		return 0, ErrNoPairsFound
	}

	codes, err := PathToAsset(e, base, target, a)
	if err != nil || len(codes) == 0 {
		logrus.Error("No path found for currency: ", target)
		return 0, ErrNoPathFound
	}

	price, err = fetchTickerPrice(e, codes, a)
	if err != nil {
		logrus.Error("Failed to fetch ticker price: ", err)
		return 0, err
	}

	return price, nil
}
