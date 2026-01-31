package producer

import (
	"context"

	"github.com/IBM/sarama"
)

type SaramaProducer struct {
	syncProducer sarama.SyncProducer
	config       *sarama.Config
}

func NewSaramaProducer(brokerList []string) (*SaramaProducer, error) {
	config := sarama.NewConfig()

	config.Producer.Return.Successes = true

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
	
	return nil
}

func (s *SaramaProducer) Close() error{
	if err := s.Close(); err != nil{
		return err
	}

	return nil
}
