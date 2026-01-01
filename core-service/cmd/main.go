package main

import (
	"context"
	"core-service/config"
	"core-service/models"
	"core-service/routes"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	server "core-service/controllers"
	"core-service/internal/file"
	"core-service/internal/observability/http"
	"core-service/internal/observability/logging"
	"core-service/internal/observability/metrics"
	"core-service/internal/observability/tracing"

	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

func main() {
	logger, err := logging.NewLogger(logging.Config{
		Service: "core-service",
		Env:     "dev",
		Level:   zapcore.InfoLevel,
	})
	if err != nil {
		panic(err)
	}

	otel_exporter_endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if otel_exporter_endpoint == "" {
		otel_exporter_endpoint = "jaeger:4317"
	}

	shutdown, err := tracing.Init(tracing.Config{
		ServiceName: "core-service",
		Endpoint:    otel_exporter_endpoint,
	})
	if err != nil {
		logger.Warn("tracing initialization failed", zap.Error(err))

	}
	defer shutdown(context.Background())

	if err := metrics.Init(); err != nil {
		logger.Fatal("metrics initialization failed", zap.Error(err))
	}

	err = config.ConnectDB()
	if err != nil {
		logger.Fatal("database connection failed", zap.Error(err))
	}

	sqlDB, err := config.DB.DB()
	if err != nil {
		logger.Fatal("failed to fetch DB handle", zap.Error(err))

	}

	if err := metrics.RegisterDBMetrics(sqlDB); err != nil {
		logger.Fatal("DB metrics registration failed", zap.Error(err))

	}

	if err := metrics.InitAuthMetrics(); err != nil {
		logger.Fatal("Auth metrics registration failed", zap.Error(err))

	}

	if err := metrics.InitGroupMetrics(); err != nil {
		logger.Fatal("Group metrics registration failed", zap.Error(err))
	}

	err = config.DB.AutoMigrate(&models.User{}, &models.Group{}, &models.GroupMember{}, &models.Task{}, &models.ChatMessage{}, &models.Attachment{})
	if err != nil {
		logger.Panic("automigrate failed", zap.Error(err))

	}

	grpcConfig := file.FileClientConfig{
		Addr:        os.Getenv("FILE_SERVICE_ADDR"),
		InternalKey: os.Getenv("FILE_SERVICE_KEY"),
	}

	fileClient, err := file.NewFileClient(grpcConfig)
	if err != nil {
		logger.Panic("Failed to connect to gRPC server", zap.Error(err))
	}

	chatServer := server.NewServer(fileClient, logger)
	go chatServer.Run()

	if err := metrics.InitChatMetrics(chatServer.ActiveConnections); err != nil {
		logger.Fatal("Chat metrics registration failed", zap.Error(err))
	}

	ChatHandler := server.NewChatHandler(chatServer)

	r := gin.New()

	r.Use(gin.Recovery())
	r.Use(otelgin.Middleware("core-service"))
	r.Use(http.LoggingMiddleware(logger))

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"},
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			return origin == "http://localhost:3000"
		},
		MaxAge: 12 * time.Hour,
	}))
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))
	routes.RegisterAuthRoutes(r)
	routes.RegisterGroupRoutes(r)
	routes.RegisterUserRoutes(r)
	routes.RegisterChatRoutes(r, ChatHandler)
	routes.RegisterMaterialRoutes(r, fileClient)
	r.Static("/uploads", "./uploads")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	if err := r.Run(":" + port); err != nil {
		logger.Panic("Could not start Server", zap.Error(err))
	}
}
