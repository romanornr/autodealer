package dealer

import (
	"context"
	"errors"
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/sirupsen/logrus"
	exchange "github.com/thrasher-corp/gocryptotrader/exchanges"
	"github.com/thrasher-corp/gocryptotrader/exchanges/account"
	"github.com/thrasher-corp/gocryptotrader/exchanges/fill"
	"github.com/thrasher-corp/gocryptotrader/exchanges/order"
	"github.com/thrasher-corp/gocryptotrader/exchanges/orderbook"
	"github.com/thrasher-corp/gocryptotrader/exchanges/stream"
	"github.com/thrasher-corp/gocryptotrader/exchanges/ticker"
	"github.com/thrasher-corp/gocryptotrader/exchanges/trade"
)

var (
	ErrWebsocketNotSupported = errors.New("websocket not supported")
	ErrWebsocketNotEnabled   = errors.New("websocket is not enabled")
)

// This code takes Stream to which you can subscribe, and wires the relevant streams to the relevant Handlers.
// Handlers are called automatically. This functionality will eventually allow for both manual and auto-generated responses to take place.
// For now, the workflow is you write the code, and the auto-generated responses should do what you tell it to do.
// This code should be good enough for daily trading, where every millisecond is important. You can just let the Listener run for an hour or so, and it will take care of the rest.

// Stream is the main entry point for the bot. It is responsible for opening the websocket connection, and then listening for data on the websocket to come in.
// When data comes through (this goroutine never dies), it handles all the types of messages available on the websocket.
// The *exchange.IBotExchange contains the underlying Golang Websocket library, which must be imported with an alias.
// By using that package with the alias "exchange", we can consolidate the Exchange package into this one without problematic circular imports.
func Stream(ctx context.Context, d *Dealer, e exchange.IBotExchange, s Strategy) error {
	ws, err := OpenWebsocket(e)
	if err != nil {
		return err
	}

	// This goroutine is supposed to never finish
	for data := range ws.ToRoutine {
		err := handleData(d, e, s, data)
		if err != nil {
			logrus.Errorf("error handling data: %s\n", err)
		}
	}

	panic("unexpected end of channel")
}

// 1.Make sure the exchange can do websockets
// 2. Make sure the exchange has websockets enabled
// 3. Get the bridge to the exchange
// 4. Connect
// 5. FlushChannels

// handleData scans for any form of data like stream Warning Messages, Funding, kline events and orderbook actions
// For Funding you’ll need to consider how funds are rolled over, which will affect trading strategies work out if it shuts exchange
// Evaluate subscriptions changes for stops, liquidation etc. If applicable, send to trade module see if the type is unrecognized
// If this is the case, you’ll want to make notes about it. Mostly like for like exchange subscription errors.
// This type is only used when you’re not able to decipher what’s happened return an error if you need to

// handleData function is a matcher for incoming subscription messages from websockets. Messages of different classes below websocket type are matched.
// These messages will be paired with a strategy to be sent on a go channel set up by dealer.
func handleData(d *Dealer, e exchange.IBotExchange, s Strategy, data interface{}) error {
	switch x := data.(type) {
	case string:
		unhandledType(data, true)
	case error:
		return x
	case stream.FundingData:
		handleError("OnFunding", s.OnFunding(d, e, x))
	case *ticker.Price:
		handleError("OnPrice", s.OnPrice(d, e, *x))
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
	case []trade.Data:
		handleError("OnTrade", s.OnTrade(d, e, x))
	case []fill.Data:
		handleError("OnFill", s.OnFill(d, e, x))
	default:
		handleError("OnUnrecognized", s.OnUnrecognized(d, e, data))
	}

	return nil
}

// handleError function checks to see if there are actually an error. This is triggered by the inclusion of the string "err" being != to the string "nil".
// If it does not equal nil, this means there is an error, so it will print out the method responsible for the error along with the error itself.
// If this is true, Go will output "error: <errormessage>". Otherwise, nothing is outputted.
func handleError(method string, err error) {
	if err != nil {
		logrus.Warnf("method %v error: %s\n", method, err)
	}
}

// OpenWebsocket function is responsible for opening a Websocket connection.
// The `req.Exchange.GetWebsocket` method performs the actual functionality of creating a new Websocket in a struct.
// Understandably, if a connection can't be opened, the error would be returned.
// OpenWebsocket checks if the client is sending a query to the interface instance, blocking on a channel, which was a select on a chan
// while the bolt process has a rule queuing, while the rest of the engine while be blocked on a chan
// So, non-blocking options were applied to get a non-blocked client and a non-blocked engine.
func OpenWebsocket(e exchange.IBotExchange) (*stream.Websocket, error) {
	// Check whether websocket is enabled.
	if !e.IsWebsocketEnabled() {
		return nil, ErrWebsocketNotEnabled
	}

	// Check whether websocket is supported.
	if !e.SupportsWebsocket() {
		return nil, ErrWebsocketNotSupported
	}

	// Instantiate a websocket.
	ws, err := e.GetWebsocket()
	if err != nil {
		return nil, err
	}

	// connect
	if !ws.IsConnecting() && !ws.IsConnected() {
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

// unhandledType function displays debug information about the error. It simply formats the given error message.
func unhandledType(data interface{}, warn bool) {
	e := log.Debug()
	if warn {
		e = log.Warn()
	}

	t := fmt.Sprintf("%T\n", data)
	e = e.Interface("data", data).Str("type", t)

	logrus.Warnf("unhandledType: %v\n", e)
}
