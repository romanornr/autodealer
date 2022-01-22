package algo

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hibiken/asynq"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"github.com/thrasher-corp/gocryptotrader/currency"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
	"github.com/thrasher-corp/gocryptotrader/exchanges/order"
	"time"
)

const (
	TypeTwapOrder = "twap"
)

type TwapOrderPayload struct {
	Exchange          string
	AccountID         string
	Pair              currency.Pair
	Asset             asset.Item // SPOT, FUTURES, INDEX
	Start             time.Time
	End               time.Time
	TargetAmountQuote float64
	Side              order.Side
	OrderType         order.Type
	Status            string
}

// NewTwapOrderTask creates a new TwapOrderTask
func NewTwapOrderTask(order TwapOrderPayload) (*asynq.Task, error) {
	payload, err := json.Marshal(order)
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeTwapOrder, payload), nil
}

func NewOrderTask(order order.Submit) (*asynq.Task, error) {
	payload, err := json.Marshal(order)
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeTwapOrder, payload), nil
}

// HandleTwapOrderTask handles a TwapOrderTask
func HandleTwapOrderTask(ctx context.Context, t *asynq.Task) error {
	var p TwapOrderPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}
	p.Status = "processing"
	logrus.Printf("Sending order to %s\n", p.Exchange)
	time.Sleep(60 * time.Second)
	p.Status = "complete"
	logrus.Printf("Order sent to %s\n complete", p.Exchange)
	// TWAP order code ...
	return nil
}

// Execute executes the TwapOrderTask
func Execute(t TwapOrderPayload) {

	// minimal size in dollars for each order
	minimalSize := decimal.NewFromFloat(5)

	// convert the target amount to a decimal
	targetQuote := decimal.NewFromFloat(t.TargetAmountQuote)

	// Calculate the average size of the order in dollars of each minute
	averageSizeFillPerMinute := AverageSizeFillPerMinute(t.Start, t.End, targetQuote)

	// minutes it takes to fill target amount (ie 900 minutes)
	targetQuoteFillTimeMinutes := targetQuote.Div(averageSizeFillPerMinute)

	// Amount of orders it takes to fill target amount
	amountOfOrders := targetQuote.Div(minimalSize)

	// minutes in between each order
	minutesBetweenOrders := targetQuoteFillTimeMinutes.Div(amountOfOrders)

	var nextExecutionTime = t.Start

	logrus.Printf("minutes it takes to fill target amount %s\n: ", targetQuoteFillTimeMinutes.String())
	logrus.Printf("amount of orders it takes to fill target amount %s\n: ", amountOfOrders.String())

	client := asynq.NewClient(asynq.RedisClientOpt{Addr: "localhost:6379"})

	for amountOfOrders.Cmp(decimal.Zero) > 0 {

		// create a new order
		// if nextExecutionTime is before t.End
		// execute order

		// next execution time is nextExecutionTime + minutesBetweenOrders
		nextExecutionTime = nextExecutionTime.Add(time.Minute * time.Duration(minutesBetweenOrders.IntPart()))

		// amountOfOrders--
		amountOfOrders = amountOfOrders.Sub(decimal.NewFromFloat(1))
		logrus.Printf("Order placed in queue. Next execution time: %v\n", nextExecutionTime)

		t1, err := NewOrderTask(order.Submit{})
		if err != nil {
			logrus.Errorf("Error creating order task: %v", err)
		}

		info, err := client.Enqueue(t1, asynq.ProcessAt(nextExecutionTime))
		if err != nil {
			logrus.Errorf("Error enqueuing order task: %v", err)
		}
		logrus.Printf("Order task enqueued: %v\n", info)
	}

}

//// Twap is a twap strategy that will attempt to execute an order and achieve the TWAP or better. A TWAP strategy underpins more sophisticated ways of buying and selling than simply executing orders en masse: for example, dumping a huge number of shares in one block is likely to affect market perceptions, with an adverse effect on the price. A TWAP strategy is often used to minimize a large order's the impact on the market and result in price improvement
//func Twap(director orderbuilder.Director, hours float64, minutes float64) {
//
//}

