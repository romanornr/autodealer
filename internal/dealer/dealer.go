package dealer

import (
	"context"
	"errors"
	"fmt"
	"github.com/rs/zerolog/log"
	gctlog "github.com/thrasher-corp/gocryptotrader/log"
	"sync"
	"time"

	util2 "github.com/romanornr/autodealer/internal/util"
	"github.com/thrasher-corp/gocryptotrader/currency"
	"github.com/thrasher-corp/gocryptotrader/exchanges/order"

	"github.com/sirupsen/logrus"
	"github.com/thrasher-corp/gocryptotrader/config"
	"github.com/thrasher-corp/gocryptotrader/engine"
	exchange "github.com/thrasher-corp/gocryptotrader/exchanges"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
)

var ErrNoAssetType = errors.New("asset type not associated with currency pair")

const (
	defaultWebsocketTrafficTimeout = time.Second * 30
)

type (
	AugmentConfigFunc func(config *config.Config) error
)

// Builder struct holds state. In this case it specifically has a definition function Augment().
// It also stores internal values such as the path the configs will be read from, the closures/recipe function it will use while conditioning config values.
// Our Augment() will run before the Build() code is called. In this case, the config itself may have been read from a filepath.
// In this case, our function augments with a value from the dealerBuilder struct/obj. That function is then run, and various further functions that had been defined for this.
// Finally, if a file wasn't found, one of the directives within our builder will be constructed a new default template as an alternative to the expected input not found (the expected input "initial")
type Builder struct {
	augment            AugmentConfigFunc
	balanceRefreshRate time.Duration
	factory            ExchangeFactory
	settings           engine.Settings
	reporters          []Reporter
}

// NewBuilder returns a new or configured keep builder
// NewBuilder function is used to build the Dealer object. When we call `dealer, err := builder.build()` we get a *dealer and an error back.
func NewBuilder() *Builder {
	var settings engine.Settings
	return &Builder{
		augment:            nil,
		balanceRefreshRate: 15,
		factory:            nil,
		settings:           settings,
		reporters:          []Reporter{},
	}
}

// Augment augments the exposed functions of the generated service interface, change this to modify the exposed service definition
// Augment function, which you can change, compiles any code in the go code in this environment upon running application.
func (b *Builder) Augment(f AugmentConfigFunc) *Builder {
	b.augment = f
	return b
}

// Balances function will set the refresh interval to fetch balances which in our could be 10 seconds for example.
// This interval will determine how often the bot will make an API call for fetching latest prices.
// My assumption is that it's done 10 seconds to minimize amount of data fetched in-between trades to reduce fees.
func (b *Builder) Balances(refresh time.Duration) *Builder {
	b.balanceRefreshRate = refresh
	return b
}

// CustomExchange function exports a variable called "CustomExchange", which will be registered in each pair's "ExchangeCreatorFunc".
// Via *Builder's CustomExchange function we can insert a custom Exchange Creator.
// Since the name is the only customization we have with this builder, we have to have a factory interface which can instantiate ExchangeCreatorFunc.
func (b *Builder) CustomExchange(name string, f ExchangeFactory) *Builder {
	b.factory = f
	return b
}

// Settings can be used to construct custom settings for the exchange. Since it is optional, the configuration would only have the parts by being assigned in code.
func (b *Builder) Settings(s engine.Settings) *Builder {
	b.settings = s
	return b
}

func (b *Builder) Reporter(r Reporter) *Builder {
	b.reporters = append(b.reporters, r)
	return b
}

