package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/IBM/sarama"
	"github.com/Krunis/events_on-the-way/packages/saramakafka/consumer"
)

func main() {
	stopCh := make(chan os.Signal, 1)
	signal.Notify(stopCh, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(stopCh)

	errCh := make(chan error, 1)

	handler := func(msg *sarama.ConsumerMessage) error{
		log.Printf("Received message created at %v from offset %d. Value: %v", msg.Timestamp, msg.Offset, string(msg.Value))
		return nil
	}

	brokers := []string{"kafka:9092"}

	topics := []string{"events_trip"}

	groupID := "A"

	consService := consumer.NewConsumerService(handler, topics, brokers, groupID)

	go func(){
		if err := consService.Start(); err != nil{
			errCh <- err
		}
	}()

	select {
	case <-stopCh:
		if err := consService.Stop(); err != nil{
			log.Printf("Failed to stop consumer: %s\n", err)
		}
	case err := <-errCh:
		log.Printf("Error while consuming: %s\n", err)
		consService.Stop()
	}
}