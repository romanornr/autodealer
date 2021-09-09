package dealer

import exchange "github.com/thrasher-corp/gocryptotrader/exchanges"

type Strategy interface {
	Init(k *Dealer, e exchange.IBotExchange) error
}