// Build constructs a new dealer from a provides settings template argument. In case the provided template is nil, a default template will be used as a starting point.
// In both cases the template config file the read from filepath. If the filepath can not be read, it may be imported from the directory.
// Only if no filepath is provided, filepath.DefaultConfig will be used as an initial config for this dealer
// If the template is a zero value but non nil, a default template will be returned. If the template config is non-zero, it will be constructed from template itself.
// In case the template config can't be successfully constructed, an error will be returned. Along with a resulting *Dealer instance
// This function will also return an error.
func (b Builder) Build() (*Dealer, error) {
	b.settings.ConfigFile = util2.ConfigFile(b.settings.ConfigFile)
	filePath, err := config.GetAndMigrateDefaultPath(b.settings.ConfigFile)
	if err != nil {
		return nil, err
	}

	var (
		conf   config.Config
		dealer = &Dealer{
			Settings:        b.settings,
			Config:          conf,
			ExchangeManager: *engine.SetupExchangeManager(),
			Root:            NewRootStrategy(),
			registry:        *NewOrderRegistry(),
			reporters:       b.reporters,
		}
	)

	// Add history strategy: a special type of strategy that may keep multiple channels of historical data available
	hist := NewHistoryStrategy()
	dealer.Root.Add("history", &hist)

	// Optionally add the balances strategy that keeps track of available balances per exchange.
	if b.balanceRefreshRate > 0 {
		dealer.Root.Add("balances", NewBalancesStrategy(b.balanceRefreshRate))
	}

	logrus.Infof("loading configuration file %s\n", filePath)
	if err := dealer.Config.ReadConfigFromFile(filePath, b.settings.EnableDryRun); err != nil {
		return nil, err
	}

	// Optionally augment config.
	if b.augment != nil {
		if err := b.augment(&dealer.Config); err != nil {
			return dealer, err
		}
	}

	// Assign custom exchange builder.
	dealer.ExchangeManager.Builder = b.factory

	// enable GCT's verbose output through our logging system
	if b.settings.Verbose {
		dealer.setupGCTLogging()
		gctlog.Infoln(gctlog.Global, "GCT logging initialized")
	}

	// Create and setup exchanges.
	if err := dealer.setupExchanges(); err != nil {
		return dealer, err
	}

	return dealer, nil
}

var ErrOrdersAlreadyExists = errors.New("order already exists")

// Dealer struct holds all the information about an instance of an autodealer (`root`, `settings`, `config`, `httpFactory`, `wg`, `ctx`, `exchangeManager`).
// It has inner structs which are instances of ExchangeManager, WithdrawManager.  It has functions such as NewExchangeManager() and return an instance of ExchangeManager.
// This is used for looking up exchange support to enable, and we control it through NewExchangeManager() and WithdrawManager instance.
type Dealer struct {
	Root            RootStrategy
	Settings        engine.Settings
	Config          config.Config
	ExchangeManager engine.ExchangeManager
	registry        OrderRegistry
	reporters       []Reporter
}

//Run is the entry point of all exchange data streams.  Strategy.On*() events for a single exchange are invoked from the same thread.
//Thus, if a strategy deals with multiple exchanges simultaneously, there may be race conditions.
func (bot *Dealer) Run(ctx context.Context) {
	var wg sync.WaitGroup

	exchgs, err := bot.ExchangeManager.GetExchanges()
	if err != nil {
		panic(err)
	}

	for _, x := range exchgs {
		wg.Add(1)

		go func(x exchange.IBotExchange) {
			defer wg.Done()

			// fetch the root strategy
			s := &bot.Root

			// Init root strategy for this exchange.
			if err := s.Init(ctx, bot, x); err != nil {
				panic(fmt.Errorf("failed to initialize strategy: %w", err))
			}

			// go into an infinite loop, either handling websocket
			// events or just plain blocked when there are none
			err := Loop(ctx, bot, x, s)
			// nolint: godox
			// TODO: handle err on terminate when context gets cancelled

			// Deinit root strategy for this exchange.
			if err := s.Deinit(bot, x); err != nil {
				panic(err)
			}

			// This function is never expected to return.  I'm panic()king
			// just to maintain the invariant.
			panic(err)
		}(x)
	}

	wg.Wait()
}

// Loop function is the main entry point for the bot. It is responsible for handling the websocket connection and dispatching events to the appropriate
// strategy. The Loop function is an infinite loop that blocks until the websocket connection is closed. It is expected to be run in a goroutine.
func Loop(ctx context.Context, d *Dealer, e exchange.IBotExchange, s Strategy) error {
	// if the exchange doesn't support websockets we still need to keep running
	if !e.IsWebsocketEnabled() {
		What(log.Warn().Str("exchange", e.GetName()), "no websocket support")
		<-ctx.Done()
		return nil
	}
	// this exchanges does support websockets, go into an
	// infinite loop of receiving/handling messages
	return Stream(ctx, d, e, s)
}

