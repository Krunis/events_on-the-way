package consumer

import (
	"context"
	"log"

	"github.com/IBM/sarama"
)

type ConsumerService struct {
	consumerGroup sarama.ConsumerGroup

	topics []string

	handler func(msg *sarama.ConsumerMessage) error

	brokers []string

	groupID string

	ctx    context.Context
	cancel context.CancelFunc
}

func NewConsumerService(handler func(msg *sarama.ConsumerMessage) error, topics []string, brokers []string, groupID string) *ConsumerService {

	ctx, cancel := context.WithCancel(context.Background())

	consumerGroup, err := sarama.NewConsumerGroup(brokers, "A", GetConfig())
	if err != nil{
		log.Fatalf("Failed to create consumer group: %s\n", err)
	}

	return &ConsumerService{
		consumerGroup: consumerGroup,
		topics:  topics,
		handler: handler,
		brokers: brokers,
		groupID: groupID,
		ctx:     ctx,
		cancel:  cancel}
}

func (c *ConsumerService) Start() error {
	for {
		select {
		case <-c.ctx.Done():
			return c.ctx.Err()
		default:
			if err := c.consumerGroup.Consume(c.ctx, c.topics, c); err != nil {
				return err
			}
		}
	}
}

func (c *ConsumerService) Stop() error{
	c.cancel()

	if err := c.consumerGroup.Close(); err != nil{
		return err
	}
	
	log.Println("Consumer service stopped")
	return nil
}