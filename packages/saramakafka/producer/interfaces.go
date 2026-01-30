package producer

import (
	"context"
)

type Producer interface {
	SendBatch(context.Context, []*KafkaEvent) error
	Close() error
}



