package twap

import (
	"context"
	"encoding/json"
	"github.com/hibiken/asynq"
)

const (
	TypeTwap  = "twap"
	TypeOrder = "order"
)

// NewTwapTask represents a task for TWAP algorithm.
func NewTwapTask(twap Payload) (*asynq.Task, error) {
	payload, err := json.Marshal(twap)
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeTwap, payload), nil
}

// HandleTwapTask handles a task for TWAP algorithm.
func HandleTwapTask(ctx context.Context, t *asynq.Task) error {
	var p Payload
	err := json.Unmarshal(t.Payload(), &p)
	if err != nil {
		return err
	}

	go Execute(p)

	return nil
}
