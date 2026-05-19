package kafka

import (
	"encoding/json"
	"fmt"

	"transport-service/pkg/models"

	"github.com/IBM/sarama"
)

const (
	//KafkaAddr  = "localhost:9092"
	KafkaAddr  = "kafka:9092" // имя сервиса в Docker
	KafkaTopic = "image_segments"
)

// SendToKafka отправляет сегмент в Kafka (Producer)
func SendToKafka(segment models.Segment) error {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Version = sarama.V2_0_0_0

	producer, err := sarama.NewSyncProducer([]string{KafkaAddr}, config)
	if err != nil {
		return fmt.Errorf("error creating producer: %w", err)
	}
	defer producer.Close()

	segmentJSON, err := json.Marshal(segment)
	if err != nil {
		return fmt.Errorf("error marshaling segment: %w", err)
	}

	msg := &sarama.ProducerMessage{
		Topic: KafkaTopic,
		Value: sarama.StringEncoder(segmentJSON),
		// Key: используем request_id для партиционирования
		Key: sarama.StringEncoder(segment.RequestID),
	}

	_, _, err = producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}
