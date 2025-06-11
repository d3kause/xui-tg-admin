package handlers

import (
	"context"

	"github.com/sirupsen/logrus"
	telebot "gopkg.in/telebot.v3"

	"xui-tg-admin/internal/config"
	"xui-tg-admin/internal/permissions"
	"xui-tg-admin/internal/services"
)

// MessageHandler defines the interface for handling Telegram messages
type MessageHandler interface {
	Handle(ctx context.Context, c telebot.Context) error
	CanHandle(accessType permissions.AccessType) bool
}

// HandlerFactory creates message handlers
type HandlerFactory struct {
	xrayService  *services.XrayService
	stateService *services.UserStateService
	qrService    *services.QRService
	config       *config.Config
	logger       *logrus.Logger
}

// NewHandlerFactory creates a new handler factory
func NewHandlerFactory(
	xrayService *services.XrayService,
	stateService *services.UserStateService,
	qrService *services.QRService,
	config *config.Config,
	logger *logrus.Logger,
) *HandlerFactory {
	return &HandlerFactory{
		xrayService:  xrayService,
		stateService: stateService,
		qrService:    qrService,
		config:       config,
		logger:       logger,
	}
}

// CreateHandler creates a message handler for the given access type
func (f *HandlerFactory) CreateHandler(accessType permissions.AccessType) MessageHandler {
	switch accessType {
	case permissions.Admin:
		return NewAdminHandler(f.xrayService, f.stateService, f.qrService, f.config, f.logger)
	case permissions.Member:
		return NewMemberHandler(f.xrayService, f.stateService, f.qrService, f.config, f.logger)
	case permissions.Demo:
		return NewDemoHandler(f.xrayService, f.stateService, f.qrService, f.config, f.logger)
	default:
		f.logger.Warnf("Unknown access type: %d", accessType)
		return NewDemoHandler(f.xrayService, f.stateService, f.qrService, f.config, f.logger)
	}
}
