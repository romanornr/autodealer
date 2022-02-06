package twap

import (
	"github.com/hibiken/asynq"
	"github.com/romanornr/autodealer/internal/orderbuilder"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"time"
)

// Execute executes the TWAP algorithm
// We can improve the algorithm by using a queue of orders and processing them in a separate goroutine TODO
func Execute(t Payload) {
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

	qty := minimalSize.InexactFloat64()

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

		ob := orderbuilder.NewOrderBuilder()
		ob.
			AtExchange(t.Exchange).
			ForAccountID(t.AccountID).
			ForCurrencyPair(t.Pair).
			WithAssetType(t.Asset).
			ForPrice(11).
			WithAmount(qty).
			UseOrderType(t.OrderType).
			SetSide(t.Side)
		//
		newOrder, err := ob.Build()
		if err != nil {
			logrus.Errorf("failed to build order: %v", err)
			return
		}

		t1, err := NewOrderTask(newOrder)
		if err != nil {
			logrus.Errorf("Error creating order task: %v", err)
		}

		info, err := client.Enqueue(t1, asynq.ProcessAt(nextExecutionTime))
		if err != nil {
			logrus.Errorf("Error enqueuing order task: %v", err)
		}
		logrus.Printf("Order task enqueued: %s\n", info.ID)
	}

}

// AverageSizeFillPerMinute returns the average size of the order that has to be filled per second to reach the target amount
func AverageSizeFillPerMinute(start time.Time, end time.Time, amount decimal.Decimal) decimal.Decimal {
	diff := end.Sub(start).Minutes()
	a := amount.Div(decimal.NewFromFloat(diff)) // average size fill per minute
	return a
}
