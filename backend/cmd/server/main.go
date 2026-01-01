package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"pulse/backend/internal/api"
	"pulse/backend/internal/ca"
	"pulse/backend/internal/db/migrations"
	"pulse/backend/internal/services"
	"pulse/backend/pkg/external/firebase"
	pulsegrpc "pulse/backend/pkg/external/grpc"
	"pulse/backend/pkg/external/logging"
)

const (
	DefaultHTTPPort = "8080"
	DefaultGRPCPort = "50051"
	DefaultLogPath  = "logs/pulse.log"
)

func main() {
	// Load environment variables
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL not set")
	}
	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort = DefaultHTTPPort
	}
	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = DefaultGRPCPort
	}
	logPath := os.Getenv("LOG_PATH")
	if logPath == "" {
		logPath = DefaultLogPath
	}

	// Initialize database
	db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Run migrations
	if err := migrations.Migrate(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize logger
	if err := logging.InitLogger(logPath, db); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	// Initialize services
	caSvc := ca.NewCAService(db)
	identitySvc := services.NewIdentityService(db)
	taskSvc := services.NewTaskService(db)

	// Initialize Firebase middleware
	authMiddleware, err := firebase.NewAuthMiddleware(db)
	if err != nil {
		log.Fatalf("Failed to initialize auth middleware: %v", err)
	}

	// Setup Fiber app
	app := fiber.New()
	app.Use(authMiddleware)
	api.SetupRoutes(app, authMiddleware, identitySvc, taskSvc)
	api.TasksRoutes(app, authMiddleware, taskSvc)

	// Setup gRPC server
	grpcServer, err := pulsegrpc.NewGRPCServer(caSvc, identitySvc, taskSvc)
	if err != nil {
		log.Fatalf("Failed to create gRPC server: %v", err)
	}

	// Start servers
	go func() {
		if err := app.Listen(":" + httpPort); err != nil {
			logging.Logger.Error("HTTP server error", zap.Error(err))
		}
	}()

	go func() {
		lis, err := net.Listen("tcp", ":"+grpcPort)
		if err != nil {
			log.Fatalf("Failed to listen for gRPC: %v", err)
		}
		if err := grpcServer.Serve(lis); err != nil {
			logging.Logger.Error("gRPC server error", zap.Error(err))
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	grpcServer.GracefulStop()

	if err := app.ShutdownWithContext(ctx); err != nil {
		logging.Logger.Error("HTTP server shutdown error", zap.Error(err))
	}

	logging.Logger.Info("Server shutdown complete")
}
