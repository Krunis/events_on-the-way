package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Krunis/events_on-the-way/packages/common"
	"github.com/Krunis/events_on-the-way/packages/polleroutbox"
	"github.com/Krunis/events_on-the-way/packages/saramakafka/producer"
)

func main() {
	brokerList := []string{"kafka:9092"}

	cfg := polleroutbox.ConfigPoll{Interval: time.Second * 5, BatchSize: 5}

	poller := polleroutbox.NewPollerOutboxService(&cfg)

	prod, err := producer.NewSaramaProducer(brokerList)
	if err != nil {
		log.Fatalf("Failed to create producer: %s", err)
	}

	log.Println("Producer created")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(stop)

	errCh := make(chan error, 1)

	go func() {
		if err := poller.Start(prod, common.GetDBConnectionString()); err != nil {
			errCh <- err
		}
	}()

	select {
	case <-stop:
		if err := poller.Stop(); err != nil{
			log.Printf("Error while stop: %s\n", err)
		}
	case err := <-errCh:
		log.Printf("Poller error: %s\n", err)

		if err := poller.Stop(); err != nil{
			log.Printf("Error while stop: %s\n", err)
		}
	}
}
