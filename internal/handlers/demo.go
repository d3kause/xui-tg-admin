package handlers

import (
	"context"
	"strings"

	"github.com/sirupsen/logrus"
	telebot "gopkg.in/telebot.v3"

	"xui-tg-admin/internal/commands"
	"xui-tg-admin/internal/config"
	"xui-tg-admin/internal/models"
	"xui-tg-admin/internal/permissions"
	"xui-tg-admin/internal/services"
)

// DemoHandler handles demo commands
type DemoHandler struct {
	BaseHandler
	commandHandlers map[string]func(telebot.Context) error
}

// NewDemoHandler creates a new demo handler
func NewDemoHandler(
	xrayService *services.XrayService,
	stateService *services.UserStateService,
	qrService *services.QRService,
	config *config.Config,
	logger *logrus.Logger,
) *DemoHandler {
	handler := &DemoHandler{
		BaseHandler: NewBaseHandler(xrayService, stateService, qrService, config, logger),
	}

	handler.initializeCommands()
	return handler
}

// CanHandle checks if the handler can handle the given access type
func (h *DemoHandler) CanHandle(accessType permissions.AccessType) bool {
	return accessType == permissions.Demo
}

// Handle handles a message from Telegram
func (h *DemoHandler) Handle(ctx context.Context, c telebot.Context) error {
	// Get user ID
	userID := c.Sender().ID

	// Get user state
	userState, err := h.stateService.GetState(userID)
	if err != nil {
		h.logger.Errorf("Failed to get user state: %v", err)
		return err
	}

	// Handle based on state
	switch userState.State {
	case models.Default:
		return h.handleDefaultState(c)
	case models.AwaitSelectUserName:
		return h.HandleSelectServer(c)
	default:
		h.logger.Warnf("Unknown state: %d", userState.State)
		return h.handleDefaultState(c)
	}
}

// initializeCommands initializes the command handlers
func (h *DemoHandler) initializeCommands() {
	h.commandHandlers = map[string]func(telebot.Context) error{
		commands.Start:            h.handleStart,
		commands.About:            h.handleAbout,
		commands.Help:             h.handleHelp,
		commands.ReturnToMainMenu: h.handleStart,
	}
}

// getButtonCommand extracts the command from button text with emoji
func (h *DemoHandler) getButtonCommand(text string) string {
	// Check for specific button patterns
	switch text {
	case "↩️ " + commands.ReturnToMainMenu:
		return commands.ReturnToMainMenu
	case "∞ " + commands.Infinite:
		return commands.Infinite
	case "✅ " + commands.Confirm:
		return commands.Confirm
	case "❌ " + commands.Cancel:
		return commands.Cancel
	}

	// For other buttons, try to extract command after emoji
	if len(text) > 2 && text[0] != '/' {
		if spaceIndex := strings.Index(text, " "); spaceIndex > 0 {
			return text[spaceIndex+1:]
		}
	}

	return text
}

// handleDefaultState handles the default state
func (h *DemoHandler) handleDefaultState(c telebot.Context) error {
	text := c.Text()
	command := h.getButtonCommand(text)

	// Check if we have a command handler for this command
	if handler, ok := h.commandHandlers[command]; ok {
		return handler(c)
	}

	// If not, show the main menu
	return h.handleStart(c)
}

// handleStart handles the /start command
func (h *DemoHandler) handleStart(c telebot.Context) error {
	// Clear user state
	err := h.stateService.ClearState(c.Sender().ID)
	if err != nil {
		h.logger.Errorf("Failed to clear user state: %v", err)
		return err
	}

	// Show main menu
	markup := h.createMainKeyboard(permissions.Demo)
	return h.sendTextMessage(c, "Welcome to X-UI Demo Bot!\n\nThis is a demo version with limited functionality. Please contact an administrator for full access.", markup)
}

// handleSelectServer handles server selection
func (h *DemoHandler) handleSelectServer(c telebot.Context) error {
	return h.HandleSelectServer(c)
}

// handleAbout handles the About command
func (h *DemoHandler) handleAbout(c telebot.Context) error {
	aboutText := `<b>X-UI Telegram Bot</b>

This bot allows you to manage your X-ray VPN configurations through Telegram.

<b>Features:</b>
• Create and manage VPN configurations
• View traffic usage statistics
• Reset traffic usage
• Generate QR codes for easy configuration

<b>Version:</b> 1.0.0
<b>Developed by:</b> X-UI Team

For more information, please contact an administrator.`

	return h.sendTextMessage(c, aboutText, h.createReturnKeyboard())
}

// handleHelp handles the Help command
func (h *DemoHandler) handleHelp(c telebot.Context) error {
	helpText := `<b>X-UI Bot Help</b>

<b>Available Commands:</b>
• <b>/start</b> - Start the bot and show the main menu
• <b>About</b> - Show information about the bot
• <b>Help</b> - Show this help message

<b>How to get full access:</b>
Contact an administrator to get full access to the bot.

<b>Need assistance?</b>
If you need help, please contact an administrator.`

	return h.sendTextMessage(c, helpText, h.createReturnKeyboard())
}
