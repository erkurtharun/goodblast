package middleware

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
	domainErrors "goodblast/internal/domain/errors"
	"goodblast/pkg/auth"
	"goodblast/pkg/constants"
	"goodblast/pkg/log"
	"net/http"
)

func HealthCheckMiddleware(engine *gin.Engine) {
	engine.GET("/healthcheck", func(context *gin.Context) {
		context.Status(http.StatusOK)
	})
}

func CorrelationIdMiddleware(context *gin.Context) {
	correlationId := context.Request.Header.Get(constants.CorrelationIdKey)
	if correlationId == "" {
		correlationId = uuid.New().String()
	}
	context.Set(constants.CorrelationIdKey, correlationId)
	context.Next()
}

func LivenessHealthCheckMiddleware(engine *gin.Engine) {
	engine.GET("/healthcheck/liveness", func(ctx *gin.Context) {
		ctx.Status(http.StatusOK)
	})
}

func ReadinessHealthCheckMiddleware(engine *gin.Engine, pgClient *bun.DB) {
	engine.GET("/healthcheck/readiness", func(ctx *gin.Context) {
		if pgClient == nil {
			fmt.Println("Service unavailable for readiness healthcheck and pgClient=null")
			ctx.JSON(http.StatusServiceUnavailable, gin.H{"status": "unavailable"})
			return
		}

		err := pgClient.Ping()
		if err != nil {
			fmt.Println("Service unavailable for liveness healthcheck and could not ping to pgClient", err.Error())
			ctx.JSON(http.StatusServiceUnavailable, gin.H{"status": "unavailable", "error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{"status": "healthy"})

	})
}

func AuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token := ctx.GetHeader("Authorization")
		if token == "" {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization token is required"})
			ctx.Abort()
			return
		}

		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}

		payload, err := auth.GetAuth().VerifyToken(token)
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			ctx.Abort()
			return
		}

		ctx.Set("userID", payload.UserID)
		ctx.Next()
	}
}

func ErrorHandlerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if rec := recover(); rec != nil {
				log.GetLogger().Errorf("Recovered from panic: %v", rec)
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"status":  http.StatusInternalServerError,
					"message": "Internal server error",
				})
			}
		}()

		c.Next()

		if len(c.Errors) > 0 {
			err := c.Errors.Last()
			mapErrorToResponse(c, err.Err)
		}
	}
}

func mapErrorToResponse(c *gin.Context, err error) {
	var customErr *domainErrors.CustomError
	if errors.As(err, &customErr) {
		c.JSON(customErr.Status, gin.H{"status": customErr.Status, "message": customErr.Message})
	}
}
