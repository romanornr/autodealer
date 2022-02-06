package twap

import (
	"context"
	"encoding/json"
	"github.com/hibiken/asynq"
	"github.com/sirupsen/logrus"
	"github.com/thrasher-corp/gocryptotrader/exchanges/order"
)

func NewOrderTask(order *order.Submit) (*asynq.Task, error) {
	payload, err := json.Marshal(order)
	if err != nil {
		return nil, err
	}

	logrus.Printf("sending order %s\n", order.ID)

	return asynq.NewTask(TypeOrder, payload), nil
}

func HandleOrderTask(ctx context.Context, task *asynq.Task) error {
	var o *order.Submit // order.Submit
	err := json.Unmarshal(task.Payload(), &o)
	if err != nil {
		return err
	}

	// example Submitting market buy order for BNB-USD amount 5.000000 at 11.000000 FTX
	logrus.Printf("Submitting %s %s order for %s amount %f at %f %s\n", o.Type.Lower(), o.Side.Lower(), o.Pair.String(), o.Amount, o.Price, o.Exchange)
	return nil
}
