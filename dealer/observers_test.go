package dealer

import (
	"github.com/sirupsen/logrus"
	exchange "github.com/thrasher-corp/gocryptotrader/exchanges"
	"github.com/thrasher-corp/gocryptotrader/exchanges/ftx"
	"github.com/thrasher-corp/gocryptotrader/exchanges/order"
	"testing"
)

func TestOnFilled(t *testing.T) {
	d, err := NewBuilder().Build()
	if err != nil {
		logrus.Errorf("expected no error, got %v", err)
	}

	factory := make(ExchangeFactory)
	testExchange := ftx.FTX{}
	factory.Register("ftx", func() (exchange.IBotExchange, error) {
		return &testExchange, nil
	})

	exchangeFactory, err := factory.NewExchangeByName("ftx")
	if err != nil {
		t.Fatalf("failed to create exchange factory: %v\n", err)
	}

	o := order.Detail{
		ImmediateOrCancel: true,
		Amount:            1,
	}

	observer := &Slots{
		OnFilledSlot: func(dealer *Dealer, e exchange.IBotExchange, x order.Detail) {
			if x.ImmediateOrCancel != o.ImmediateOrCancel {
				t.Errorf("expected FillOrKill to be true")
			}
		},
	}

	observer.OnFilled(d, exchangeFactory, o)
	d.OnOrder(exchangeFactory, o)
}
