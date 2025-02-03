package service

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"goodblast/internal/application/controller/response"
	"goodblast/pkg/cache"
	"strconv"
	"time"
)

type ILeaderboardService interface {
	UpdateUserScore(userID int64, score int64, country string) error
	GetGlobalLeaderboard(limit int64) ([]response.LeaderboardEntry, error)
	GetCountryLeaderboard(country string, limit int64) ([]response.LeaderboardEntry, error)
	GetUserRank(userID int64) (int64, error)
}

type LeaderboardService struct {
	redisClient *redis.Client
}

func NewLeaderboardService(redisClient *redis.Client) ILeaderboardService {
	cache.InitCache()
	return &LeaderboardService{
		redisClient: redisClient,
	}
}

func (s *LeaderboardService) UpdateUserScore(userID int64, score int64, country string) error {
	ctx := context.Background()
	_, err := s.redisClient.ZAdd(ctx, "leaderboard:global", redis.Z{
		Score:  float64(score),
		Member: userID,
	}).Result()
	if err != nil {
		return err
	}

	_, err = s.redisClient.ZAdd(ctx, "leaderboard:"+country, redis.Z{
		Score:  float64(score),
		Member: userID,
	}).Result()
	if err != nil {
		return err
	}

	cacheKeyGlobal := "leaderboard:global"
	cacheKeyCountry := fmt.Sprintf("leaderboard:country:%s", country)
	cache.SetCache(cacheKeyGlobal, nil, 1*time.Millisecond)
	cache.SetCache(cacheKeyCountry, nil, 1*time.Millisecond)

	return nil
}

func (s *LeaderboardService) GetGlobalLeaderboard(limit int64) ([]response.LeaderboardEntry, error) {
	cacheKey := "leaderboard:global"
	var leaderboard []response.LeaderboardEntry

	cached, err := cache.GetCache(cacheKey, &leaderboard)
	if cached {
		return leaderboard, nil
	}

	ctx := context.Background()
	users, err := s.redisClient.ZRevRangeWithScores(ctx, "leaderboard:global", 0, limit-1).Result()
	if err != nil {
		return nil, err
	}

	for rank, user := range users {
		userID, _ := strconv.ParseInt(user.Member.(string), 10, 64)
		leaderboard = append(leaderboard, response.LeaderboardEntry{
			UserID: userID,
			Score:  int64(user.Score),
			Rank:   int64(rank + 1),
		})
	}

	cache.SetCache(cacheKey, leaderboard, 30*time.Second)

	return leaderboard, nil
}

func (s *LeaderboardService) GetCountryLeaderboard(country string, limit int64) ([]response.LeaderboardEntry, error) {
	cacheKey := fmt.Sprintf("leaderboard:country:%s", country)
	var leaderboard []response.LeaderboardEntry

	cached, err := cache.GetCache(cacheKey, &leaderboard)
	if cached {
		return leaderboard, nil
	}

	ctx := context.Background()
	users, err := s.redisClient.ZRevRangeWithScores(ctx, "leaderboard:"+country, 0, limit-1).Result()
	if err != nil {
		return nil, err
	}

	for rank, user := range users {
		userID, _ := strconv.ParseInt(user.Member.(string), 10, 64)
		leaderboard = append(leaderboard, response.LeaderboardEntry{
			UserID: userID,
			Score:  int64(user.Score),
			Rank:   int64(rank + 1),
		})
	}

	cache.SetCache(cacheKey, leaderboard, 30*time.Second)

	return leaderboard, nil
}

func (s *LeaderboardService) GetUserRank(userID int64) (int64, error) {
	ctx := context.Background()
	rank, err := s.redisClient.ZRevRank(ctx, "leaderboard:global", fmt.Sprintf("%d", userID)).Result()
	if err != nil {
		return -1, err
	}
	return rank + 1, nil
}
