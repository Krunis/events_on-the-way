package polleroutbox

import (
	"context"
	"fmt"
	"time"

	"github.com/Krunis/events_on-the-way/packages/common"
	"github.com/Krunis/events_on-the-way/packages/saramakafka/producer"
	"github.com/jackc/pgx/v4/pgxpool"
)

type ConfigPoll struct {
	interval  time.Duration
	batchSize uint8
}

type PollerOutboxService struct {
	dbPool             *pgxpool.Pool
	dbConnectionString string

	cfg *ConfigPoll

	producer producer.Producer

	lifecycle struct {
		ctx    context.Context
		cancel context.CancelFunc
	}
}

func NewPollerOutboxService(dbConnectionString string) *PollerOutboxService {
	ctx, cancel := context.WithCancel(context.Background())

	return &PollerOutboxService{
		lifecycle: struct {
			ctx    context.Context
			cancel context.CancelFunc
		}{ctx: ctx, cancel: cancel},
	}
}

func (p *PollerOutboxService) StartPollingOutbox(producer producer.Producer) error {
	var err error

	p.dbPool, err = common.ConnectToDB(p.lifecycle.ctx, p.dbConnectionString)
	if err != nil {
		return err
	}

	p.producer = producer

	ticker := time.NewTicker(p.cfg.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(p.lifecycle.ctx, time.Second * 10)
			defer cancel()

			if err := p.processBatch(ctx); err != nil{
				return fmt.Errorf("failed to process: %w", err)
			}
		case <-p.lifecycle.ctx.Done():
			return p.lifecycle.ctx.Err()
		}
	}
}

func (p *PollerOutboxService) processBatch(ctx context.Context) error {
	rows, err := p.GetRowsFromOutbox()
	if err != nil {
		return fmt.Errorf("failed to select from outbox: %w", err)
	}

	if len(rows) == 0{
		return nil
	}

	events := make([]*producer.KafkaEvent, len(rows))

	for i, row := range rows{
		events[i] = producer.NewKafkaEvent(row)
	}

	if err := p.producer.SendBatch(ctx, events); err != nil{
		return err
	}


}
