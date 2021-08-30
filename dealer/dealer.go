package dealer

import (
	"errors"
	"github.com/romanornr/autodealer/util"
	"github.com/sirupsen/logrus"
	"github.com/thrasher-corp/gocryptotrader/common"
	"github.com/thrasher-corp/gocryptotrader/config"
	"github.com/thrasher-corp/gocryptotrader/currency"
	"github.com/thrasher-corp/gocryptotrader/engine"
	exchange "github.com/thrasher-corp/gocryptotrader/exchanges"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
	"sync"
)

var ErrOrdersAlreadyExists = errors.New("order already exists")

// Dealer If this property in the program that shows the demand of a crypto on an exchange

// Dealer is the main struct for the dealer
// It contains the root strategy, the settings, the config, the exchange manager, and the order registry
type Dealer struct {
	// Root RootStrategy
	Settings        engine.Settings
	Config          config.Config
	ExchangeManager engine.ExchangeManager
	// registry OrderRegistry
}

// “Dealer”, is getting injected with basic configuration properties such as an engine.Settings, login credentials and persistence information
// The root strategy is a used as a placeholder for future functions. A registry object is created and set as a property of this object
// The program gets the keys of an "Assets Currency" table from the configuration table in gocryptotrader.conf
// a green flag that will be passed to the config.DEALER.SETTINGS.ACCOUNT when the new config.DEALER.DefaultCurrency property is evaluated to initialize the system?

// *"Continuous Poller"
// The Keep object will reset the previous values in Engine and request a New
// Initialization every time a New Keep object is instantiated with a new root strategy
// ******************************
// The allocated memory for any other object will be reused for instances of the dealer.RootStrategy methods
// The new config.dealer.defaultCurrency property for the config.DEALER.SETTINGS.ACCOUNT will be initialized using the dealer.Config.DefaultCurrency property
// The suggested name of the property is in order to remove confusion about the green flag and make it more intuitive
// There is a possibility for a "synchronize up to X times a second property that can be passed

// 1. First you create a variable called `dealer` which is of type `*Dealer` This refers to the above struct
// 2. Next is the declaration of NewDealer. Notice that `settings` is of type engine.Settings, whereas `dealer` is of type `*Dealer`.
// This means that NewDealer may take any struct type as its first parameter (where the struct needs to fulfill the Signature of engine.Settings such as
// having a ConfigFile or a EnableDryRun field), and that dealer is a pointer to a Dealer struct, not a naked Dealer. Note also that the returned value
// is a pointer to Dealer as well. In other words: NewDealer takes any settings struct, and returns a pointer to a Dealer struct satisfying some arbitrary interface.
// The interface is satisfied just by having a field called `Root` of type RootStrategy.
// 3. Here we set the defaults for the engine.Settings (which is the incoming struct type), and instantiate the The configuration struct type.
// 4. Reading the configuration from file happens in two parts: First we get the default path to the file using GetAndMigrateDefaultPath, then we set ReadConfigFromFile by the path.
// 5. Dealer will request a new initialization every time a new Dealer object has been initiated with a new root strategy
// 6. The allocated memory for any object will be reused for different instances of the Dealer.NewRootStrategy methods

func New(settings engine.Settings) (*Dealer, error) {
	settings.ConfigFile = util.ConfigFile(settings.ConfigFile)
	var conf config.Config
	dealer := &Dealer{
		Settings:        engine.Settings{},
		Config:          conf,
		ExchangeManager: *engine.SetupExchangeManager(),
	}
	filePath, err := config.GetAndMigrateDefaultPath(dealer.Settings.ConfigFile)
	if err != nil {
		return dealer, err
	}
	logrus.Infof("Loading configuration from %s", filePath)

	if err := dealer.Config.ReadConfigFromFile(filePath, dealer.Settings.EnableDryRun); err != nil {
		return dealer, err
	}

	if err := dealer.SetupExchanges(GCTLog{nil}); err != nil {
		return dealer, err
	}

	return dealer, nil
}

// GCTLog struct has functions for each log type - the Warnf(), Errorf(), and Debugf() functions. The LoadExchange() method for Keep wants an *out log pointer of GCTLog type.
// The bot variable is an interface which does not contain a struct that has methods for each log type and variable has to be changed to a struct for GCTLog type or a new struct needs to be created that has functions for each log type and use that as the input for LoadExchange().
// For this code, it is preferred that GCTLog struct is changed to a struct of a log type
type GCTLog struct {
	ExchangeSys interface{}
}

func (g GCTLog) Warnf(_ interface{}, data string, v ...interface{}) {
	logrus.Errorf(data, v...)
}

func (g GCTLog) Errorf(_ interface{}, data string, v ...interface{}) {
	logrus.Errorf(data, v...)
}

func (g GCTLog) Debugf(_ interface{}, data string, v ...interface{}) {
	logrus.Debugf(data, v...)
}

func (bot *Dealer) LoadExchange(name string, wg *sync.WaitGroup) error {
	return bot.loadExchange(name, wg, GCTLog{nil})
}

var (
	ErrNoExchangesLoaded    = errors.New("no exchanges have been loaded")
	ErrExchangeFailedToLoad = errors.New("exchange failed to load")
)

