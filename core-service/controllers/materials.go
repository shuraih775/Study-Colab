package controllers

import (
	"net/http"

	"core-service/internal/file"

	"core-service/internal/observability/logging"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
)

type FileHandler struct {
	GrpcClient *file.Client
}

type UploadRequest struct {
	FileName string `json:"file_name" binding:"required"`
	FileType string `json:"file_type" binding:"required"`
}

var materialTracer = otel.Tracer("controllers.file")

func (h *FileHandler) GeneratePresignedURL(c *gin.Context) {
	ctx := c.Request.Context()
	log := logging.Logger(ctx)

	ctx, span := materialTracer.Start(ctx, "file.generate_upload_url")
	defer span.End()

	var req UploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		span.AddEvent("invalid_payload")
		log.Warn("invalid file metadata", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	span.SetAttributes(
		attribute.String("file.name", req.FileName),
		attribute.String("file.type", req.FileType),
	)

	presignedURL, fileID, err := h.GrpcClient.GenerateUploadURL(req.FileName, req.FileType)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "file service call failed")

		log.Error(
			"failed to generate upload URL from file service",
			zap.String("file_name", req.FileName),
			zap.String("file_type", req.FileType),
			zap.Error(err),
		)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to communicate with File Service",
		})
		return
	}

	span.SetAttributes(attribute.String("file.id", fileID))
	span.SetStatus(codes.Ok, "upload url generated")

	log.Info(
		"presigned url generated",
		zap.String("file_id", fileID),
		zap.String("file_name", req.FileName),
	)

	c.JSON(http.StatusOK, gin.H{
		"upload_url": presignedURL,
		"file_id":    fileID,
	})
}