//import (
//	"fmt"
//	"github.com/shopspring/decimal"
//	"github.com/sirupsen/logrus"
//	"github.com/thrasher-corp/gocryptotrader/currency"
//	exchange "github.com/thrasher-corp/gocryptotrader/exchanges"
//	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
//	"github.com/thrasher-corp/gocryptotrader/exchanges/ftx"
//	"github.com/thrasher-corp/gocryptotrader/exchanges/order"
//	"gopkg.in/errgo.v2/fmt/errors"
//	"math/big"
//	"math/rand"
//	"net/http"
//	"time"
//)
//
//type TWAP struct {
//	Exchange     exchange.IBotExchange
//	Pair         currency.Pair
//	Asset        asset.Item // SPOT, FUTURES, INDEX
//	MaxChangePct float64
//	Start        time.Time
//	End          time.Time
//	WapPrice     float64
//	OverBought   bool
//	Side         order.Side
//}
//
//func (t TWAP) Render(w http.ResponseWriter, r *http.Request) error {
//	return nil
//}
//
//func (t TWAP) Handle() error {
//	time.Sleep(time.Minute * 5)
//	return nil
//}
//
////type WAPService struct {
////	Queue chan queue.Queuable
////}
////
////func NewWAPService(q chan queue.Queuable) *WAPService {
////	service := &WAPService{
////		Queue: q,
////	}
////	return service
////}
//
//func (t *TWAP) Execute(targetAmount float64, baseMinSize, baseMaxSize float64, mode string) {
//	rand.Seed(time.Now().Unix())
//	var wapAmount decimal.Decimal
//	var filledInTotal decimal.Decimal
//
//	logrus.Infof("Starting execution")
//	logrus.Info("starting TWAP....")
//
//	// stop loop when end twap end time is over current time
//	for t.End.Unix() > time.Now().Unix() {
//		rand.Seed(time.Now().UnixNano())
//		randomMinutes := rand.Intn(5-1) + 1
//		randomSeconds := time.Duration(rand.Intn(60-1)+1) * time.Second
//
//		// keep track of total filled quantity of the TWAP
//		if filledInTotal.GreaterThanOrEqual(decimal.NewFromFloat(targetAmount)) {
//			logrus.Infof("TWAP finished: target amount %.8f filled: %s\n", targetAmount, filledInTotal.String())
//			break
//		}
//
//		// next execution is some random minute and random second
//		nextExecution := time.Duration(randomMinutes)*time.Minute + randomSeconds
//
//		// Calculate the quantity that has to be filled on average per second so it can reach the target before the TWAP end time
//		wapAmount = t.AverageSizeFillPerSecond(decimal.NewFromFloat(targetAmount)).Mul(decimal.NewFromFloat(nextExecution.Seconds()))
//		logrus.Infof("TWAP quantity %s for next execution: %s\n:", t.Pair.Base, wapAmount.String())
//
//		f := ftx.FTX{Base: *t.Exchange.GetBase()}
//
//		f.GetMarket("SOL")
//		tickerInfo, err := t.Exchange.FetchTicker(t.Pair, t.Asset)
//		if err != nil {
//			logrus.Errorf("fetch ticker failed: %s\n", err)
//		}
//
//		// estimate how much USD/USDT/BTC is used for next execution
//		logrus.Infof("Estimated amount of %s used for next execution: %s", t.Pair.Quote, wapAmount.Mul(decimal.NewFromFloat(tickerInfo.Last)).String())
//
//		//if wapAmount.LessThanOrEqual(decimal.NewFromFloat(baseMinSize)) {
//		//	for wapAmount.LessThanOrEqual(decimal.NewFromFloat(baseMinSize)) {
//		//		nextExecution = +nextExecution
//		//		wapAmount = wapAmount.Add(t.AverageSizeFillPerSecond(decimal.NewFromFloat(targetAmount).Sub(filledInTotal)).Mul(decimal.NewFromFloat(nextExecution.Seconds())))
//		//	}
//		//	logrus.Infof("Increased WAP amount to reach minimal size: %s\n", wapAmount.String())
//		//}
//
//		if filledInTotal.Add(wapAmount).GreaterThanOrEqual(decimal.NewFromFloat(targetAmount)) {
//			wapAmount = decimal.NewFromFloat(targetAmount).Sub(filledInTotal)
//		}
//
//		wapResponse, err := t.Submit(tickerInfo.Ask, wapAmount, mode)
//		if err != nil {
//			logrus.Errorf("wapresponse error: %s\n", err)
//			<-time.After(time.Minute * 5)
//			continue
//		}
//		filledInTotal = filledInTotal.Add(wapAmount)
//		progress := decimal.NewFromFloat(100.0).Div(decimal.NewFromFloat(targetAmount)).Mul(filledInTotal).Round(2)
//		logrus.Infof("Buying %s %s at price %.8f\t%s\t%.2f minutes left\t%s%%\n", wapAmount.String(), t.Pair.String(), wapResponse.Rate, time.Now().Format(time.RFC822), t.End.Sub(time.Now()).Minutes(), progress.String())
//
//		// next amount to buy is determined by averageSize required per minute to meet the target amount multiplied by the minutes waiting for the next order
//		logrus.Infof("Next order execution takes place after: %s\n", nextExecution.String())
//		<-time.After(nextExecution)
//	}
//}
//

