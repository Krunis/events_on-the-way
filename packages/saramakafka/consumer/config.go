package consumer

import "github.com/IBM/sarama"

func GetConfig() *sarama.Config{
	config := sarama.NewConfig()

	config.Consumer.Offsets.Initial = sarama.OffsetNewest
	
	return config
}