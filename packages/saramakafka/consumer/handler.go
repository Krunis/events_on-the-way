package consumer

import (
	"log"

	"github.com/IBM/sarama"
)

func (c *ConsumerService) Setup(session sarama.ConsumerGroupSession) error {
	log.Println("Setup")
	return nil
}

func (c *ConsumerService) Cleanup(session sarama.ConsumerGroupSession) error {
	log.Println("Cleanup")
	return nil
}

func (c *ConsumerService) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		
			select{
			case msg, ok := <-claim.Messages():
				if !ok{
					return nil
				}

				if err := c.handler(msg); err != nil{
					log.Printf("Failed to proceed message: %v\n", msg)
					continue
				}

				
				
			
			case <-session.Context().Done():
				return session.Context().Err()
		
	}
}}


