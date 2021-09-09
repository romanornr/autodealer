package dealer

import (
	exchange "github.com/thrasher-corp/gocryptotrader/exchanges"
	"github.com/thrasher-corp/gocryptotrader/exchanges/ftx"
	"testing"
)

func TestExchangeFactoryRegister(t *testing.T) {
	factory := make(ExchangeFactory)
	testExchange := ftx.FTX{}
	factory.Register("ftx", func() (exchange.IBotExchange, error) {
		return &testExchange, nil
	})

	exchangeFactory, err := factory.NewExchangeByName("ftx")
	if err != nil {
		t.Fatalf("failed to create exchange factory: %v\n", err)
	}

	if exchangeFactory.GetName() != testExchange.GetName() {
		t.Errorf("NewExchangeByName failed: incorrect exchange name: %v\n", testExchange.GetName())
	}
}

func TestExchangeFactoryCreatorNotRegistered(t *testing.T) {
	factory := make(ExchangeFactory)
	_, err := factory.NewExchangeByName("invalid")
	if err != ErrCreatorNotRegistered {
		t.Errorf("Exchange creation failed: expected error: %v\n", ErrCreatorNotRegistered)
	}
}