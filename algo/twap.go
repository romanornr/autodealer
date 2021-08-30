package algo

import (
	"fmt"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"github.com/thrasher-corp/gocryptotrader/currency"
	exchange "github.com/thrasher-corp/gocryptotrader/exchanges"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
	"github.com/thrasher-corp/gocryptotrader/exchanges/ftx"
	"github.com/thrasher-corp/gocryptotrader/exchanges/order"
	"gopkg.in/errgo.v2/fmt/errors"
	"math/big"
	"math/rand"
	"net/http"
	"time"
)

type TWAP struct {
	Exchange     exchange.IBotExchange
	Pair         currency.Pair
	Asset        asset.Item // SPOT, FUTURES, INDEX
	MaxChangePct float64
	Start        time.Time
	End          time.Time
	WapPrice     float64
	OverBought   bool
	Side         order.Side
}

func (t TWAP) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (t TWAP) Handle() error {
	time.Sleep(time.Minute * 5)
	return nil
}

//type WAPService struct {
//	Queue chan queue.Queuable
//}
//
//func NewWAPService(q chan queue.Queuable) *WAPService {
//	service := &WAPService{
//		Queue: q,
//	}
//	return service
//}

func (t *TWAP) Execute(targetAmount float64, baseMinSize, baseMaxSize float64, mode string) {
	rand.Seed(time.Now().Unix())
	var wapAmount decimal.Decimal
	var filledInTotal decimal.Decimal

	logrus.Infof("Starting execution")
	logrus.Info("starting TWAP....")

	// stop loop when end twap end time is over current time
	for t.End.Unix() > time.Now().Unix() {
		rand.Seed(time.Now().UnixNano())
		randomMinutes := rand.Intn(5-1) + 1
		randomSeconds := time.Duration(rand.Intn(60-1)+1) * time.Second

		// keep track of total filled quantity of the TWAP
		if filledInTotal.GreaterThanOrEqual(decimal.NewFromFloat(targetAmount)) {
			logrus.Infof("TWAP finished: target amount %.8f filled: %s\n", targetAmount, filledInTotal.String())
			break
		}

		// next execution is some random minute and random second
		nextExecution := time.Duration(randomMinutes)*time.Minute + randomSeconds

		// Calculate the quantity that has to be filled on average per second so it can reach the target before the TWAP end time
		wapAmount = t.AverageSizeFillPerSecond(decimal.NewFromFloat(targetAmount)).Mul(decimal.NewFromFloat(nextExecution.Seconds()))
		logrus.Infof("TWAP quantity %s for next execution: %s\n:", t.Pair.Base, wapAmount.String())

		f := ftx.FTX{Base: *t.Exchange.GetBase()}

		f.GetMarket("SOL")
		tickerInfo, err := t.Exchange.FetchTicker(t.Pair, t.Asset)
		if err != nil {
			logrus.Errorf("fetch ticker failed: %s\n", err)
		}

		// estimate how much USD/USDT/BTC is used for next execution
		logrus.Infof("Estimated amount of %s used for next execution: %s", t.Pair.Quote, wapAmount.Mul(decimal.NewFromFloat(tickerInfo.Last)).String())

		//if wapAmount.LessThanOrEqual(decimal.NewFromFloat(baseMinSize)) {
		//	for wapAmount.LessThanOrEqual(decimal.NewFromFloat(baseMinSize)) {
		//		nextExecution = +nextExecution
		//		wapAmount = wapAmount.Add(t.AverageSizeFillPerSecond(decimal.NewFromFloat(targetAmount).Sub(filledInTotal)).Mul(decimal.NewFromFloat(nextExecution.Seconds())))
		//	}
		//	logrus.Infof("Increased WAP amount to reach minimal size: %s\n", wapAmount.String())
		//}

		if filledInTotal.Add(wapAmount).GreaterThanOrEqual(decimal.NewFromFloat(targetAmount)) {
			wapAmount = decimal.NewFromFloat(targetAmount).Sub(filledInTotal)
		}

		wapResponse, err := t.Submit(tickerInfo.Ask, wapAmount, mode)
		if err != nil {
			logrus.Errorf("wapresponse error: %s\n", err)
			<-time.After(time.Minute * 5)
			continue
		}
		filledInTotal = filledInTotal.Add(wapAmount)
		progress := decimal.NewFromFloat(100.0).Div(decimal.NewFromFloat(targetAmount)).Mul(filledInTotal).Round(2)
		logrus.Infof("Buying %s %s at price %.8f\t%s\t%.2f minutes left\t%s%%\n", wapAmount.String(), t.Pair.String(), wapResponse.Rate, time.Now().Format(time.RFC822), t.End.Sub(time.Now()).Minutes(), progress.String())

		// next amount to buy is determined by averageSize required per minute to meet the target amount multiplied by the minutes waiting for the next order
		logrus.Infof("Next order execution takes place after: %s\n", nextExecution.String())
		<-time.After(nextExecution)
	}
}

