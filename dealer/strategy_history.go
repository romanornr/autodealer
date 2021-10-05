package dealer

import (
	"context"
	"errors"
	exchange "github.com/thrasher-corp/gocryptotrader/exchanges"
	"github.com/thrasher-corp/gocryptotrader/exchanges/account"
	"github.com/thrasher-corp/gocryptotrader/exchanges/order"
	"github.com/thrasher-corp/gocryptotrader/exchanges/orderbook"
	"github.com/thrasher-corp/gocryptotrader/exchanges/stream"
	"github.com/thrasher-corp/gocryptotrader/exchanges/ticker"
	"strings"
	"sync"
	"time"
)

// Historian is a wrapper around a "stateless" function. It basically will pack a function for a certain duration upon a single invocation, and a sender may invoke this repeatedly for as often as need be.
// The goal is to create a rolling history of an object so, say, every minute one may know the average price or ask price or receive quantity since the beginning of its interaction with the system.
// A strong user would not expect a constant lag of history, but certainly a lag in respect to the previous time intervals.
// 2 Types of history objects
// Time ranges are separated in 2 Types:
// 1. interval based history
// These types refer to history objects whose info updates once after an interval of time.
// These were the base implementation since the Binance historical endpoint cursors return all events at once, but then we were required to make history return ALL updates, not ALL events. So the trivia is that the end of the 5-minute interval is not exact. Hence, interval based history is not an accurate model due to this.
// 2. ticker subscribed history
// Subscribe history assumes that for each subscribe/unsubscribe event of `f(x)`, predictor of events is not assumed to be same as previous predictor. I'm not sure of the exact reason, but empirically this seems true. For more details ;P simulate-middlewares.go gives some explanation on histories
// Some changes have been done to the implementation of the Historian from earlier versions of GCT.
// Both versions of the circular array had the same interface, of which only `Len` and `Last` were used of.
// To subscribe history didn't have a verification of subscription changes, and of the interval history, it would shortly recover from a panic
// # Explanation of what the code does
// This code uses internal interfaces of `exchange.IBotExchange` and `Historian`. I will walk you through the code and explain how it works and why we want to use this method instead of submitting events instantly to the Bot.
// ## 1. Architecture
// The Engine serves as a middleware for exchanges and acts as a glue between all components. The Engine backs notifications from exchange implementation.
// * External notices from an exchange implementation are routed from the Engine through a `notifyHandler` or Driver to a hub.
// * Responsible for automatically registering External notices from an exchange implementation to the Engine. This includes subscribing to exchange exchange exchange global ticker streams which basically means subscribing to all market tickers across all tokens.
// * Also responsible for submitting orders (sell or buy) of our bot to the api.
// * The Engine pulls notifications from the notification hub, filters out the ones that are only related to our bot trades and order changes. It then makes sense of these notifications and filters out any that aren't valid for our bot through channels of the exchanges and trade hubs and vice versa.
// And obviously: this means this was designed to work without needing to be connected to the bot and/or thread-safe. This enables any other piece of functionality to publish/subscribe to the Engine notification hub.
// ## 2. Packets
//
// As you probably noticed above, each notification is actually a struct with several features. The Bot primarily needs the price updates. These are primarily sent through the `exchange.IBotExchange` exchange interface `OnPrice()` event.
// Each packet hasn't been designed to be JSON. If you json this context, you probably won't be able to rebuild it back into a format that I would've designed. Tracks is purposefully designed to work with prices and trades instead of orders because orders tend to have a lot of exchange redirects, flags, etc., which just complicates the logic a bit more and makes the code a few times longer than necessary.
// Since prices are sent through the environment, the packages come with static exchange attributes such as format and quality of service (QoS). Since the data is coming from the exchanges and loosely coupled, we're transporting all of these attributes along with our price data. Our exchange is acting as a middleman between the pricing and trade events and our bot, and hence needs to know how and how to send and read exchange prices and trades. This information is frequently sent, and frequently sent across different networks and different kinds of software, so we searched for a solution that can work while using the minimum amount of processing power.
// The dynamic attributes, namely the price "exchange" are represented as strings that are exchange comma-separated values instead of as actual type-ins because:
// * first: JSON is always a string and doesn't populate arrays and other values and first and foremost, we're trying to minimize compile-time and other overhead.
// * secondly: we should never be relying on the availability of the features of our packages, since they can change.

// Array interface serves to represent a dynamic array of unknown type. For the purpose of this solution we
// use a `CircularArray` of `string`. Each Historian Object create and own an instance of such a CircularArray.
// The Dealer holds a totally Stateful History Struct. A `HiStr` keep a pointer to a specific CircularArray Instance.
// The `HiStr` has a LIFO Queue which stores updates according to their insertion order.
type Array interface {
	At(index int) interface{}
	Len() int
	Last() interface{}

	Floats() []float64
	LastFloat() float64
}

// Historian struct creates an array that is responsible for storing data.
// The event this struct executes is `OnOrder`, that appends to the array the order.
// the interval is `OnOrder` is optional; since it returns once every interval
// epoch is an attribute of the struct, its mutability allows us to isolate the state quirks.
// The `state` itself is an interface that allows us to replace with different arrays
type Historian struct {
	f        func(state Array)
	interval time.Duration
	epoch    int64
	state    Array
}

// NewHistorian function above returns a pointer to golang native type Historian, NOT a reference to it.
// This, in turn, returns a pointer to golang native type CircularArray, NOT a reference to a new instance of a CircularArray.
func NewHistorian(interval time.Duration, stateLength int, f func(array Array)) Historian {
	state := NewCircularArray(stateLength)
	return Historian{
		f:        f,
		interval: interval,
		epoch:    0,
		state:    &state,
	}
}

