package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	ckafka "github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"goodblast/internal/application/service"
	"goodblast/internal/domain/events"
	"goodblast/pkg/log"
)

type TournamentEntryConsumer struct {
	consumer          *ckafka.Consumer
	topic             string
	tournamentService service.ITournamentService
}

func NewTournamentEntryConsumer(
	consumer *ckafka.Consumer,
	topic string,
	tournamentService service.ITournamentService,
) *TournamentEntryConsumer {
	return &TournamentEntryConsumer{
		consumer:          consumer,
		topic:             topic,
		tournamentService: tournamentService,
	}
}

func (tc *TournamentEntryConsumer) StartConsume(ctx context.Context) error {
	err := tc.consumer.SubscribeTopics([]string{tc.topic}, nil)
	if err != nil {
		return fmt.Errorf("failed to subscribe to topic %s: %w", tc.topic, err)
	}

	go func() {
		for {
			msg, err := tc.consumer.ReadMessage(-1)
			if err != nil {
				log.GetLogger().Errorf(fmt.Sprintf("Kafka consumer read error: %v", err))
				continue
			}
			tc.handleMessage(ctx, msg)
		}
	}()
	return nil
}

func (tc *TournamentEntryConsumer) handleMessage(ctx context.Context, msg *ckafka.Message) {
	var payload events.EnterTournamentPayload
	if err := json.Unmarshal(msg.Value, &payload); err != nil {
		log.GetLogger().Errorf(fmt.Sprintf("Failed to unmarshal tournament entry payload: %v", err))
		return
	}

	log.GetLogger().Infof(fmt.Sprintf("Received a tournament entry message for userID=%d from topic=%s",
		payload.UserID, *msg.TopicPartition.Topic))

	err := tc.tournamentService.EnterTournament(ctx, payload.UserID)
	if err != nil {
		log.GetLogger().Errorf(fmt.Sprintf("EnterTournamentTransaction failed for userID=%d, err=%v", payload.UserID, err))
		return
	}

	log.GetLogger().Infof(fmt.Sprintf("Successfully processed tournament entry for userID=%d", payload.UserID))
}
