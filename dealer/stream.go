package dealer

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	exchange "github.com/thrasher-corp/gocryptotrader/exchanges"
	"github.com/thrasher-corp/gocryptotrader/exchanges/account"
	"github.com/thrasher-corp/gocryptotrader/exchanges/order"
	"github.com/thrasher-corp/gocryptotrader/exchanges/orderbook"
	"github.com/thrasher-corp/gocryptotrader/exchanges/stream"
)

var (
	ErrWebsocketNotSupported = errors.New("websocket not supported")
	ErrWebsocketNotEnabled   = errors.New("websocket is not enabled")
)

// Stream opens a websocket connection for the data stream, passing the market data to be processed as it's received.
// data passed to the FSM, there's a corresponding response channel depending on the data passed to it
// routing the messages to the appropriate channels.
func Stream(d *Dealer, e exchange.IBotExchange, s Strategy) error {
	ws, err := OpenWebsocket(e)
	if err != nil {
		return err
	}

	// This goroutine is supposed to never finish
	for data := range ws.ToRoutine {
		data := data
		go func() {
			err := handleData(d, e, s, data)
			if err != nil {
				logrus.Errorf("error handling data: %s\n", err)
			}
		}()
	}
	panic("unexpected end of channel")
}

// 1.Make sure the exchange can do websockets
// 2. Make sure the exchange has websockets enabled
// 3. Get the bridge to the exchange
// 4. Connect
// 5. FlushChannels

// OpenWebsocket checks if the client is sending a query to the interface instance, blocking on a channel, which was a select on a chan
// while the bolt process has a rule queuing, while the rest of the engine while be blocked on a chan
// So, non-blocking options were applied to get a non-blocked client and a non-blocked engine.
func OpenWebsocket(e exchange.IBotExchange) (*stream.Websocket, error) {
	if !e.IsWebsocketEnabled() {
		return nil, ErrWebsocketNotEnabled
	}

	if !e.SupportsWebsocket() {
		return nil, ErrWebsocketNotEnabled
	}

	ws, err := e.GetWebsocket()
	if err != nil {
		return nil, err
	}

	if !ws.IsConnecting() && !ws.IsConnecting() {
		err = ws.Connect()
		if err != nil {
			return nil, err
		}

		err = ws.FlushChannels()
		if err != nil {
			return nil, err
		}
	}
	return ws, nil
}

// handleData scans for any form of data like stream Warning Messages, Funding, kline events and orderbook actions
// For Funding you’ll need to consider how funds are rolled over, which will affect trading strategies work out if it shuts exchange
// Evaluate subscriptions changes for stops, liquidation etc. If applicable, send to trade module see if the type is unrecognized
// If this is the case, you’ll want to make notes about it. Mostly like for like exchange subscription errors.
// This type is only used when you’re not able to decipher what’s happened return an error if you need to

// handleData looks at the data passing it for appropriate security,
// addons, & addons, then passes it up to the bot core through the data channel.
// OnError is used when something goes wrong, lets the user know there was a problem.
func handleData(d *Dealer, e exchange.IBotExchange, s Strategy, data interface{}) error {
	switch x := data.(type) {
	case string:
		unhandledType(data, true)
	case error:
		return x
	case stream.FundingData:
		handleError("OnFunding", s.OnFunding(d, e, x))
	case *stream.KlineData:
		handleError("OnKline", s.OnKline(d, e, *x))
	case *orderbook.Base:
		handleError("OnOrderBook", s.OnOrderBook(d, e, *x))
	case *order.Detail:
		d.OnOrder(e, *x)
		handleError("OnOrder", s.OnOrder(d, e, *x))
	case *order.Modify:
		handleError("OnModify", s.OnModify(d, e, *x))
	case order.ClassificationError:
		unhandledType(data, true)
		if x.Err == nil {
			panic("unexpected error")
		}
		return x.Err
	case stream.UnhandledMessageWarning:
		unhandledType(data, true)
	case account.Change:
		handleError("OnBalanceChange", s.OnBalanceChange(d, e, x))
	default:
		handleError("OnUnrecognized", s.OnUnrecognized(d, e, x))
	}
	return nil
}

func unhandledType(data interface{}, warn bool) {
	if warn {
		logrus.Warn()
	}

	t := fmt.Sprintf("%T\n", data)
	logrus.Warnf("unhandledType: %v\n", t)
}

func handleError(method string, err error) {
	if err != nil {
		logrus.Warnf("method %v error: %s\n", method, err)
	}
}
