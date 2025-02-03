package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"goodblast/internal/application/service"
	"goodblast/internal/domain/events"
	"goodblast/pkg/log"
)

type ProgressUpdateConsumer struct {
	tournamentService service.ITournamentService
	kafkaConfig       *kafka.ConfigMap
}

func NewProgressUpdateConsumer(tournamentService service.ITournamentService, kafkaConfig *kafka.ConfigMap) *ProgressUpdateConsumer {
	return &ProgressUpdateConsumer{
		tournamentService: tournamentService,
		kafkaConfig:       kafkaConfig,
	}
}

func (puc *ProgressUpdateConsumer) StartProgressUpdateConsumer(topic string) {
	consumer, err := kafka.NewConsumer(puc.kafkaConfig)
	if err != nil {
		log.GetLogger().Error(fmt.Sprintf("Failed to create Kafka consumer: %v", err))
		return
	}

	err = consumer.SubscribeTopics([]string{topic}, nil)
	if err != nil {
		log.GetLogger().Error(fmt.Sprintf("Failed to subscribe to topic %s: %v", topic, err))
		return
	}

	for {
		msg, err := consumer.ReadMessage(-1)
		if err == nil {
			var updateMessage events.ProgressUpdateMessage
			err := json.Unmarshal(msg.Value, &updateMessage)
			if err != nil {
				log.GetLogger().Error(fmt.Sprintf("Failed to unmarshal progress update message: %v", err))
				continue
			}

			log.GetLogger().Info(fmt.Sprintf("Received progress update for user %d", updateMessage.UserID))

			err = puc.tournamentService.UpdateTournamentScore(context.Background(), updateMessage)
			if err != nil {
				log.GetLogger().Error(fmt.Sprintf("Failed to update tournament score for user %d: %v", updateMessage.UserID, err))
			}
		} else {
			log.GetLogger().Error(fmt.Sprintf("Consumer error: %v", err))
		}
	}
}
