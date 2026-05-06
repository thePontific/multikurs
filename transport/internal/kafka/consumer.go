package kafka

import (
	"encoding/json"
	"fmt"

	"transport-service/pkg/models"

	"github.com/IBM/sarama"
)

// StartConsumer запускает consumer для чтения сегментов из Kafka
func StartConsumer(addSegmentFunc func(models.Segment)) error {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true
	config.Consumer.Offsets.Initial = sarama.OffsetNewest

	consumer, err := sarama.NewConsumer([]string{KafkaAddr}, config)
	if err != nil {
		return fmt.Errorf("error creating consumer: %w", err)
	}
	defer consumer.Close()

	partitionConsumer, err := consumer.ConsumePartition(KafkaTopic, 0, sarama.OffsetNewest)
	if err != nil {
		return fmt.Errorf("error opening topic: %w", err)
	}
	defer partitionConsumer.Close()

	fmt.Println("🚀 Consumer Kafka запущен, ожидание сегментов...")

	for {
		select {
		case msg := <-partitionConsumer.Messages():
			var segment models.Segment
			if err := json.Unmarshal(msg.Value, &segment); err != nil {
				fmt.Printf("Ошибка парсинга сегмента: %v\n", err)
				continue
			}
			// добавляем сегмент в хранилище для последующей сборки
			addSegmentFunc(segment)

		case err := <-partitionConsumer.Errors():
			fmt.Printf("Ошибка consumer: %s\n", err.Error())
		}
	}
}
