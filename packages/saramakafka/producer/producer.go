package producer

import (
	"context"
	"log"
	"time"

	"github.com/IBM/sarama"
)

type SaramaProducer struct {
	syncProducer sarama.SyncProducer
	config       *sarama.Config
}

func NewSaramaProducer(brokerList []string) (*SaramaProducer, error) {
	config := sarama.NewConfig()

	config.Producer.RequiredAcks = sarama.WaitForAll
    config.Producer.Idempotent = true
    config.Net.MaxOpenRequests = 1

	config.Producer.Retry.Max = 10
    config.Producer.Retry.Backoff = 100 * time.Millisecond

	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true

	config.Producer.Partitioner = sarama.NewHashPartitioner

	config.Producer.Timeout = 30 * time.Second
    config.Net.DialTimeout = 30 * time.Second
    config.Net.ReadTimeout = 30 * time.Second
    config.Net.WriteTimeout = 30 * time.Second


	producer, err := sarama.NewSyncProducer(brokerList, config)
	if err != nil {
		return nil, err
	}

	return &SaramaProducer{
		syncProducer: producer,
		config:       config,}, nil
}

func (p *SaramaProducer) SendBatch(ctx context.Context, events []*KafkaEvent) error{
	batch := make([]*sarama.ProducerMessage, len(events))

	for i, event := range events{
		batch[i] = &sarama.ProducerMessage{
			Topic: "events_trip",
			Key: sarama.ByteEncoder(event.Key()),
			Value: sarama.ByteEncoder(event.Value()),
			Timestamp: event.Timestamp,
		}
	
	}

	if err := p.syncProducer.SendMessages(batch); err != nil{
		return err
	}
	
	log.Println("Sent")

	return nil
}

func (s *SaramaProducer) Close() error{
	if err := s.syncProducer.Close(); err != nil{
		return err
	}

	return nil
}
