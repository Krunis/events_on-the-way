package polleroutbox

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/Krunis/events_on-the-way/packages/common"
	"github.com/Krunis/events_on-the-way/packages/saramakafka/producer"
	"github.com/jackc/pgx/v4/pgxpool"
)

type ConfigPoll struct {
	Interval  time.Duration
	BatchSize uint8
}

type PollerOutboxService struct {
	dbPool             *pgxpool.Pool

	cfg *ConfigPoll

	producer producer.Producer

	lifecycle struct {
		ctx    context.Context
		cancel context.CancelFunc
	}

	stopOnce sync.Once
}

func NewPollerOutboxService(cfg *ConfigPoll) *PollerOutboxService {
	ctx, cancel := context.WithCancel(context.Background())

	return &PollerOutboxService{
		cfg: cfg,

		lifecycle: struct {
			ctx    context.Context
			cancel context.CancelFunc
		}{ctx: ctx, cancel: cancel},
	}
}

func (p *PollerOutboxService) Start(producer producer.Producer, dbConnectionString string) error {
	var err error

	p.dbPool, err = common.ConnectToDB(p.lifecycle.ctx, dbConnectionString)
	if err != nil {
		return err
	}

	p.producer = producer

	ticker := time.NewTicker(p.cfg.Interval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(p.lifecycle.ctx, time.Second * 10)
			defer cancel()

			if err := p.processBatch(ctx); err != nil {
				log.Printf("Failed to process batch: %s\n", err)
			}
		case <-p.lifecycle.ctx.Done():
			return p.lifecycle.ctx.Err()
		}
	}
}

func (p *PollerOutboxService) processBatch(ctx context.Context) error {
	tx, err := p.dbPool.Begin(ctx)
	if err != nil{
		return err
	}
	defer tx.Rollback(ctx)

	rows, err := p.GetRowsFromOutbox(ctx, tx)
	if err != nil {
		return fmt.Errorf("failed to select from outbox: %w", err)
	}

	if len(rows) == 0 {
		return nil
	}

	events := make([]*producer.KafkaEvent, len(rows))

	for i, row := range rows {
		events[i] = producer.NewKafkaEvent(row)
	}

	if err := p.producer.SendBatch(ctx, events); err != nil {
		return err
	}

	if err := p.MarkAsSentOutbox(ctx, tx, rows); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil{
		return err
	}
	
	log.Println("Commited")

	return nil
}

func (p *PollerOutboxService) Stop() error {
	var err error

	p.stopOnce.Do(func() {
		gracefuleShutdown := time.Second * 1

		shutdownStart := time.Now()

		p.lifecycle.cancel()

		if p.producer != nil {
			log.Println("Stopping producer..")

			if err = p.producer.Close(); err != nil {
				log.Printf("Failed to close producer: %s", err)
			}

			log.Println("Producer stopped")

		}

		if p.dbPool != nil {
			p.dbPool.Close()
			log.Println("Database pool stopped")
		}

		shutdownDuration := time.Since(shutdownStart)

		if gracefuleShutdown < shutdownDuration {
			log.Printf("time since shutdown: %v", shutdownDuration)
		}

		log.Println("Poller stopped")
	})

	return err
}
