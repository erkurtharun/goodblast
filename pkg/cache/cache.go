package cache

import (
	"encoding/json"
	"github.com/patrickmn/go-cache"
	"time"
)

var leaderboardCache *cache.Cache

func InitCache() {
	leaderboardCache = cache.New(30*time.Second, 60*time.Second)
}

func SetCache(key string, data interface{}, expiration time.Duration) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return
	}
	leaderboardCache.Set(key, jsonData, expiration)
}

func GetCache(key string, target interface{}) (bool, error) {
	cachedData, found := leaderboardCache.Get(key)
	if !found {
		return false, nil
	}

	jsonData, ok := cachedData.([]byte)
	if !ok {
		return false, nil
	}

	err := json.Unmarshal(jsonData, target)
	if err != nil {
		return false, err
	}

	return true, nil
}