func (bot *Dealer) AddHistorian(
	exchangeName,
	eventName string,
	interval time.Duration,
	stateLength int,
	f func(Array),
) error {
	strategy, err := bot.Root.Get("history")
	if err != nil {
		return err
	}

	hist, ok := strategy.(*HistoryStrategy)
	if !ok {
		panic("")
	}

	return hist.AddHistorian(exchangeName, eventName, interval, stateLength, f)
}

// GetOrderValue function retrieves order details from the given bot's store.
func (bot *Dealer) GetOrderValue(exchangeName, orderID string) (OrderValue, bool) {
	return bot.registry.GetOrderValue(exchangeName, orderID)
}

// getExchange function returns an interface to IBotExchange from either an instance or a name of an exchange
func (bot *Dealer) getExchange(x interface{}) exchange.IBotExchange {
	switch x := x.(type) {
	case exchange.IBotExchange:
		return x
	case string:
		e, err := bot.ExchangeManager.GetExchangeByName(x)
		if err != nil {
			panic(fmt.Sprintf("unable to find %s exchange\n", x))
		}
		return e
	default:
		panic("exchangeOrName should be either an instance of exchange.IBotExchange or a string\n")
	}
}

// +----------------------+
// | Keep: Exchange state |
// +----------------------+

// GetActiveOrders function is a wrapper around the GetActiveOrders function in the exchange package.
func (bot *Dealer) GetActiveOrders(ctx context.Context, exchangeOrName interface{}, request order.GetOrdersRequest) ([]order.Detail, error) {
	e := bot.getExchange(exchangeOrName)

	timer := time.Now()

	defer bot.ReportLatency(GetActiveOrdersLatencyMetric, timer, e.GetName())

	resp, err := e.GetActiveOrders(ctx, &request)
	if err != nil {
		bot.ReportEvent(GetActiveOrdersErrorMetric, e.GetName())
		return resp, err
	}
	return resp, nil
}

// +------------------------+
// | Keep: Order submission |
// +------------------------+

// SubmitOrder function, simply makes an Order.Submit which contains the giving parameters and the current name of requested Exchange
// Submit it to the giving requested Exchange, called at the user orders code. It returns the order map response if successful, otherwise error.
func (bot *Dealer) SubmitOrder(ctx context.Context, exchangeOrName interface{}, submit order.Submit) (order.SubmitResponse, error) {
	return bot.SubmitOrderUD(ctx, exchangeOrName, submit, nil)
}

// SubmitOrderUD is similar to the SubmitOrder, except that this function also adds the order map into the Orders map
// and its corresponding ID and name and then return it and its error and additional notes and errors which cause the metric to move asynchronous processing.
func (bot *Dealer) SubmitOrderUD(ctx context.Context, exchangeOrName interface{}, submit order.Submit, userData interface{}) (order.SubmitResponse, error) {
	e := bot.getExchange(exchangeOrName)

	// Make sure order.Submit.Exchange is properly populated
	if submit.Exchange == "" {
		submit.Exchange = e.GetName()
	}

	bot.ReportEvent(SubmitOrderMetric, e.GetName())

	defer bot.ReportLatency(SubmitOrderLatencyMetric, time.Now(), e.GetName())
	resp, err := e.SubmitOrder(ctx, &submit)
	if err != nil {
		// post an error metric event
		bot.ReportEvent(SubmitOrderErrorMetric, e.GetName())
		return resp, err
	}

	// store the order in the registry
	if !bot.registry.Store(e.GetName(), resp, userData) {
		return resp, ErrOrdersAlreadyExists
	}
	return resp, err
}

