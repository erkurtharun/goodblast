package postgres

import (
	"context"
	"crypto/tls"
	"database/sql"
	"fmt"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/extra/bundebug"
	"goodblast/pkg/log"
	"time"
)

type IDatabase interface {
	Connect(config ConnectionConfig) *bun.DB
}

type goodBlastDatabase struct{}

type ConnectionConfig struct {
	Host         string
	DatabaseName string
	Username     string
	Password     string
	Env          string
}

func NewDatabaseConnection() IDatabase {
	return &goodBlastDatabase{}
}

func (db *goodBlastDatabase) Connect(config ConnectionConfig) *bun.DB {
	connector := pgdriver.NewConnector(
		pgdriver.WithNetwork("tcp"),
		pgdriver.WithAddr(config.Host),
		pgdriver.WithTLSConfig(&tls.Config{InsecureSkipVerify: true}),
		pgdriver.WithUser(config.Username),
		pgdriver.WithPassword(config.Password),
		pgdriver.WithDatabase(config.DatabaseName),
		pgdriver.WithDialTimeout(1*time.Second),
		pgdriver.WithReadTimeout(5*time.Second),
		pgdriver.WithWriteTimeout(5*time.Second),
		pgdriver.WithConnParams(map[string]interface{}{
			"search_path": "public",
		}),
		pgdriver.WithInsecure(true),
	)
	sqlDB := sql.OpenDB(connector)

	logger := log.GetLogger()
	logger.Info("Initializing database connection...")

	if err := retryToConnect(3, 2*time.Second, func() error {
		return sqlDB.PingContext(context.Background())
	}); err != nil {
		logger.Errorf("Could not connect to database: %#v", err)
		panic(fmt.Sprintf("Could not connect to database: %#v", err))
	}

	logger.Info("Database connection established successfully")

	isDebug := config.Env == "default"
	debugHook := bundebug.NewQueryHook(
		bundebug.WithEnabled(isDebug),
		bundebug.WithVerbose(isDebug),
	)

	dbInstance := bun.NewDB(sqlDB, pgdialect.New())
	dbInstance.AddQueryHook(debugHook)
	dbInstance.SetConnMaxLifetime(5 * time.Minute)
	dbInstance.SetMaxIdleConns(10)
	dbInstance.SetMaxOpenConns(30)
	return dbInstance
}

func retryToConnect(attempts int, initialSleep time.Duration, fn func() error) error {
	sleep := initialSleep
	for i := 0; i < attempts; i++ {
		if err := fn(); err != nil {
			if i == attempts-1 {
				return err
			}
			time.Sleep(sleep)
			sleep *= 2
			continue
		}
		return nil
	}
	return fmt.Errorf("failed to connect after %d attempts", attempts)
}
