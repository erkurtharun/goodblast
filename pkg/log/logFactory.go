package log

import (
	"github.com/sirupsen/logrus"
	"goodblast/config"
	"os"
	"sync"
)

var (
	loggerInstance *logrus.Logger
	once           sync.Once
)

func InitLogger(config appconfig.Config) {
	once.Do(func() {
		logger := logrus.New()

		loggerMetaData := logrus.Fields{
			"app_name": config.AppName,
			"profile":  config.Env,
		}

		logger.SetLevel(logrus.InfoLevel)
		logrus.SetOutput(os.Stdout)

		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
			ForceColors:   true,
		})

		logger.AddHook(&MetadataHook{Fields: loggerMetaData})

		logger.WithFields(loggerMetaData).Info("Logger Initialized")

		loggerInstance = logger
	})
}

func GetLogger() *logrus.Logger {
	if loggerInstance == nil {
		panic("logger is not initialized. Call InitLogger first.")
	}
	return loggerInstance
}

type MetadataHook struct {
	Fields logrus.Fields
}

func (hook *MetadataHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (hook *MetadataHook) Fire(entry *logrus.Entry) error {
	for k, v := range hook.Fields {
		entry.Data[k] = v
	}
	return nil
}
