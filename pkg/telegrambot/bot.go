package telegrambot

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	telebot "gopkg.in/telebot.v3"

	"xui-tg-admin/internal/config"
	"xui-tg-admin/internal/handlers"
	"xui-tg-admin/internal/permissions"
	"xui-tg-admin/internal/services"
)

// Bot represents a Telegram bot
type Bot struct {
	bot          *telebot.Bot
	config       *config.Config
	handlers     map[permissions.AccessType]handlers.MessageHandler
	stateService *services.UserStateService
	permCtrl     *permissions.PermissionController
	logger       *logrus.Logger
}

// NewBot creates a new Telegram bot
func NewBot(
	cfg *config.Config,
	stateService *services.UserStateService,
	xrayService *services.XrayService,
	qrService *services.QRService,
	permCtrl *permissions.PermissionController,
	logger *logrus.Logger,
) (*Bot, error) {
	// Create bot settings
	settings := telebot.Settings{
		Token:  cfg.Telegram.Token,
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
		OnError: func(err error, c telebot.Context) {
			logger.Errorf("Telegram bot error: %v", err)
			if c != nil {
				c.Send("An error occurred. Please try again later.")
			}
		},
	}

	// Create bot instance
	b, err := telebot.NewBot(settings)
	if err != nil {
		return nil, fmt.Errorf("failed to create Telegram bot: %w", err)
	}

	// Create handler factory
	factory := handlers.NewHandlerFactory(xrayService, stateService, qrService, cfg, logger)

	// Create bot
	bot := &Bot{
		bot:          b,
		config:       cfg,
		handlers:     make(map[permissions.AccessType]handlers.MessageHandler),
		stateService: stateService,
		permCtrl:     permCtrl,
		logger:       logger,
	}

	// Initialize handlers for different access types
	bot.handlers[permissions.Admin] = factory.CreateHandler(permissions.Admin)
	bot.handlers[permissions.Member] = factory.CreateHandler(permissions.Member)
	bot.handlers[permissions.Demo] = factory.CreateHandler(permissions.Demo)

	// Setup middleware
	bot.setupMiddleware()

	return bot, nil
}

// Start starts the bot
func (b *Bot) Start(ctx context.Context) error {
	b.logger.Info("Starting Telegram bot")

	// Setup context for graceful shutdown
	go func() {
		<-ctx.Done()
		b.logger.Info("Stopping Telegram bot")
		b.bot.Stop()
	}()

	// Start the bot
	b.bot.Start()
	return nil
}

// setupMiddleware sets up the bot middleware
func (b *Bot) setupMiddleware() {
	// Add middleware for all updates
	b.bot.Use(func(next telebot.HandlerFunc) telebot.HandlerFunc {
		return func(c telebot.Context) error {
			// Log incoming message
			b.logger.Infof("Received message from %d: %s", c.Sender().ID, c.Text())

			// Pass to the next handler
			return next(c)
		}
	})

	// Handle all messages
	b.bot.Handle(telebot.OnText, b.handleUpdate)
	b.bot.Handle(telebot.OnCallback, b.handleUpdate)
	b.bot.Handle("/start", b.handleUpdate)
}

// handleUpdate handles an update from Telegram
func (b *Bot) handleUpdate(c telebot.Context) error {
	// Get user ID
	userID := c.Sender().ID

	// Get access type
	accessType := b.permCtrl.GetAccessType(userID)

	// Get handler for access type
	handler, ok := b.handlers[accessType]
	if !ok {
		b.logger.Warnf("No handler for access type %d", accessType)
		return c.Send("You don't have permission to use this bot.")
	}

	// Handle the update
	ctx := context.Background()
	return handler.Handle(ctx, c)
}
