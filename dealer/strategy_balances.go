package dealer

//var (
//	ErrAccountIndexOfRange = errors.New("no account with this index exist")
//	ErrCurrencyNotFound = errors.New("currency not found in holdings")
//	ErrHoldingsNotFound = errors.New("holdings not found")
//)
//
//type BalancesStrategy struct {
//	balances sync.Map
//	ticker TickerStrategy
//}
//
////NewBalancesStrategy creates a new *BalancesStrategy using a given new TickerStrategy
//func NewBalancesStrategy(refreshRate time.Duration) Strategy {
//	b := &BalancesStrategy{
//		balances: sync.Map{},
//		ticker:   TickerStrategy{
//			Interval: refreshRate,
//			TickFunc: nil,
//			tickers: sync.Map{},
//		},
//	}
//	b.ticker.TickFunc = b.tick
//	return b
//}
//
//// tick runs at the interval given by RefreshRate. It updates the balance for the passed Owner.
//func (b *BalancesStrategy) tick(d *Dealer, e exchange.IBotExchange) {
//	holdings, err := e.UpdateAccountInfo(asset.Spot)
//	if err != nil {
//		logrus.Errorf("exchange. %s\n", e.GetName())
//	} else {
//		b.Store(holdings)
//	}
//}
//
//// Store stores the holdings from an account on a given exchange
//func (b *BalancesStrategy) Store(holdings account.Holdings) {
//	b.balances.Store(holdings.Exchange, holdings)
//}
//
//// Load returns an amortized holding from self.holdings of exchange exchangeName of the accountID of the key code of the asset you want
//// in exchange to go by. In loading balance of account accountID of base asset name of code you risk a panic if there is a lack
//// of a BaseBalance on exchange by accountID for base asset name of code.
//func (b *BalancesStrategy) Load(exchangeName string) (holdings account.Holdings, loaded bool) {
//	var ok bool
//	pointer, loaded := b.balances.Load(exchangeName)
//	if loaded {
//		holdings, ok = pointer.(account.Holdings)
//		if !ok {
//			logrus.Panicf("have %T, want account.Holdings", pointer)
//		}
//	}
//	return holdings, ok
//}
//
//// Currency using the literal descriptions of the items in the balances array
//func (b *BalancesStrategy) Currency(exchangeName string, code string, accountID string) (account.Balance, error) {
//	holdings, loaded := b.Load(exchangeName)
//	if !loaded {
//		var empty account.Balance
//		return empty, ErrHoldingsNotFound
//	}
//
//	for _, sub := range holdings.Accounts {
//		if sub.ID == accountID {
//			for _, balance := range sub.Currencies {
//				if balance.CurrencyName.String() == code {
//					return balance, nil
//				}
//			}
//		}
//	}
//	return account.Balance{}, ErrCurrencyNotFound
//}
//
//// +--------------------+
//// | Strategy interface |
//// +--------------------+
//
//// Init sets up the strategy
//func (b *BalancesStrategy) Init(d *Dealer, e exchange.IBotExchange) error {
//	return b.ticker.Init(d, e)
//}
//
//func (b *BalancesStrategy) OnFunding(d *Dealer, e exchange.IBotExchange, x stream.FundingData) error {
//	return nil
//}
//
//func (b *BalancesStrategy) OnPrice(d *Dealer, e exchange.IBotExchange, x ticker.Price) error {
//	return nil
//}
//
//func (b *BalancesStrategy) OnKline(d *Dealer, e exchange.IBotExchange, x stream.KlineData) error {
//	return nil
//}
//
//func (b *BalancesStrategy) OnOrderBook(d *Dealer, e exchange.IBotExchange, x orderbook.Base) error {
//	return nil
//}
//
//func (b *BalancesStrategy) OnOrder(d *Dealer, e exchange.IBotExchange, x order.Detail) error {
//	return nil
//}
//
//func (b *BalancesStrategy) OnModify(d *Dealer, e exchange.IBotExchange, x order.Modify) error {
//	return nil
//}
//
//func (b *BalancesStrategy) OnBalanceChange(d *Dealer, e exchange.IBotExchange, x account.Change) error {
//	return nil
//}
//
//func (b *BalancesStrategy) OnUnrecognized(d *Dealer, e exchange.IBotExchange, x interface{}) error {
//	return nil
//}
//
//func (b *BalancesStrategy) Deinit(d *Dealer, e exchange.IBotExchange) error {
//	return b.ticker.Init(d, e)
//}
