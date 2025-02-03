package appconfig

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"sync"
	"time"
)

type ConfigKey string

type DynamicConfig struct {
	TournamentCutoffHour        int    `json:"tournamentCutoffHour"`
	MinimumTournamentEntryLevel int    `json:"minimumTournamentEntryLevel"`
	TournamentEntranceCoins     int    `json:"tournamentEntranceCoins"`
	Reward1                     int    `json:"reward1"`
	Reward2                     int    `json:"reward2"`
	Reward3                     int    `json:"reward3"`
	Reward4to10                 int    `json:"reward4to10"`
	CoinPerLevel                int    `json:"coinPerLevel"`
	TokenTTL                    int    `json:"tokenTTL"`
	TournamentEntryTopic        string `json:"tournamentEntryTopic"`
	UserProgressUpdateTopic     string `json:"userProgressUpdateTopic"`
	LeaderboardUpdateTopic      string `json:"leaderboardUpdateTopic"`
}

type IDynamicConfigService interface {
	Initialize() error
	GetConfig() DynamicConfig
	WebhookHandler(ctx *gin.Context)
}

type DynamicConfigService struct {
	config           DynamicConfig
	dynamicConfigURL string
	githubToken      string
	mutex            sync.RWMutex
}

var (
	dynamicConfigInstance *DynamicConfigService
	dynamicConfigOnce     sync.Once
)

func GetDynamicConfigService(config *Config) *DynamicConfigService {
	dynamicConfigOnce.Do(func() {
		dynamicConfigInstance = &DynamicConfigService{
			dynamicConfigURL: config.ToggleConfigURL,
			githubToken:      config.GithubToken,
		}
		err := dynamicConfigInstance.Initialize()
		if err != nil {
			logrus.Errorf("Failed to initialize DynamicConfigService: %v", err)
		}
	})
	return dynamicConfigInstance
}

func (f *DynamicConfigService) Initialize() error {
	return f.loadConfigFromGitHub()
}

func (f *DynamicConfigService) loadConfigFromGitHub() error {
	client := http.Client{
		Timeout: 10 * time.Second,
	}

	urlWithTimestamp := fmt.Sprintf("%s?t=%d", f.dynamicConfigURL, time.Now().UnixNano())

	req, err := http.NewRequest("GET", urlWithTimestamp, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+f.githubToken)
	req.Header.Set("Accept", "application/vnd.github.v3+json") // GitHub API v3 formatı
	req.Header.Set("Cache-Control", "no-store, no-cache, must-revalidate")
	req.Header.Set("Pragma", "no-cache")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("could not fetch config from GitHub: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch config from GitHub, status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read GitHub config response: %v", err)
	}

	type GitHubContentResponse struct {
		Content string `json:"content"`
	}

	var githubResponse GitHubContentResponse
	if err := json.Unmarshal(body, &githubResponse); err != nil {
		return fmt.Errorf("failed to parse GitHub API response: %v", err)
	}

	decodedContent, err := base64.StdEncoding.DecodeString(githubResponse.Content)
	if err != nil {
		return fmt.Errorf("failed to decode Base64 content: %v", err)
	}

	var config DynamicConfig
	if err := json.Unmarshal(decodedContent, &config); err != nil {
		return fmt.Errorf("failed to parse JSON config from GitHub: %v", err)
	}

	f.mutex.Lock()
	defer f.mutex.Unlock()
	f.config = config

	logrus.Infof("Successfully updated dynamic config from GitHub: %+v", f.config)

	return nil
}

func (f *DynamicConfigService) GetConfig() DynamicConfig {
	f.mutex.RLock()
	defer f.mutex.RUnlock()
	return f.config
}

func (f *DynamicConfigService) WebhookHandler(ctx *gin.Context) {
	logrus.Info("Received webhook request to update dynamic config")
	if err := f.loadConfigFromGitHub(); err != nil {
		logrus.Errorf("Failed to reload config from GitHub: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reload config"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Configuration updated successfully"})
}
