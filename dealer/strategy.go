package dealer

import (
	exchange "github.com/thrasher-corp/gocryptotrader/exchanges"
	"github.com/thrasher-corp/gocryptotrader/exchanges/stream"
)

type Strategy interface {
	Init(k *Dealer, e exchange.IBotExchange) error
	OnFunding(d *Dealer, e exchange.IBotExchange, x stream.FundingData) error
}
