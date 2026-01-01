package http

import (
	"time"

	"core-service/internal/observability/logging"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

func LoggingMiddleware(base *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// ----- request_id -----
		reqID := c.GetHeader("X-Request-Id")
		if reqID == "" {
			reqID = uuid.NewString()
		}

		ctx := c.Request.Context()
		traceID := ""

		span := trace.SpanFromContext(ctx)
		if sc := span.SpanContext(); sc.IsValid() {
			traceID = sc.TraceID().String()
		}

		// ----- attach logger to context -----
		ctx = logging.WithLogger(
			c.Request.Context(),
			base,
			reqID,
			traceID,
		)
		c.Request = c.Request.WithContext(ctx)

		// ----- panic safety -----
		defer func() {
			if r := recover(); r != nil {
				logging.Logger(ctx).Error(
					"panic recovered",
					zap.Any("panic", r),
				)
				c.AbortWithStatus(500)
			}
		}()

		c.Next()

		// ----- request log (single line) -----
		logging.Logger(ctx).Info(
			"http_request",
			zap.String("method", c.Request.Method),
			zap.String("path", c.FullPath()),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("duration", time.Since(start)),
		)
	}
}
