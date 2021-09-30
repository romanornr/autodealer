package dealer

import (
	"context"
	"github.com/sirupsen/logrus"
	"testing"
)

func TestStream(t *testing.T) {
	d, err := NewBuilder().Build()
	if err != nil {
		logrus.Errorf("expected no error, got %v", err)
	}

	exchange, _ := d.ExchangeManager.GetExchangeByName("FTX")
	if err != nil {
		t.Fatalf("failed to create exchange factory: %v\n", err)
	}

	logrus.Infof("exchange %s websocket enabled %T\n", exchange.GetName(), exchange.IsWebsocketEnabled())
	s := NewRootStrategy()
	s.Add("test", &s)

	logrus.Info("opening stream")
	go func() {
		err = Stream(context.Background(), d, exchange, &s)
	}()
	_, err = exchange.GetWebsocket()
	if err != nil {
		t.Errorf("expected no error, got %s\n", err)
	}
}
