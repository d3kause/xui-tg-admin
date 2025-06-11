package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"

	"xui-tg-admin/internal/config"
	"xui-tg-admin/internal/permissions"
	"xui-tg-admin/internal/services"
	"xui-tg-admin/pkg/telegrambot"
)

func main() {
	// Setup logger
	logger := setupLogger()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("Failed to load configuration:", err)
	}

	// Initialize services
	stateService := services.NewUserStateService(logger)
	xrayService := services.NewXrayService(cfg, logger)
	qrService := services.NewQRService(logger)

	// Setup permission controller
	permController := permissions.NewController(cfg.Telegram.AdminIDs, logger)

	// Initialize bot
	bot, err := telegrambot.NewBot(cfg, stateService, xrayService, qrService, permController, logger)
	if err != nil {
		logger.Fatal("Failed to create bot:", err)
	}

	// Setup context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		<-sigCh
		logger.Info("Received shutdown signal")
		cancel()
	}()

	// Start bot
	logger.Info("Starting X-UI Telegram bot")
	if err := bot.Start(ctx); err != nil {
		logger.Fatal("Bot failed:", err)
	}
}

// setupLogger sets up the logger
func setupLogger() *logrus.Logger {
	logger := logrus.New()

	// Set log level from environment variable or default to info
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}

	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		log.Printf("Invalid log level %s, defaulting to info", logLevel)
		level = logrus.InfoLevel
	}

	logger.SetLevel(level)

	// Set formatter
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})

	return logger
}
