package algo

import (
	"github.com/thrasher-corp/gocryptotrader/currency"
	exchange "github.com/thrasher-corp/gocryptotrader/exchanges"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
)

// MatchPairsForCurrency returns a list of pairs that match the given currency
func MatchPairsForCurrency(e exchange.IBotExchange, code currency.Code, assetType asset.Item) currency.Pairs {
	availablePairs, err := e.GetAvailablePairs(assetType)
	if err != nil {
		return nil
	}

	matchingPairs := currency.Pairs{}
	for _, pair := range availablePairs {
		if pair.Base.String() == code.String() {
			matchingPairs = append(matchingPairs, pair)
		}
	}

	return matchingPairs
}

// Get the dollar value of the given asset. However, there might not be a direct conversion to USD so we need to use the exchange's conversion rate
// to get the value in USD. Possibly use an intermediate currency pair to convert to USD.
