package producer

import (
	"fmt"
	ckafka "github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"sync"
)

type Producer struct {
	producer *ckafka.Producer
}

var (
	instance *Producer
	once     sync.Once
)

func NewKafkaProducer(configMap *ckafka.ConfigMap) (*Producer, error) {
	var err error
	once.Do(func() {
		p, e := ckafka.NewProducer(configMap)
		if e != nil {
			err = fmt.Errorf("failed to create kafka producer: %w", e)
			return
		}
		instance = &Producer{producer: p}
	})
	if err != nil {
		return nil, err
	}
	return instance, nil
}

func (k *Producer) GetProducer() *ckafka.Producer {
	return k.producer
}

func ProduceFireAndForget(producer *ckafka.Producer, topic string, data []byte) error {
	err := producer.Produce(&ckafka.Message{
		TopicPartition: ckafka.TopicPartition{
			Topic:     &topic,
			Partition: ckafka.PartitionAny,
		},
		Value: data,
	}, nil)
	if err != nil {
		return fmt.Errorf("failed to enqueue message: %w", err)
	}

	return nil
}