// SubmitOrders method calls the SubmitOrder method then Contains method to check for an exchage name in xs slice.
// If contains is true, you will return an error since an exchange was reported. If contains is false, you will continue with the execution of the function.
// ListOrder method will not be executed if Contains method returns an error.
func (bot *Dealer) SubmitOrders(ctx context.Context, e exchange.IBotExchange, xs ...order.Submit) error {
	var wg util2.ErrorWaitGroup
	bot.ReportEvent(SubmitBulkOrderLatencyMetric, e.GetName())
	defer bot.ReportLatency(SubmitBulkOrderLatencyMetric, time.Now(), e.GetName())

	for _, x := range xs {
		wg.Add(1)

		go func(x order.Submit) {
			_, err := bot.SubmitOrder(ctx, e, x)
			wg.Done(err)
		}(x)
	}
	return wg.Wait()
}

// ModifyOrder method calls the SubmitOrder method then Contains method to check for an exchange name in xs slice.
// If contains is true, you will return an error since an exchange was reported. If contains is false, you will continue with the execution of the function.
// CreateOrder method will not be executed if Contains method returns an error.
func (bot *Dealer) ModifyOrder(ctx context.Context, exchangeOrName interface{}, mod order.Modify) (order.Modify, error) {
	e := bot.getExchange(exchangeOrName)
	bot.ReportEvent(ModifyOrderMetric, e.GetName())

	defer bot.ReportLatency(ModifyOrderLatencyMetric, time.Now(), e.GetName())

	resp, err := e.ModifyOrder(ctx, &mod)
	if err != nil {
		bot.ReportEvent(ModifyOrderErrorMetric, e.GetName())
		return resp, err
	}
	return resp, nil
}

// CancelOrder calls to make an Order.Cancel which includes the giving submitted order, the name to the requested submitted order
// and the current name of Exchange, calls to it to make Order.Cancel, otherwise return error. It returns the waiting for canceled order if successful, otherwise error.
// Filters the Orders map if the given Order ID exists and deletes its map from it.
func (bot *Dealer) CancelOrder(ctx context.Context, exchangeOrName interface{}, x order.Cancel) error {
	e := bot.getExchange(exchangeOrName)
	if x.Exchange == "" {
		x.Exchange = e.GetName()
	}

	bot.ReportEvent(CancelOrderMetric, e.GetName())
	defer bot.ReportLatency(CancelOrderLatencyMetric, time.Now(), e.GetName())

	if err := e.CancelOrder(ctx, &x); err != nil {
		bot.ReportEvent(CancelOrderErrorMetric, e.GetName())
		return err
	}
	return nil
}

// +-------------------------+
// | Keep: Event observation |
// +-------------------------+

// OnOrder function calls the GetOrderValue method to see if an order exists with that dealer.
// We have a GetValue method in the handler file. We modify the obtained value by setting its UserData to the OnFilled Observer required to perform the appropriate strategy.
// In n our instance, our order may include two methods. One provides us with a transaction, while the other provides us with profit and loss information (P&L).
// When we get an order, we set the User data to OnFilledObserver using the Value property and leave the handler code. Within the Handler.
// OnFilled, we check two criteria to verify whether they are present in Value in order to optimize the strategy's execution.
func (bot *Dealer) OnOrder(e exchange.IBotExchange, x order.Detail) {
	if x.Status == order.Filled {
		value, ok := bot.GetOrderValue(e.GetName(), x.ID)
		if !ok {
			return
		}

		if obs, ok := value.UserData.(OnFilledObserver); ok {
			obs.OnFilled(bot, e, x)
		}
	}
}

// +----------------------+
// | Metric reports       |
// +----------------------+

// ReportLatency will report the latency of the bot to the metrics server.
func (bot *Dealer) ReportLatency(m Metric, t time.Time, labels ...string) {
	for _, r := range bot.reporters {
		r.Latency(m, time.Since(t), labels...)
	}
}

// ReportEvent will report an event to the metrics server.
func (bot *Dealer) ReportEvent(m Metric, labels ...string) {
	for _, r := range bot.reporters {
		r.Event(m, labels...)
	}
}

// ReportValue will report a value to the metrics server
func (bot *Dealer) ReportValue(m Metric, v float64, labels ...string) {
	for _, r := range bot.reporters {
		r.Value(m, v, labels...)
	}
}

