package cmd

import (
	"context"
	"fmt"
	ckafka "github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/uptrace/bun"
	appconfig "goodblast/config"
	"goodblast/docs"
	"goodblast/internal/application/controller"
	"goodblast/internal/application/repository"
	"goodblast/internal/application/service"
	"goodblast/internal/infrastructure/kafka/consumer"
	"goodblast/internal/infrastructure/kafka/producer"
	"goodblast/internal/infrastructure/postgres"
	"goodblast/internal/infrastructure/redisclient"
	"goodblast/internal/middleware"
	"goodblast/internal/validation"
	"goodblast/pkg/auth"
	"goodblast/pkg/log"
	"goodblast/pkg/server"
	"net/http"
	"time"
)

func Execute() {
	config, err := appconfig.LoadConfig()
	if err != nil {
		logrus.Error(fmt.Sprintf("Failed to load configuration: %v", err))
		return
	}

	log.InitLogger(config)

	setupSwaggerInfo(config)

	engine := setupGinEngine()

	if config.Env != "prod" {
		setupSwagger(engine, &config)
	}

	database := setupDatabase(&config)
	redisCl := setupRedis(&config)

	middleware.LivenessHealthCheckMiddleware(engine)
	middleware.ReadinessHealthCheckMiddleware(engine, database)

	dynamicConfigService := appconfig.GetDynamicConfigService(&config)
	engine.POST("/webhook", dynamicConfigService.WebhookHandler)

	auth.InitAuth([]byte(config.TokenSecretKey), time.Duration(dynamicConfigService.GetConfig().TokenTTL)*time.Hour)

	validator := validation.NewRequestValidator()

	kafkaConfig := &ckafka.ConfigMap{
		"bootstrap.servers": config.KafkaBootstrapServers,
		"security.protocol": config.KafkaSecurityProtocol,
		"sasl.mechanisms":   config.KafkaSaslMechanism,
		"sasl.username":     config.KafkaSaslUsername,
		"sasl.password":     config.KafkaSaslPassword,
		"client.id":         config.KafkaClientId,
	}

	producerSingleton, err := producer.NewKafkaProducer(kafkaConfig)
	if err != nil {
		panic(err)
	}

	producerInstance := producerSingleton.GetProducer()

	// Repositories
	userRepository := repository.NewUserRepository(database)
	tournamentRepository := repository.NewTournamentRepository(database)
	groupRepository := repository.NewGroupRepository(database)
	tournamentUserRepository := repository.NewTournamentUserRepository(database)
	tournamentRewardRepository := repository.NewTournamentRewardRepository(database)

	// Clients

	// Services
	userService := service.NewUserService(
		userRepository, tournamentRepository, tournamentUserRepository,
		dynamicConfigService, producerInstance)
	tournamentService := service.NewTournamentService(database,
		tournamentRepository, groupRepository, tournamentUserRepository,
		userRepository, tournamentRewardRepository, dynamicConfigService,
		producerInstance)
	leaderBoardService := service.NewLeaderboardService(redisCl)

	// Controllers
	userController := controller.NewUserController(userService, validator)
	tournamentController := controller.NewTournamentController(tournamentService)
	leaderBoardController := controller.NewLeaderboardController(leaderBoardService)

	// Cache

	// Endpoints
	internal := engine.Group("/internal")
	internal.POST("/user", userController.CreateUser)
	internal.POST("/user/login", userController.Login)

	internal.Use(middleware.AuthMiddleware())
	internal.POST("/user/progress", userController.UpdateProgress)

	internalTournament := engine.Group("/internal/tournament")
	internalTournament.POST("/create-daily", tournamentController.CreateDailyTournament)
	internalTournament.POST("/start", tournamentController.StartTournament)
	internalTournament.POST("/close", tournamentController.CloseTournament)
	internalTournament.GET("/active", tournamentController.GetActiveTournament)
	internalTournament.Use(middleware.AuthMiddleware())
	internalTournament.POST("/enter", tournamentController.EnterTournament)
	internalTournament.POST("/reward/claim", tournamentController.ClaimReward)

	internalLeaderboard := engine.Group("/internal/leaderboard")
	internalLeaderboard.GET("/global", leaderBoardController.GetGlobalLeaderboard)
	internalLeaderboard.GET("/country/:country", leaderBoardController.GetCountryLeaderboard)
	internalLeaderboard.GET("/user/:userId", leaderBoardController.GetUserRank)

	setupCronJobs(tournamentService)

	// Kafka Consumer

	consumerConfig := &ckafka.ConfigMap{
		"bootstrap.servers":  config.KafkaBootstrapServers,
		"security.protocol":  config.KafkaSecurityProtocol,
		"sasl.mechanisms":    config.KafkaSaslMechanism,
		"sasl.username":      config.KafkaSaslUsername,
		"sasl.password":      config.KafkaSaslPassword,
		"client.id":          config.KafkaClientId,
		"session.timeout.ms": config.KafkaSessionTimeout,
		"group.id":           config.KafkaConsumerGroupId,
		"auto.offset.reset":  config.KafkaAutoOffsetReset,
	}

	kafkaConsumer, err := ckafka.NewConsumer(consumerConfig)
	if err != nil {
		panic(fmt.Errorf("failed to create kafka consumer: %w", err))
	}

	tournamentEntryConsumer := consumer.NewTournamentEntryConsumer(
		kafkaConsumer,
		dynamicConfigService.GetConfig().TournamentEntryTopic,
		tournamentService,
	)

	err = tournamentEntryConsumer.StartConsume(context.Background())
	if err != nil {
		log.GetLogger().Errorf(fmt.Sprintf("Failed to start Kafka consumer: %v", err))
	} else {
		log.GetLogger().Info("Kafka consumer started successfully.")
	}

	leaderBoardConsumer := consumer.NewLeaderboardConsumer(
		redisCl,
		consumerConfig,
	)

	leaderBoardConsumer.StartLeaderboardUpdateConsumer(dynamicConfigService.GetConfig().LeaderboardUpdateTopic)

	progressUpdateConsumer := consumer.NewProgressUpdateConsumer(
		tournamentService,
		consumerConfig,
	)

	progressUpdateConsumer.StartProgressUpdateConsumer(dynamicConfigService.GetConfig().UserProgressUpdateTopic)

	log.GetLogger().Info("Starting GoodBlast API...")

	server.NewServer(engine, database).StartHTTPServer(&config)

}

