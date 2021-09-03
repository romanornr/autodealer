package dealer

import (
	"errors"
	"github.com/sirupsen/logrus"
	exchange "github.com/thrasher-corp/gocryptotrader/exchanges"
	"github.com/thrasher-corp/gocryptotrader/exchanges/stream"
)

var (
	ErrWebsocketNotSupported = errors.New("websocket not supported")
	ErrWebsocketNotEnabled   = errors.New("websocket is not enabled")
)

// Stream opens a websocket connection for the data stream, passing the market data to be processed as it's received.
// data passed to the FSM, there's a corresponding response channel depending on the data passed to it
// routing the messages to the appropriate channels.
func Stream(d *Dealer, e exchange.IBotExchange) error {
	ws, err := OpenWebsocket(e)
		if err != nil {
			return err
		}

		// This goroutine is supposed to never finish
		for data := range ws.ToRoutine {
			data := data
			go func() {
				err := handleData(d, e, data)
				if err != nil {
					logrus.Errorf("error handling data: %s\n", err)
				}
			}()
		}
		panic("unexpected end of channel")
}

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

func handleData(d *Dealer, e exchange.IBotExchange, data interface{}) error {
	//switch x := data.(type) {
	//case string:
	//	unhandledType(data, true)
	//case error:
	//	return x
	//case *ticker.Price:
	//	handleError("OnPrice", s.Onprice(k, e, *x))
	//case stream.KlineData:
	//	handleError("OnKline", s.Onkline(k, e))
	//	default:
	//}
	return nil
}
