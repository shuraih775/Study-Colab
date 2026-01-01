package controllers

import (
	"core-service/config"
	"core-service/models"
	"core-service/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"core-service/internal/observability/logging"
	"core-service/internal/observability/metrics"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

var tracer = otel.Tracer("controllers.auth")

func Register(c *gin.Context) {
	ctx := c.Request.Context()
	log := logging.Logger(ctx)

	ctx, span := tracer.Start(ctx, "auth.register")
	defer span.End()

	var input models.RegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		span.AddEvent("invalid_payload")
		log.Warn("invalid register payload", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hashedPassword, err := utils.HashPassword(input.Password)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "password hashing failed")
		log.Error("password hashing failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	user := models.User{
		Username: input.Username,
		Email:    input.Email,
		Password: hashedPassword,
	}

	if err := config.DB.WithContext(ctx).Create(&user).Error; err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "user creation failed")
		log.Error("user creation failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	span.SetAttributes(attribute.String("user.id", user.ID.String()))

	token, err := utils.GenerateToken(user.ID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "token generation failed")
		log.Error("token generation failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.SetCookie("token", token, 3600*24, "/", "localhost", false, true)

	span.SetStatus(codes.Ok, "user registered")
	log.Info("user registered", zap.String("user_id", user.ID.String()))
	c.JSON(http.StatusOK, gin.H{"message": "Registration successful"})
}

func Login(c *gin.Context) {
	ctx := c.Request.Context()
	log := logging.Logger(ctx)

	ctx, span := tracer.Start(ctx, "auth.login")
	defer span.End()

	var input models.LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		metrics.AuthLoginFailure.Add(ctx, 1)
		span.AddEvent("invalid_payload")
		log.Warn("invalid login payload", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := config.DB.WithContext(ctx).Where("email = ?", input.Email).First(&user).Error; err != nil {
		metrics.AuthLoginFailure.Add(ctx, 1)
		span.AddEvent("invalid_credentials")
		log.Warn("login failed", zap.String("email", input.Email))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	if !utils.CheckPasswordHash(input.Password, user.Password) {
		metrics.AuthLoginFailure.Add(ctx, 1)
		span.AddEvent("invalid_credentials")
		log.Warn("login failed", zap.String("user_id", user.ID.String()))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	span.SetAttributes(attribute.String("user.id", user.ID.String()))

	token, err := utils.GenerateToken(user.ID)
	if err != nil {
		metrics.AuthLoginFailure.Add(ctx, 1)
		span.RecordError(err)
		span.SetStatus(codes.Error, "token generation failed")
		log.Error("token generation failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.SetCookie("token", token, 3600*24, "/", "localhost", false, true)

	span.SetStatus(codes.Ok, "login successful")
	metrics.AuthLoginSuccess.Add(ctx, 1)
	log.Info("user logged in", zap.String("user_id", user.ID.String()))
	c.JSON(http.StatusOK, gin.H{"message": "Login successful"})
}

func ChangePassword(c *gin.Context) {
	ctx := c.Request.Context()
	log := logging.Logger(ctx)

	userID := c.MustGet("user_id").(uuid.UUID)

	ctx, span := tracer.Start(ctx, "auth.change_password")
	defer span.End()

	span.SetAttributes(attribute.String("user.id", userID.String()))

	var input struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=6"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		span.AddEvent("invalid_payload")
		log.Warn("invalid password payload", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := config.DB.WithContext(ctx).First(&user, "id = ?", userID).Error; err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "user lookup failed")
		log.Warn("user not found", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.OldPassword)); err != nil {
		span.AddEvent("invalid_old_password")
		log.Warn("old password mismatch", zap.String("user_id", userID.String()))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Old password is incorrect"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "password hashing failed")
		log.Error("password hashing failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not update password"})
		return
	}

	if err := config.DB.WithContext(ctx).Model(&user).Update("password", string(hashedPassword)).Error; err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "db update failed")
		log.Error("password update failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not update password"})
		return
	}

	span.SetStatus(codes.Ok, "password changed")
	log.Info("password updated", zap.String("user_id", userID.String()))
	c.JSON(http.StatusOK, gin.H{"message": "Password updated successfully"})
}

func Logout(c *gin.Context) {
	ctx := c.Request.Context()
	log := logging.Logger(ctx)

	userID := c.MustGet("user_id").(uuid.UUID)

	token, err := utils.GenerateToken(userID)
	if err != nil {
		log.Error(
			"token generation failed",
			zap.String("user_id", userID.String()),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.SetCookie("token", token, 0, "/", "localhost", false, true)

	log.Info(
		"user logged out",
		zap.String("user_id", userID.String()),
	)
	c.JSON(http.StatusOK, gin.H{"message": "Logout successful"})
}

func GetCurrentUser(c *gin.Context) {
	log := logging.Logger(c.Request.Context())

	user, exists := c.Get("user")
	if user == nil || exists == false {
		log.Error(
			"user missing from context",
		)
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Could not get User from context"})

	}
	c.JSON(http.StatusOK, user)
}

func DeleteUser(c *gin.Context) {
	ctx := c.Request.Context()
	log := logging.Logger(ctx)

	userID := c.MustGet("user_id").(uuid.UUID)

	ctx, span := tracer.Start(ctx, "auth.delete_user")
	defer span.End()

	span.SetAttributes(
		attribute.String("user.id", userID.String()),
	)

	if err := config.DB.WithContext(ctx).
		Delete(&models.User{}, "id = ?", userID).Error; err != nil {

		span.RecordError(err)
		span.SetStatus(codes.Error, "user deletion failed")

		log.Error("user deletion failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	span.SetStatus(codes.Ok, "user deleted")

	log.Info(
		"user deleted",
		zap.String("user_id", userID.String()),
	)

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}
