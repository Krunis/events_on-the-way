package producer

import (
	"context"

	"github.com/IBM/sarama"
)

type SaramaProducer struct {
	syncProducer sarama.SyncProducer
	config       *sarama.Config
}

func NewProducerService(brokerList []string) (*SaramaProducer, error) {
	config := sarama.NewConfig()

	producer, err := sarama.NewSyncProducer(brokerList, config)
	if err != nil {
		return nil, err
	}

	return &SaramaProducer{
		syncProducer: producer,
		config:       config,}, nil
}

func (s *SaramaProducer) SendBatch(ctx context.Context, events []*KafkaEvent) error{
	batch := make([]*sarama.ProducerMessage, len(events))

	for i, event := range events{
		batch[i] = &sarama.ProducerMessage{
			Topic: "events_trip",
			Key: sarama.ByteEncoder(event.Key()),
			Value: sarama.ByteEncoder(event.Value()),
			Timestamp: event.Timestamp,
		}
	
	}

	if err := s.syncProducer.SendMessages(batch); err != nil{
		return err
	}
	
	return nil
}

func (s *SaramaProducer) Close() error{

}