// LoadExchange creates an exchange object for the loaded exchange by calling ExchangeManager.NewExchangeByName.
// We check that the exchange loaded supports the expected base currency by calling CurrencyPairs.IsAssetEnabled.
// call to the exchange object's Setup function which checks the exchange for its name and retrieves all the configurable values for the exchange. Setup is called by both the ExchangeManager and the Base.
// call to validate credentials, which checks whether or not the exchange supports the asset's currency. If validation is successful, we log an INFO message and pass.
// check the actual auth status of the exchange and make sure that there is no mismatch between the configured auth and the actual auth. If there is a mismatch with isAuthenticatedSupport and AuthenticatedSupport status, we log a WARN message and set the AutheticatedSupport attributes to false.
// We test exchange name is set correctly and make sure that the exchange is set up  normal and we then start the exchange. This last step is performed by both the ExchangeManager and the Base.
func (bot *Dealer) loadExchange(name string, wg *sync.WaitGroup, gctlog GCTLog) error {
	exch, err := bot.ExchangeManager.NewExchangeByName(name)
	if err != nil {
		return err
	}
	if exch.GetBase() == nil {
		return ErrExchangeFailedToLoad
	}

	var localWG sync.WaitGroup
	localWG.Add(1)
	go func() {
		exch.SetDefaults()
		localWG.Done()
	}()
	exchCfg, err := bot.Config.GetExchangeConfig(name)
	if err != nil {
		return err
	}

	if bot.Settings.EnableAllPairs &&
		exchCfg.CurrencyPairs != nil {
		assets := exchCfg.CurrencyPairs.GetAssetTypes(false)
		for x := range assets {
			var pairs currency.Pairs
			pairs, err = exchCfg.CurrencyPairs.GetPairs(assets[x], false)
			if err != nil {
				gctlog.Errorf(gctlog.ExchangeSys,
					"%s: Failed to get pairs for asset type %s. Error: %s\n", exch.GetName(),
					assets[x].String(),
					err)
				return err
			}
			exchCfg.CurrencyPairs.StorePairs(assets[x], pairs, true)
		}
	}

	if bot.Settings.EnableExchangeVerbose {
		exchCfg.Verbose = true
	}
	if exchCfg.Features != nil {
		if bot.Settings.EnableExchangeWebsocketSupport &&
			exchCfg.Features.Supports.Websocket {
			exchCfg.Features.Enabled.Websocket = true
		}
		if bot.Settings.EnableExchangeAutoPairUpdates &&
			exchCfg.Features.Supports.RESTCapabilities.AutoPairUpdates {
			exchCfg.Features.Enabled.AutoPairUpdates = true
		}
		if bot.Settings.DisableExchangeAutoPairUpdates {
			if exchCfg.Features.Supports.RESTCapabilities.AutoPairUpdates {
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

	localWG.Wait()
	if !bot.Settings.EnableExchangeHTTPRateLimiter {
		gctlog.Warnf(gctlog.ExchangeSys,
			"Loaded exchange %s rate limiting has been turned off.\n",
			exch.GetName(),
		)
		err = exch.DisableRateLimiter()
		if err != nil {
			gctlog.Errorf(gctlog.ExchangeSys,
				"Loaded exchange %s rate limiting cannot be turned off: %s.\n",
				exch.GetName(),
				err,
			)
		}
	}

	exchCfg.Enabled = true
	err = exch.Setup(exchCfg)
	if err != nil {
		exchCfg.Enabled = false
		return err
	}

	bot.ExchangeManager.Add(exch)
	base := exch.GetBase()
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
		err = exch.ValidateCredentials(useAsset)
		if err != nil {
			gctlog.Warnf(gctlog.ExchangeSys,
				"%s: Cannot validate credentials, authenticated support has been disabled, Error: %s\n",
				base.Name,
				err)
			base.API.AuthenticatedSupport = false
			base.API.AuthenticatedWebsocketSupport = false
			exchCfg.API.AuthenticatedSupport = false
			exchCfg.API.AuthenticatedWebsocketSupport = false
		}
	}

	if wg != nil {
		exch.Start(wg)
	} else {
		tempWG := sync.WaitGroup{}
		exch.Start(&tempWG)
		tempWG.Wait()
	}

	return nil
}

// SetupExchanges will setup all the servers that are needed to log transactions on an exchange.
// The first thing is initialize sync.WaitGroup. The WaitGroup is used to wait for a group of go-routines to finish before continuing.
// Serialize section also has a WaitGroup, but it does not have a defer wg.Done statement.
// If there was no gesture like this at the end of the "for x:" loop the process would be unresponsive because it would not wait for the others go-routines to finish.
// The "admin" bot will try to get a list of all exchanges and look for any that say they are disabled. If an exchange is disabled, the bot will not load it.
// The idea is to have control over what settings are enabled and which ones are not.
// defer wg.Done() statement is to finish executing the go-routine once the for loop has finished. This is important if we are going to do more than one iteration of the loop with different exchanges.

// SetupExchanges is a function that sets up the exchanges
// It takes in a GCTLog struct
// It returns an error
func (bot *Dealer) SetupExchanges(gctlog GCTLog) error {
	var wg sync.WaitGroup
	configs := bot.Config.GetAllExchangeConfigs()

	// DELETED: parameters -> dryRun...()

	for x := range configs {
		if !configs[x].Enabled && !bot.Settings.EnableAllExchanges {
			gctlog.Debugf(gctlog.ExchangeSys, "%s: Exchange support: Disabled\n", configs[x].Name)
			continue
		}
		wg.Add(1)
		go func(c config.ExchangeConfig) {
			defer wg.Done()
			err := bot.LoadExchange(c.Name, &wg)
			if err != nil {
				gctlog.Errorf(gctlog.ExchangeSys, "LoadExchange %s failed: %s\n", c.Name, err)
				return
			}
			gctlog.Debugf(gctlog.ExchangeSys,
				"%s: Exchange support: Enabled (Authenticated API support: %s - Verbose mode: %s).\n",
				c.Name,
				common.IsEnabled(c.API.AuthenticatedSupport),
				common.IsEnabled(c.Verbose),
			)
		}(configs[x])
	}
	wg.Wait()
	if len(bot.ExchangeManager.GetExchanges()) == 0 {
		return ErrNoExchangesLoaded
	}
	return nil
}
