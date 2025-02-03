package server

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
	appconfig "goodblast/config"
	"goodblast/pkg/log"
	"net/http"
	"os/signal"
	"syscall"
	"time"
)

type Server struct {
	engine       *gin.Engine
	dbConnection *bun.DB
}

func NewServer(engine *gin.Engine, dbConnection *bun.DB) *Server {
	return &Server{
		engine:       engine,
		dbConnection: dbConnection,
	}
}

func (s *Server) StartHTTPServer(config *appconfig.Config) {
	addr := ":" + config.Port
	logger := log.GetLogger()
	logger.Infof("Starting server on http://localhost%s", addr)
	httpServer := &http.Server{
		Addr:    addr,
		Handler: s.engine,
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		logger.Info("Shutting down server")
		if err := s.dbConnection.Close(); err != nil {
			logger.Warnf("Error closing Postgres connection: %#v", err)
		} else {
			logger.Info("Postgres connection closed successfully")
		}
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			logger.Errorf("Error during server shutdown: %v", err)
		}
		logger.Info("Server exited")
	}()

	if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Errorf("Failed to start server: %v", err)
		panic("cannot start server")
	}
}
