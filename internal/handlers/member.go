package handlers

import (
	"context"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	telebot "gopkg.in/telebot.v3"

	"xui-tg-admin/internal/commands"
	"xui-tg-admin/internal/config"
	"xui-tg-admin/internal/models"
	"xui-tg-admin/internal/permissions"
	"xui-tg-admin/internal/services"
)

// MemberHandler handles member commands
type MemberHandler struct {
	BaseHandler
	commandHandlers map[string]func(telebot.Context) error
}

// NewMemberHandler creates a new member handler
func NewMemberHandler(
	xrayService *services.XrayService,
	stateService *services.UserStateService,
	qrService *services.QRService,
	config *config.Config,
	logger *logrus.Logger,
) *MemberHandler {
	handler := &MemberHandler{
		BaseHandler: NewBaseHandler(xrayService, stateService, qrService, config, logger),
	}

	handler.initializeCommands()
	return handler
}

// CanHandle checks if the handler can handle the given access type
func (h *MemberHandler) CanHandle(accessType permissions.AccessType) bool {
	return accessType == permissions.Member
}

// Handle handles a message from Telegram
func (h *MemberHandler) Handle(ctx context.Context, c telebot.Context) error {
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
func (h *MemberHandler) initializeCommands() {
	h.commandHandlers = map[string]func(telebot.Context) error{
		commands.Start:            h.handleStart,
		commands.CreateNewConfig:  h.handleCreateNewConfig,
		commands.ViewConfigsInfo:  h.handleViewConfigsInfo,
		commands.ReturnToMainMenu: h.handleStart,
	}
}

// getButtonCommand extracts the command from button text with emoji
func (h *MemberHandler) getButtonCommand(text string) string {
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
func (h *MemberHandler) handleDefaultState(c telebot.Context) error {
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
func (h *MemberHandler) handleStart(c telebot.Context) error {
	// Clear user state
	err := h.stateService.ClearState(c.Sender().ID)
	if err != nil {
		h.logger.Errorf("Failed to clear user state: %v", err)
		return err
	}

	// Show main menu
	markup := h.createMainKeyboard(permissions.Member)
	return h.sendTextMessage(c, "Welcome to X-UI Member Bot!", markup)
}

// handleSelectServer handles server selection
func (h *MemberHandler) handleSelectServer(c telebot.Context) error {
	return h.HandleSelectServer(c)
}

// handleCreateNewConfig handles the Create New Config command
func (h *MemberHandler) handleCreateNewConfig(c telebot.Context) error {
	// Validate server selection
	if err := h.validateServerSelection(c.Sender().ID); err != nil {
		return h.handleSelectServer(c)
	}

	// Get subscription URL for the user's Telegram ID
	username := fmt.Sprintf("tg_%d", c.Sender().ID)
	subURL, err := h.xrayService.GetSubscriptionURL(context.Background(), username)
	if err != nil {
		h.logger.Errorf("Failed to get subscription URL: %v", err)
		return h.sendTextMessage(c, fmt.Sprintf("Failed to get subscription URL: %v", err), nil)
	}

	// Send subscription URL
	err = h.sendTextMessage(c, fmt.Sprintf("Your subscription URL:\n\n%s", subURL), h.createReturnKeyboard())
	if err != nil {
		return err
	}

	// Send QR code
	return h.sendQRCode(c, subURL)
}

// handleViewConfigsInfo handles the View Configs Info command
func (h *MemberHandler) handleViewConfigsInfo(c telebot.Context) error {
	// Validate server selection
	if err := h.validateServerSelection(c.Sender().ID); err != nil {
		return h.handleSelectServer(c)
	}

	// Get inbounds
	inbounds, err := h.xrayService.GetInbounds(context.Background())
	if err != nil {
		h.logger.Errorf("Failed to get inbounds: %v", err)
		return h.sendTextMessage(c, fmt.Sprintf("Failed to get inbounds: %v", err), nil)
	}

	// Find client with matching Telegram ID
	tgID := fmt.Sprintf("%d", c.Sender().ID)
	var found bool
	var message string

	for _, inbound := range inbounds {
		for _, clientStat := range inbound.ClientStats {
			// This is a simplified check; in a real implementation, you would need to
			// extract the client details from the inbound settings to check the TgId field
			if clientStat.Email == fmt.Sprintf("tg_%s", tgID) {
				found = true

				// Format traffic usage
				upGB := float64(clientStat.Up) / (1024 * 1024 * 1024)
				downGB := float64(clientStat.Down) / (1024 * 1024 * 1024)
				totalGB := float64(clientStat.Total) / (1024 * 1024 * 1024)

				message = fmt.Sprintf("Your configuration:\n\n"+
					"Email: %s\n"+
					"Upload: %.2f GB\n"+
					"Download: %.2f GB\n"+
					"Total: %.2f GB\n"+
					"Status: %s",
					clientStat.Email,
					upGB,
					downGB,
					totalGB,
					getStatusText(clientStat.Enable))

				break
			}
		}
		if found {
			break
		}
	}

	if !found {
		message = "You don't have any active configurations. Please use 'Create New Config' to create one."
	}

	return h.sendTextMessage(c, message, h.createReturnKeyboard())
}

// getStatusText returns a human-readable status text
func getStatusText(enabled bool) string {
	if enabled {
		return "Active"
	}
	return "Disabled"
}
