package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/redis/go-redis/v9"
	"goodblast/internal/domain/events"
	"goodblast/pkg/log"
	"strconv"
)

type LeaderboardConsumer struct {
	redisClient *redis.Client
	kafkaConfig *kafka.ConfigMap
}

func NewLeaderboardConsumer(redisClient *redis.Client, kafkaConfig *kafka.ConfigMap) *LeaderboardConsumer {
	return &LeaderboardConsumer{
		redisClient: redisClient,
		kafkaConfig: kafkaConfig,
	}
}

func (lc *LeaderboardConsumer) StartLeaderboardUpdateConsumer(topic string) {
	consumer, err := kafka.NewConsumer(lc.kafkaConfig)
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
			var updateMessage events.LeaderboardUpdateMessage
			err := json.Unmarshal(msg.Value, &updateMessage)
			if err != nil {
				log.GetLogger().Error(fmt.Sprintf("Failed to unmarshal leaderboard update message: %v", err))
				continue
			}

			ctx := context.Background()

			// **Global LeaderBoard**
			_, err = lc.redisClient.ZAdd(ctx, "leaderboard:global", []redis.Z{{
				Score:  float64(updateMessage.Score),
				Member: strconv.FormatInt(updateMessage.UserID, 10),
			}}...).Result()
			if err != nil {
				log.GetLogger().Error(fmt.Sprintf("Failed to update global leaderboard for user %d: %v", updateMessage.UserID, err))
			}

			// **Country LeaderBoard**
			_, err = lc.redisClient.ZAdd(ctx, fmt.Sprintf("leaderboard:%s", updateMessage.Country), []redis.Z{{
				Score:  float64(updateMessage.Score),
				Member: strconv.FormatInt(updateMessage.UserID, 10),
			}}...).Result()
			if err != nil {
				log.GetLogger().Error(fmt.Sprintf("Failed to update country leaderboard for user %d in %s: %v", updateMessage.UserID, updateMessage.Country, err))
			}

			// **Tournament LeaderBoard**
			_, err = lc.redisClient.ZAdd(ctx, fmt.Sprintf("leaderboard:tournament:%d", updateMessage.TournamentID), []redis.Z{{
				Score:  float64(updateMessage.Score),
				Member: strconv.FormatInt(updateMessage.UserID, 10),
			}}...).Result()
			if err != nil {
				log.GetLogger().Error(fmt.Sprintf("Failed to update tournament leaderboard for user %d in tournament %d: %v", updateMessage.UserID, updateMessage.TournamentID, err))
			}

			log.GetLogger().Info(fmt.Sprintf("Updated leaderboards for user %d - Score: %d, Country: %s, Tournament: %d",
				updateMessage.UserID, updateMessage.Score, updateMessage.Country, updateMessage.TournamentID))
		} else {
			log.GetLogger().Error(fmt.Sprintf("Consumer error: %v", err))
		}
	}
}
