package twap

import (
	"github.com/thrasher-corp/gocryptotrader/currency"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
	"github.com/thrasher-corp/gocryptotrader/exchanges/order"
	"time"
)

// Payload is the payload for the TWAP algorithm
type Payload struct {
	Exchange          string
	AccountID         string
	Pair              currency.Pair
	Asset             asset.Item // SPOT, FUTURES, INDEX
	Start             time.Time
	End               time.Time
	TargetAmountQuote float64
	Side              order.Side
	OrderType         order.Type
	Status            string
}