// AverageSizeFillPerMinute returns the average size of the order that has to be filled per second to reach the target amount
func AverageSizeFillPerMinute(start time.Time, end time.Time, amount decimal.Decimal) decimal.Decimal {
	diff := end.Sub(start).Minutes()
	a := amount.Div(decimal.NewFromFloat(diff)) // average size fill per minute
	return a
}

//
//func (t TWAP) AverageSizeFillPerSecond(amount decimal.Decimal) decimal.Decimal {
//	diff := t.End.Sub(time.Now()).Seconds()
//	a := amount.Div(decimal.NewFromFloat(diff)) // average size fill per minute
//	return a
//}
//
//func (t TWAP) Submit(price float64, amount decimal.Decimal, mode string) (order.SubmitResponse, error) {
//	wapAmount, _ := amount.Float64()
//	wapAmount, _ = new(big.Float).SetPrec(2).SetFloat64(wapAmount).Float64()
//	logrus.Infof("fetching ticker order from")
//	ticker, err := t.Exchange.FetchTicker(t.Pair, t.Asset)
//	if err != nil {
//		logrus.Errorf("fetch ticker failed: %s\n", err)
//		return order.SubmitResponse{}, err
//	}
//
//	pctChange := (100.00 / price) * ticker.Last
//
//	if t.MaxChangePct > pctChange {
//		return order.SubmitResponse{}, fmt.Errorf("change in price percentage exceeds limit risk %.2f\n", pctChange)
//	}
//
//	if t.Side == order.Buy && t.OverBought == true {
//		return order.SubmitResponse{}, fmt.Errorf("failed to buy %s because the price is above the TWAP\n", t.Pair.String())
//	}
//
//	if t.Side == order.Sell && t.OverBought != true {
//		return order.SubmitResponse{}, fmt.Errorf("failed to buy %s because the price is above the TWAP\n", t.Pair.String())
//	}
//
//	if t.Side == order.Buy {
//		price = ticker.Ask
//	}
//
//	if t.Side == order.Sell {
//		price = ticker.Bid
//	}
//
//	var orderType order.Type
//	if mode == "passive" {
//		orderType = order.Limit
//	}
//
//	if mode == "aggressive" {
//		orderType = order.Market
//	}
//
//	wapOrder := order.Submit{
//		Price:           price,
//		Amount:          wapAmount,
//		Exchange:        t.Exchange.GetName(),
//		InternalOrderID: "",
//		Type:            orderType,
//		Side:            t.Side,
//		AssetType:       t.Asset,
//		Pair:            t.Pair,
//	}
//
//	wapResponse, err := t.Exchange.SubmitOrder(&wapOrder) // TODO LOT_SIZE MinQty
//	if err != nil {
//		logrus.Errorf("submit order failed: %s\n", err)
//		return order.SubmitResponse{}, errors.Newf("submit oder failed: %s", err)
//	}
//
//	wapResponse.Rate = price
//	return wapResponse, err
//}
//
////func (w WAPService) Run (t TWAP){
////	w.Queue <- t
////}
//
////// testing purpose
////func (w WAPService) Run(t TWAP, message chan string) {
////	fmt.Println("start")
////	time.Sleep(time.Second * 3)
////	message <- "executing...."
////	time.Sleep(time.Second * 3)
////	message <- "Almost...."
////	time.Sleep(time.Second * 2)
////	message <- "finished wap"
////	//go twap.Execute()
////}