// ActivateAsset will activate the asset for the given exchange.
func (bot *Dealer) ActivateAsset(e exchange.IBotExchange, a asset.Item) error {
	base := e.GetBase()

	if err := base.CurrencyPairs.SetAssetEnabled(a, true); err != nil && !errors.Is(err, currency.ErrAssetAlreadyEnabled) {
		return err
	}
	return nil
}

// ActivatePair will activate the pair for the given exchange.
func (bot *Dealer) ActivatePair(e exchange.IBotExchange, a asset.Item, p currency.Pair) error {
	base := e.GetBase()

	if err := base.CurrencyPairs.IsAssetEnabled(a); err != nil {
		return err
	}

	// updated enabled pairs
	enabledpairs, err := base.CurrencyPairs.GetPairs(a, true)
	if err != nil {
		return err
	}

	enabledpairs = append(enabledpairs, p)
	base.CurrencyPairs.StorePairs(a, enabledpairs, true)

	// updated available pairs
	availablepairs, err := base.CurrencyPairs.GetPairs(a, false)
	if err != nil {
		return err
	}

	availablepairs = append(availablepairs, p)
	base.CurrencyPairs.StorePairs(a, availablepairs, false)

	return nil
}

// GetExchanges is a wrapper of GCT's Engine.GetExchanges
func (bot *Dealer) GetExchanges() []exchange.IBotExchange {
	exch, err := bot.ExchangeManager.GetExchanges()
	if err != nil {
		logrus.Errorf("unable to get exchanges: %s", err)
		return []exchange.IBotExchange{}
	}

	return exch
}

var (
	ErrNoExchangesLoaded    = errors.New("no exchanges have been loaded")
	ErrExchangeFailedToLoad = errors.New("exchange failed to load")
)

// loadExchange is an adapted version of GCT's Engine.LoadExchange.
func (bot *Dealer) loadExchange(exchCfg *config.Exchange, wg *sync.WaitGroup) error {
	exch, err := bot.ExchangeManager.NewExchangeByName(exchCfg.Name)
	if err != nil {
		return err
	}

	base := exch.GetBase()
	if base == nil {
		return ErrExchangeFailedToLoad
	}

	exch.SetDefaults()
	// overwrite whatever name the exchange wrapper has decided with the name that's in the config,
	// this is due to the exchange alias functionality that we offer over GCT.
	base.Name = exchCfg.Name

	if bot.Settings.EnableAllPairs &&
		exchCfg.CurrencyPairs != nil {
		assets := exchCfg.CurrencyPairs.GetAssetTypes(false)
		for x := range assets {
			var pairs currency.Pairs

			pairs, err = exchCfg.CurrencyPairs.GetPairs(assets[x], false)
			if err != nil {
				return err
			}

			exchCfg.CurrencyPairs.StorePairs(assets[x], pairs, true)
		}
	}

	if bot.Settings.EnableExchangeVerbose {
		exchCfg.Verbose = true
	}

	// nolint: nestif
	if exchCfg.Features != nil {
		if bot.Settings.EnableExchangeWebsocketSupport &&
			base.Features.Supports.Websocket {
			exchCfg.Features.Enabled.Websocket = true

			if exchCfg.WebsocketTrafficTimeout <= 0 {
				What(log.Info().
					Str("exchange", exchCfg.Name).
					Dur("default", defaultWebsocketTrafficTimeout),
					"Websocket response traffic timeout value not set, setting default")

				exchCfg.WebsocketTrafficTimeout = defaultWebsocketTrafficTimeout
			}
		}

		if bot.Settings.EnableExchangeAutoPairUpdates &&
			base.Features.Supports.RESTCapabilities.AutoPairUpdates {
			exchCfg.Features.Enabled.AutoPairUpdates = true
		}

		if bot.Settings.DisableExchangeAutoPairUpdates {
			if base.Features.Supports.RESTCapabilities.AutoPairUpdates {
				exchCfg.Features.Enabled.AutoPairUpdates = false
			}
		}
	}

	if bot.Settings.HTTPUserAgent != "" {
		exchCfg.HTTPUserAgent = bot.Settings.HTTPUserAgent
	}

	if bot.Settings.HTTPProxy != "" {
		exchCfg.ProxyAddress = bot.Settings.HTTPProxy
	}

	if bot.Settings.HTTPTimeout != exchange.DefaultHTTPTimeout {
		exchCfg.HTTPTimeout = bot.Settings.HTTPTimeout
	}

	if bot.Settings.EnableExchangeHTTPDebugging {
		exchCfg.HTTPDebugging = bot.Settings.EnableExchangeHTTPDebugging
	}

	if !bot.Settings.EnableExchangeHTTPRateLimiter {
		What(log.Info().
			Str("exchange", exch.GetName()),
			"Rate limiting has been turned off")

		if err := exch.DisableRateLimiter(); err != nil {
			What(log.Error().
				Err(err).
				Str("exchange", exch.GetName()),
				"unable to disable exchange rate limiter")
		}
	}

	exchCfg.Enabled = true

	if err := exch.Setup(exchCfg); err != nil {
		exchCfg.Enabled = false

		return err
	}

	bot.ExchangeManager.Add(exch)

	if base.API.AuthenticatedSupport ||
		base.API.AuthenticatedWebsocketSupport {
		assetTypes := base.GetAssetTypes(false)

		var useAsset asset.Item

		for a := range assetTypes {
			err = base.CurrencyPairs.IsAssetEnabled(assetTypes[a])
			if err != nil {
				continue
			}

			useAsset = assetTypes[a]

			break
		}

		if err := exch.ValidateCredentials(context.TODO(), useAsset); err != nil {
			gctlog.Warnf(gctlog.ExchangeSys,
				"%s: Cannot validate credentials: %s\n",
				base.Name,
				err)
		}
	}

	if wg != nil {
		if err := exch.Start(wg); err != nil {
			return fmt.Errorf("unable to start exchange: %w", err)
		}
	} else {
		tempWG := sync.WaitGroup{}

		if err := exch.Start(&tempWG); err != nil {
			return fmt.Errorf("unable to start exchange: %w", err)
		}

		tempWG.Wait()
	}

	return nil
}

