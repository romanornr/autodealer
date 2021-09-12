package dealer

import (
	exchange "github.com/thrasher-corp/gocryptotrader/exchanges"
	"github.com/thrasher-corp/gocryptotrader/exchanges/stream"
)


// Strategy is a interface and defines all function needed for a user defined strategy. The RootStrategy provides a way to create and
// use strategies and action according to given strategies.
type Strategy interface {
	Init(k *Dealer, e exchange.IBotExchange) error
	OnFunding(d *Dealer, e exchange.IBotExchange, x stream.FundingData) error
}
