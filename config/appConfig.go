package appconfig

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"os"
	"reflect"
)

type Config struct {
	AppName string
	Env     string
	Port    string

	// Swagger configuration
	SwaggerBaseUrl  string
	SwaggerUsername string
	SwaggerPassword string

	// Postgres configuration
	GoodBlastDBUrl   string
	GoodBlastDBName  string
	PostgresUsername string
	PostgresPassword string

	TokenSecretKey  string
	ToggleConfigURL string
	GithubToken     string

	// Kafka configuration
	KafkaBootstrapServers string
	KafkaSecurityProtocol string
	KafkaSaslMechanism    string
	KafkaSaslUsername     string
	KafkaSaslPassword     string
	KafkaClientId         string
	KafkaSessionTimeout   string
	KafkaConsumerGroupId  string
	KafkaAutoOffsetReset  string

	// Redis configuration
	RedisHost               string
	RedisPort               string
	RedisPassword           string
	RedisDB                 int
	RedisConnectionProtocol int
}

func LoadConfig() (Config, error) {
	setDefaults()
	logrus.SetOutput(os.Stdout)

	viper.AddConfigPath(".")
	viper.SetConfigName("app")
	viper.SetConfigType("env")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		logrus.Info(fmt.Scanf("Failed to load configuration file: %v. ENV values will be used.", err))
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return Config{}, err
	}

	if err := validateConfig(config); err != nil {
		return Config{}, err
	}

	// print the configuration
	logrus.Info(fmt.Sprintf("Configuration loaded successfully : %#v", config))

	return config, nil
}

func validateConfig(cfg Config) error {
	v := reflect.ValueOf(cfg)
	t := reflect.TypeOf(cfg)
	for i := 0; i < v.NumField(); i++ {
		if v.Field(i).Interface() == "" {
			return errors.New("missing required configuration: " + t.Field(i).Name)
		}
	}
	return nil
}

func setDefaults() {
	viper.SetDefault("AppName", viper.BindEnv("AppName"))
	viper.SetDefault("Env", viper.BindEnv("Env"))
	viper.SetDefault("Port", viper.BindEnv("Port"))
	viper.SetDefault("SwaggerBaseUrl", viper.BindEnv("SwaggerBaseUrl"))
	viper.SetDefault("SwaggerUsername", viper.BindEnv("SwaggerUsername"))
	viper.SetDefault("SwaggerPassword", viper.BindEnv("SwaggerPassword"))
	viper.SetDefault("GoodBlastDBUrl", viper.BindEnv("GoodBlastDBUrl"))
	viper.SetDefault("GoodBlastDBName", viper.BindEnv("GoodBlastDBName"))
	viper.SetDefault("PostgresUsername", viper.BindEnv("PostgresUsername"))
	viper.SetDefault("PostgresPassword", viper.BindEnv("PostgresPassword"))
	viper.SetDefault("TokenSecretKey", viper.BindEnv("TokenSecretKey"))
	viper.SetDefault("ToggleConfigURL", viper.BindEnv("ToggleConfigURL"))
	viper.SetDefault("GithubToken", viper.BindEnv("GithubToken"))
	viper.SetDefault("KafkaBootstrapServers", viper.BindEnv("KafkaBootstrapServers"))
	viper.SetDefault("KafkaSecurityProtocol", viper.BindEnv("KafkaSecurityProtocol"))
	viper.SetDefault("KafkaSaslMechanism", viper.BindEnv("KafkaSaslMechanism"))
	viper.SetDefault("KafkaSaslUsername", viper.BindEnv("KafkaSaslUsername"))
	viper.SetDefault("KafkaSaslPassword", viper.BindEnv("KafkaSaslPassword"))
	viper.SetDefault("KafkaClientId", viper.BindEnv("KafkaClientId"))
	viper.SetDefault("KafkaSessionTimeout", viper.BindEnv("KafkaSessionTimeout"))
	viper.SetDefault("KafkaConsumerGroupId", viper.BindEnv("KafkaConsumerGroupId"))
	viper.SetDefault("KafkaAutoOffsetReset", viper.BindEnv("KafkaAutoOffsetReset"))
	viper.SetDefault("RedisHost", viper.BindEnv("RedisHost"))
	viper.SetDefault("RedisPort", viper.BindEnv("RedisPort"))
	viper.SetDefault("RedisPassword", viper.BindEnv("RedisPassword"))
	viper.SetDefault("RedisDB", viper.BindEnv("RedisDB"))
	viper.SetDefault("RedisConnectionProtocol", viper.BindEnv("RedisConnectionProtocol"))
}