func setupSwaggerInfo(config appconfig.Config) {
	docs.SwaggerInfo.Title = "GoodBlast API"
	docs.SwaggerInfo.Description = "GoodBlast API documentation."
	docs.SwaggerInfo.Version = "v1"
	docs.SwaggerInfo.Host = config.SwaggerBaseUrl
	docs.SwaggerInfo.Schemes = []string{"http", "https"}
}

func setupGinEngine() *gin.Engine {
	engine := gin.New()
	engine.Use(middleware.CorrelationIdMiddleware)
	engine.Use(middleware.ErrorHandlerMiddleware())
	middleware.HealthCheckMiddleware(engine)
	engine.Use(gin.Recovery())
	engine.GET("/_monitoring/prometheus", gin.WrapH(promhttp.Handler()))
	return engine
}

func setupSwagger(engine *gin.Engine, config *appconfig.Config) {
	engine.GET("/swagger/*any", basicAuthForSwagger(config), ginSwagger.WrapHandler(swaggerfiles.Handler))
	engine.GET("/", func(ctx *gin.Context) {
		ctx.Redirect(http.StatusMovedPermanently, ctx.Request.URL.Host+"/swagger/index.html")
	})
}

func basicAuthForSwagger(config *appconfig.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, password, hasAuth := c.Request.BasicAuth()
		if hasAuth && user == config.SwaggerUsername && password == config.SwaggerPassword {
			c.Next()
		} else {
			c.Writer.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
	}
}

func setupDatabase(config *appconfig.Config) *bun.DB {
	pgConnectionConfig := postgres.ConnectionConfig{
		Env:          config.Env,
		Host:         config.GoodBlastDBUrl,
		Username:     config.PostgresUsername,
		Password:     config.PostgresPassword,
		DatabaseName: config.GoodBlastDBName,
	}

	log.GetLogger().Info(fmt.Sprintf("Database connection established, database: %s", config.GoodBlastDBName))
	return postgres.NewDatabaseConnection().Connect(pgConnectionConfig)
}

func setupRedis(config *appconfig.Config) *redis.Client {
	redisCl := redisclient.NewClient(config)

	log.GetLogger().Info(fmt.Sprintf("Redis connection established on %s:%s", config.RedisHost, config.RedisPort))
	return redisCl.Client
}

func setupCronJobs(tournamentService service.ITournamentService) *cron.Cron {
	c := cron.New(cron.WithLocation(time.UTC))

	c.AddFunc("59 23 * * *", func() {
		activeTournament, err := tournamentService.GetActiveTournament(context.Background())
		if err != nil {
			log.GetLogger().Error(fmt.Sprintf("Failed to fetch active tournament: %v", err))
			return
		}
		if activeTournament == nil {
			log.GetLogger().Warn("No active tournament found at 23:59.")
			return
		}

		err = tournamentService.CloseTournament(context.Background(), activeTournament.ID)
		if err != nil {
			log.GetLogger().Error(fmt.Sprintf("Failed to close tournament: %v", err))
			return
		}
		log.GetLogger().Info("Daily tournament closed at 23:59 UTC.")
		// todo: distribute rewards
	})

	c.AddFunc("0 0 * * *", func() {
		err := createAndStartNewTournament(tournamentService)
		if err != nil {
			log.GetLogger().Error(fmt.Sprintf("Failed to create/start new tournament at 00:00: %v", err))
			return
		}
		log.GetLogger().Info("New daily tournament created & started at 00:00 UTC.")
	})

	c.Start()
	return c
}

func createAndStartNewTournament(tournamentService service.ITournamentService) error {
	tournament, err := tournamentService.CreateDailyTournament(context.Background())
	if err != nil {
		return err
	}

	err = tournamentService.StartTournament(context.Background(), tournament.ID)
	if err != nil {
		return err
	}

	return nil
}
