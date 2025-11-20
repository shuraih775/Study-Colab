package controllers

import (
	"log"
	"net/http"

	"core-service/internal/file"

	"github.com/gin-gonic/gin"
)

type FileHandler struct {
	GrpcClient *file.Client
}

type UploadRequest struct {
	FileName string `json:"file_name" binding:"required"`
	FileType string `json:"file_type" binding:"required"`
}

func (h *FileHandler) GeneratePresignedURL(c *gin.Context) {
	var req UploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	presignedURL, fileID, err := h.GrpcClient.GenerateUploadURL(req.FileName, req.FileType)

	if err != nil {

		log.Printf(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to communicate with File Service"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"upload_url": presignedURL,
		"file_id":    fileID,
	})
}