func (t TWAP) AverageSizeFillPerMinute(amount decimal.Decimal) decimal.Decimal {
	diff := t.End.Sub(t.Start).Minutes()
	a := amount.Div(decimal.NewFromFloat(diff)) // average size fill per minute
	return a
}

func (t TWAP) AverageSizeFillPerSecond(amount decimal.Decimal) decimal.Decimal {
	diff := t.End.Sub(time.Now()).Seconds()
	a := amount.Div(decimal.NewFromFloat(diff)) // average size fill per minute
	return a
}

func (t TWAP) Submit(price float64, amount decimal.Decimal, mode string) (order.SubmitResponse, error) {
	wapAmount, _ := amount.Float64()
	wapAmount, _ = new(big.Float).SetPrec(2).SetFloat64(wapAmount).Float64()
	logrus.Infof("fetching ticker order from")
	ticker, err := t.Exchange.FetchTicker(t.Pair, t.Asset)
	if err != nil {
		logrus.Errorf("fetch ticker failed: %s\n", err)
		return order.SubmitResponse{}, err
	}

	pctChange := (100.00 / price) * ticker.Last

	if t.MaxChangePct > pctChange {
		return order.SubmitResponse{}, fmt.Errorf("change in price percentage exceeds limit risk %.2f\n", pctChange)
	}

	if t.Side == order.Buy && t.OverBought == true {
		return order.SubmitResponse{}, fmt.Errorf("failed to buy %s because the price is above the TWAP\n", t.Pair.String())
	}

	if t.Side == order.Sell && t.OverBought != true {
		return order.SubmitResponse{}, fmt.Errorf("failed to buy %s because the price is above the TWAP\n", t.Pair.String())
	}

	if t.Side == order.Buy {
		price = ticker.Ask
	}

	if t.Side == order.Sell {
		price = ticker.Bid
	}

	var orderType order.Type
	if mode == "passive" {
		orderType = order.Limit
	}

	if mode == "aggressive" {
		orderType = order.Market
	}

	wapOrder := order.Submit{
		Price:           price,
		Amount:          wapAmount,
		Exchange:        t.Exchange.GetName(),
		InternalOrderID: "",
		Type:            orderType,
		Side:            t.Side,
		AssetType:       t.Asset,
		Pair:            t.Pair,
	}

	wapResponse, err := t.Exchange.SubmitOrder(&wapOrder) // TODO LOT_SIZE MinQty
	if err != nil {
		logrus.Errorf("submit order failed: %s\n", err)
		return order.SubmitResponse{}, errors.Newf("submit oder failed: %s", err)
	}

	wapResponse.Rate = price
	return wapResponse, err
}

//func (w WAPService) Run (t TWAP){
//	w.Queue <- t
//}

//// testing purpose
//func (w WAPService) Run(t TWAP, message chan string) {
//	fmt.Println("start")
//	time.Sleep(time.Second * 3)
//	message <- "executing...."
//	time.Sleep(time.Second * 3)
//	message <- "Almost...."
//	time.Sleep(time.Second * 2)
//	message <- "finished wap"
//	//go twap.Execute()
//}
