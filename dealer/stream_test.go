package dealer

import (
	"github.com/sirupsen/logrus"
	exchange "github.com/thrasher-corp/gocryptotrader/exchanges"
	"github.com/thrasher-corp/gocryptotrader/exchanges/ftx"
	"testing"
	"time"
)

func TestStream(t *testing.T) {
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
	//

	//tickerStrategy := TickerStrategy{
	//	Interval: time.Millisecond * 5,
	//	TickerFunc: func(d *Dealer, e exchange.IBotExchange) {
	//		t.Log("failed barf")
	//	},
	//}
	//
	//err = tickerStrategy.Init(d, exchangeFactory)
	//if err != nil {
	//	t.Fatalf("Expected no error, got %v", err)
	//}

	s := NewRootStrategy()
	s.Add("test", &s)


	for i := 0; i < 1; i++ {

		go func() {
			time.Sleep(5 * time.Second)
		}()
		logrus.Info("opening stream")
		err = Stream(d, exchangeFactory, &s)
		if err != nil {
			t.Errorf("expected no error, got %s\n", err)
		}
	}
}