// Push is called once per update, so it is guaranteed to be executed once per interval.
// The push function not update the state of the historian
func (u *Historian) Push(x interface{}) {
	u.state.(*CircularArray).Push(x)
}

// update function is called once every time a new event is fired.

//The strategy is stateful, meaning that it maintains a "history" of previous events.

// Update function will update the underlying circular array elements and perform a callback
// when either: we last updated >= interval ago, or the current state is the first state - we can just update and bind everything new.
func (u *Historian) Update(now time.Time, x interface{}) {
	// If there is an interval specified, we should update once each interval.
	if u.interval != 0 {
		// Compute the current epoch.
		epoch := now.UnixNano() / u.interval.Nanoseconds()
		// If we're in the same epoch as the last update, return.
		if u.epoch == epoch {
			return
		}
	}
	u.state.(*CircularArray).Push(x)
	u.f(u.state)
}

// Floats returns the underlying array, but cast to []float64, for easy sorting/graphing, without any modifications.
func (u *Historian) Floats() []float64 {
	return u.state.Floats()
}

// +-----------------+
// | HistoryStrategy |
// +-----------------+

var ErrUnknownEvent = errors.New("unknown event")

// HistoryStrategy struct is to reduce the amount of data processed by the bot. The idea is to gather the required data points (e.g. OHLCV)
// and store them in `Historian` units (see `NewHistorian` above) instead of asking exchanges to provide the data directly.
type HistoryStrategy struct {
	// mutex ensure write serialization of onPriceUnits/onOrderUnits
	mu           sync.Mutex
	onPriceUnits map[string][]*Historian
	onOrderUnits map[string][]*Historian
}

// NewHistoryStrategy defines Two maps, one for the History event on the Price field, and one for the Order event.
func NewHistoryStrategy() HistoryStrategy {
	return HistoryStrategy{
		mu:           sync.Mutex{},
		onPriceUnits: make(map[string][]*Historian),
		onOrderUnits: make(map[string][]*Historian),
	}
}

func (r *HistoryStrategy) BindOnPrice(unit *Historian) {}

// AddHistorian function is called at exchange initialization. First, it creates internal data structures necessary for the history to work.
// AddHistorian function is very useful when you expect the same event to be triggered constantly over a certain time interval.
// `onPriceUnits` member is a map of arrays of historians. Each array contains exactly one historian, since every history strategy needs to be attached to a single exchange.
// `onOrderUnits` member is a map of arrays of historians. Each array contains exactly one historian, since every history strategy needs to be attached to a single exchange.
func (r *HistoryStrategy) AddHistorian(exchangeName, eventName string, interval time.Duration, stateLength int, f func(array Array)) error {
	key := strings.ToLower(exchangeName)
	historian := NewHistorian(interval, stateLength, f)

	r.mu.Lock()
	defer r.mu.Lock()

	switch eventName {
	case "OnPrice":
		xs := r.onPriceUnits[key]
		r.onOrderUnits[key] = append(xs, &historian)
	case "OnOrder":
		xs := r.onOrderUnits[key]
		r.onOrderUnits[key] = append(xs, &historian)
	default:
		return ErrUnknownEvent
	}
	return nil
}

// +----------+
// | Strategy |
// +----------+

// Init function of the HistoryStrategy registers the onPriceUnits, onOrderUnits, onPriceUnitsPerCurrency, onOrderUnitsPerCurrency collectors.
func (r *HistoryStrategy) Init(ctx context.Context, d *Dealer, e exchange.IBotExchange) error {
	key := strings.ToLower(e.GetName())
	r.mu.Lock()
	defer r.mu.Unlock()
	r.onPriceUnits[key] = make([]*Historian, 0)
	r.onOrderUnits[key] = make([]*Historian, 0)
	return nil
}

func (r *HistoryStrategy) OnFunding(d *Dealer, e exchange.IBotExchange, x stream.FundingData) error {
	return nil
}

func (r *HistoryStrategy) OnPrice(d *Dealer, e exchange.IBotExchange, x ticker.Price) error {
	lastUpdated := x.LastUpdated
	if lastUpdated.IsZero() {
		lastUpdated = time.Now()
	}
	return fire(r.onPriceUnits, e, lastUpdated, x)
}

func (r *HistoryStrategy) OnKline(d *Dealer, e exchange.IBotExchange, x stream.KlineData) error {
	return nil
}

func (r *HistoryStrategy) OnOrderBook(d *Dealer, e exchange.IBotExchange, x orderbook.Base) error {
	return nil
}

func (r *HistoryStrategy) OnOrder(d *Dealer, e exchange.IBotExchange, x order.Detail) error {
	return fire(r.onOrderUnits, e, x.Date, x)
}

func (r *HistoryStrategy) OnModify(d *Dealer, e exchange.IBotExchange, x order.Modify) error {
	return nil
}

func (r *HistoryStrategy) OnBalanceChange(d *Dealer, e exchange.IBotExchange, x account.Change) error {
	return nil
}

func (r *HistoryStrategy) OnUnrecognized(d *Dealer, e exchange.IBotExchange, x interface{}) error {
	return nil
}

func (r *HistoryStrategy) Deinit(d *Dealer, e exchange.IBotExchange) error {
	return nil
}

func fire(units map[string][]*Historian, e exchange.IBotExchange, now time.Time, x interface{}) error {
	key := strings.ToLower(e.GetName())

	// MT note: if historians do not get added and removed dynamically, this method is
	// completely safe to be used in a MT environment, because:
	//   1. reading (without concurrent writing) a map is MT-safe,
	//   2. all On*() events for a single exchange are invoked from the same thread.
	for _, unit := range units[key] {
		unit.Update(now, x)
	}
	return nil
}