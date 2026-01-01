package firebase

import (
	"context"
	"errors"
	"os"
	"strings"

	firebase "firebase.google.com/go"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
	"google.golang.org/api/option"

	"pulse/backend/pkg/external/logging"

	"gorm.io/gorm"
)

const (
	headerAuthorization    = "Authorization"
	bearerPrefix           = "Bearer "
	envFirebaseProjectID   = "FIREBASE_PROJECT_ID"
	envFirebaseCredentials = "FIREBASE_CREDENTIALS"

	errorCodeUnauthorized  = "UNAUTHORIZED"
	errorCodeForbidden     = "FORBIDDEN"
	errorCodeInvalidConfig = "INVALID_CONFIG"

	msgNoAuthHeader      = "No authorization header"
	msgInvalidToken      = "Invalid token"
	msgAdminRequired     = "Admin access required"
	msgProjectIDNotSet   = "FIREBASE_PROJECT_ID not set"
	msgCredentialsNotSet = "FIREBASE_CREDENTIALS not set"
)

func envelopeError(c *fiber.Ctx, status int, code, message string) error {
	return c.Status(status).JSON(fiber.Map{
		"data":  nil,
		"meta":  fiber.Map{},
		"error": fiber.Map{"code": code, "message": message},
	})
}

func NewAuthMiddleware(db *gorm.DB) (fiber.Handler, error) {
	ctx := context.Background()
	projectID := os.Getenv(envFirebaseProjectID)
	
	// Development Bypass
	if os.Getenv("GO_ENV") == "development" {
		logging.Logger.Warn("⚠️ AUTHENTICATION BYPASS ENABLED (GO_ENV=development) ⚠️")
		return func(c *fiber.Ctx) error {
			// Mock admin user
			c.Locals("user_id", "dev-admin")
			return c.Next()
		}, nil
	}

	if projectID == "" {
		logging.Logger.Error(msgProjectIDNotSet)
		return nil, errors.New(msgProjectIDNotSet)
	}

	credPath := os.Getenv(envFirebaseCredentials)
	if credPath == "" {
		logging.Logger.Error(msgCredentialsNotSet)
		return nil, errors.New(msgCredentialsNotSet)
	}
	opt := option.WithCredentialsFile(credPath)
	app, err := firebase.NewApp(ctx, &firebase.Config{ProjectID: projectID}, opt)
	if err != nil {
		return nil, err
	}

	authClient, err := app.Auth(ctx)
	if err != nil {
		return nil, err
	}

	return func(c *fiber.Ctx) error {
		authHeader := c.Get(headerAuthorization)
		if authHeader == "" {
			return envelopeError(c, fiber.StatusUnauthorized, errorCodeUnauthorized, msgNoAuthHeader)
		}

		if !strings.HasPrefix(authHeader, bearerPrefix) {
			return envelopeError(c, fiber.StatusUnauthorized, errorCodeUnauthorized, msgInvalidToken)
		}

		token := strings.TrimPrefix(authHeader, bearerPrefix)
		if token == "" {
			return envelopeError(c, fiber.StatusUnauthorized, errorCodeUnauthorized, msgInvalidToken)
		}

		verifiedToken, err := authClient.VerifyIDToken(ctx, token)
		if err != nil {
			logging.Logger.Warn("Invalid ID token", zap.Error(err))
			return envelopeError(c, fiber.StatusUnauthorized, errorCodeUnauthorized, msgInvalidToken)
		}

		claimRaw, hasClaim := verifiedToken.Claims["admin"]
		isAdmin, isBool := claimRaw.(bool)
		if !hasClaim || !isBool || !isAdmin {
			return envelopeError(c, fiber.StatusForbidden, errorCodeForbidden, msgAdminRequired)
		}

		c.Locals("user_id", verifiedToken.UID)
		return c.Next()
	}, nil
}
