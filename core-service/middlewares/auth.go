package middlewares

import (
	"net/http"

	"core-service/config"
	"core-service/internal/observability/logging"
	"core-service/models"
	"core-service/utils"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
)

var jwtTracer = otel.Tracer("middlewares.auth")

func JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		log := logging.Logger(ctx)

		ctx, span := jwtTracer.Start(ctx, "auth.jwt")
		defer span.End()

		token, err := c.Cookie("token")
		if err != nil || token == "" {
			span.AddEvent("missing_token")
			log.Warn("missing auth token")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
			return
		}

		claims, err := utils.ParseToken(token)
		if err != nil {
			span.AddEvent("invalid_token")
			log.Warn("invalid auth token", zap.Error(err))
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		span.SetAttributes(
			attribute.String("user.id", claims.UserID.String()),
		)

		var user models.User
		if err := config.DB.WithContext(ctx).First(&user, claims.UserID).Error; err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "user lookup failed")

			log.Warn(
				"user not found for token",
				zap.String("user_id", claims.UserID.String()),
			)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
			return
		}

		span.SetStatus(codes.Ok, "authenticated")

		c.Set("user", user)
		c.Set("user_id", claims.UserID)

		c.Next()
	}
}