// setupExchanges is an (almost) unchanged copy of Engine.SetupExchanges.
//
//nolint
func (bot *Dealer) setupExchanges() error {
	var wg sync.WaitGroup

	configs := bot.Config.GetAllExchangeConfigs()

	for x := range configs {
		if !configs[x].Enabled && !bot.Settings.EnableAllExchanges {
			What(log.Debug().
				Str("exchange", configs[x].Name),
				"exchange disabled")
			continue
		}
		wg.Add(1)
		go func(c config.Exchange) {
			defer wg.Done()
			err := bot.loadExchange(&c, &wg)
			if err != nil {
				What(log.Error().
					Err(err).
					Str("exchange", c.Name),
					"exchange load failed")

				return
			}

			What(log.Debug().
				Str("exchange", c.Name).
				Bool("authenticated", c.API.AuthenticatedSupport).
				Bool("verbose", c.Verbose),
				"exchange load failed")
		}(configs[x])
	}
	wg.Wait()
	if len(bot.GetExchanges()) == 0 {
		return ErrNoExchangesLoaded
	}
	return nil
}

// GetExchangeByName function returns an exchange interface by name.
func (bot *Dealer) GetExchangeByName(name string) (exchange.IBotExchange, error) {
	return bot.ExchangeManager.GetExchangeByName(name)
}

// GetEnabledPairAssetType function returns a list of enabled asset types for a given exchange.
func (bot *Dealer) GetEnabledPairAssetType(e exchange.IBotExchange, c currency.Pair) (asset.Item, error) {
	b := e.GetBase()

	assetTypes := b.GetAssetTypes(true)
	for i := range assetTypes {
		enabled, err := b.GetEnabledPairs(assetTypes[i])
		if err != nil {
			return asset.Spot, err
		}

		if enabled.Contains(c, true) {
			return assetTypes[i], nil
		}
	}
	return asset.Spot, ErrNoAssetType
}
