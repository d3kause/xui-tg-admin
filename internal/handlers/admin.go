package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	telebot "gopkg.in/telebot.v3"

	"xui-tg-admin/internal/commands"
	"xui-tg-admin/internal/config"
	"xui-tg-admin/internal/helpers"
	"xui-tg-admin/internal/models"
	"xui-tg-admin/internal/permissions"
	"xui-tg-admin/internal/services"
	"xui-tg-admin/internal/validation"
)

// AdminHandler handles admin commands
type AdminHandler struct {
	BaseHandler
	commandHandlers map[string]func(telebot.Context) error
}

// NewAdminHandler creates a new admin handler
func NewAdminHandler(
	xrayService *services.XrayService,
	stateService *services.UserStateService,
	qrService *services.QRService,
	config *config.Config,
	logger *logrus.Logger,
) *AdminHandler {
	handler := &AdminHandler{
		BaseHandler: NewBaseHandler(xrayService, stateService, qrService, config, logger),
	}

	handler.initializeCommands()
	return handler
}

// CanHandle checks if the handler can handle the given access type
func (h *AdminHandler) CanHandle(accessType permissions.AccessType) bool {
	return accessType == permissions.Admin
}

// Handle handles a message from Telegram
func (h *AdminHandler) Handle(ctx context.Context, c telebot.Context) error {
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
	case models.AwaitingInputUserName:
		return h.processUserName(c)
	case models.AwaitingDuration:
		return h.processDuration(c)
	case models.AwaitSelectUserName:
		return h.processSelectUser(c)
	case models.AwaitMemberAction:
		return h.processMemberAction(c)
	case models.AwaitConfirmMemberDeletion:
		return h.processConfirmDeletion(c)
	case models.AwaitConfirmResetUsersNetworkUsage:
		return h.processConfirmResetUsersNetworkUsage(c)
	default:
		h.logger.Warnf("Unknown state: %d", userState.State)
		return h.handleDefaultState(c)
	}
}

// initializeCommands initializes the command handlers
func (h *AdminHandler) initializeCommands() {
	h.commandHandlers = map[string]func(telebot.Context) error{
		commands.Start:             h.handleStart,
		commands.AddMember:         h.handleAddMember,
		commands.EditMember:        h.handleEditMember,
		commands.DeleteMember:      h.handleDeleteMember,
		commands.OnlineMembers:     h.handleGetOnlineMembers,
		commands.NetworkUsage:      h.handleGetUsersNetworkUsage,
		commands.DetailedUsage:     h.handleGetDetailedUsersInfo,
		commands.ResetNetworkUsage: h.handleResetUsersNetworkUsage,
		commands.ReturnToMainMenu:  h.handleStart,
		commands.Cancel:            h.handleStart,
	}
}

// getButtonCommand extracts the command from button text with emoji
func (h *AdminHandler) getButtonCommand(text string) string {
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
	case "🔗 " + commands.ViewConfig:
		return commands.ViewConfig
	case "🔄 " + commands.ResetTraffic:
		return commands.ResetTraffic
	case "🗑️ " + commands.Delete:
		return commands.Delete
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
func (h *AdminHandler) handleDefaultState(c telebot.Context) error {
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
func (h *AdminHandler) handleStart(c telebot.Context) error {
	// Clear user state
	err := h.stateService.ClearState(c.Sender().ID)
	if err != nil {
		h.logger.Errorf("Failed to clear user state: %v", err)
		return err
	}

	// Get user state
	_, err = h.stateService.GetState(c.Sender().ID)
	if err != nil {
		h.logger.Errorf("Failed to get user state: %v", err)
		return err
	}

	// Show main menu with welcome message only for /start command
	markup := h.createMainKeyboard(permissions.Admin)
	if c.Text() == commands.Start {
		return h.sendTextMessage(c, "🚀 <b>Welcome to X-UI Admin Panel!</b>\n\nYou have administrator privileges. Use the menu below to manage your VPN users, monitor connections, and configure settings.", markup)
	}

	// For return to main menu, show only the keyboard without any message
	return h.sendTextMessage(c, "🏠 <b>Main Menu</b>\n\nSelect an action:", markup)
}

// handleAddMember handles the Add Member command
func (h *AdminHandler) handleAddMember(c telebot.Context) error {

	// Set state to awaiting username
	err := h.stateService.WithConversationState(c.Sender().ID, models.AwaitingInputUserName)
	if err != nil {
		h.logger.Errorf("Failed to set state: %v", err)
		return err
	}

	// Show return keyboard
	markup := h.createReturnKeyboard()
	return h.sendTextMessage(c, "👤 <b>Add New User</b>\n\n📝 Please enter a username for the new user:\n\n<i>• Use only letters, numbers, and underscores\n• 3-20 characters long\n• Example: john_doe, user123</i>", markup)
}

// handleEditMember handles the Edit Member command
func (h *AdminHandler) handleEditMember(c telebot.Context) error {
	// Проверяем доступность сервиса
	_, err := h.stateService.GetState(c.Sender().ID)
	if err != nil {
		h.logger.Errorf("Failed to get user state: %v", err)
		return err
	}

	// Показываем список пользователей с сортировкой по дате добавления
	return h.showMembersWithSort(c, models.SortByCreationOrder, "edit")
}

// handleDeleteMember handles the Delete Member command
func (h *AdminHandler) handleDeleteMember(c telebot.Context) error {
	// Проверяем доступность сервиса
	_, err := h.stateService.GetState(c.Sender().ID)
	if err != nil {
		h.logger.Errorf("Failed to get user state: %v", err)
		return err
	}

	// Показываем список пользователей с сортировкой по дате добавления
	return h.showMembersWithSort(c, models.SortByCreationOrder, "delete")
}

// handleGetOnlineMembers handles the Online Members command
func (h *AdminHandler) handleGetOnlineMembers(c telebot.Context) error {

	// Get online users
	onlineUsers, err := h.xrayService.GetOnlineUsers(context.Background())
	if err != nil {
		h.logger.Errorf("Failed to get online users: %v", err)
		return h.sendTextMessage(c, "❌ <b>Connection Error</b>\n\nCouldn't retrieve online users. Please check your server connection and try again.", h.createMainKeyboard(permissions.Admin))
	}

	// Format message
	var message string
	if len(onlineUsers) == 0 {
		message = "💤 <b>No Active Connections</b>\n\nNo users are currently connected to the VPN server."
	} else {
		message = fmt.Sprintf("🟢 <b>Active Connections (%d)</b>\n\n", len(onlineUsers))
		for _, user := range onlineUsers {
			message += fmt.Sprintf("👤 %s\n", user)
		}
	}

	return h.sendTextMessage(c, message, h.createMainKeyboard(permissions.Admin))
}

// handleGetUsersNetworkUsage handles the Network Usage command
func (h *AdminHandler) handleGetUsersNetworkUsage(c telebot.Context) error {

	// Get inbounds
	inbounds, err := h.xrayService.GetInbounds(context.Background())
	if err != nil {
		h.logger.Errorf("Failed to get inbounds: %v", err)
		return h.sendTextMessage(c, "❌ <b>Connection Error</b>\n\nCouldn't retrieve network usage data. Please check your server connection and try again.", h.createReturnKeyboard())
	}

	// Format beautiful network usage report
	message := helpers.FormatNetworkUsageReport(inbounds)

	return h.sendTextMessage(c, message, h.createReturnKeyboard())
}

// handleResetUsersNetworkUsage handles the Reset Network Usage command
func (h *AdminHandler) handleResetUsersNetworkUsage(c telebot.Context) error {
	// Set state to awaiting confirmation for reset
	err := h.stateService.WithConversationState(c.Sender().ID, models.AwaitConfirmResetUsersNetworkUsage)
	if err != nil {
		h.logger.Errorf("Failed to set state: %v", err)
		return err
	}

	// Show confirm keyboard
	markup := h.createConfirmKeyboard()
	return h.sendTextMessage(c, "⚠️ <b>Reset All Network Usage</b>\n\nThis will reset traffic statistics for <b>ALL users</b> in the system.\n\n<b>⚠️ This action cannot be undone!</b>\n\nAre you sure you want to proceed?", markup)
}

// processUserName processes the username input
func (h *AdminHandler) processUserName(c telebot.Context) error {
	// Get username from message
	username := c.Text()

	// Check for return to main menu
	if h.getButtonCommand(username) == commands.ReturnToMainMenu {
		return h.handleStart(c)
	}

	// Validate username format
	if err := validation.ValidateUsername(username); err != nil {
		return h.sendTextMessage(c, fmt.Sprintf("❌ <b>Invalid Username</b>\n\n%s\n\n💡 <b>Requirements:</b>\n• 3-20 characters\n• Letters, numbers, underscores only\n• Example: john_doe, user123\n\nPlease try again:", err.Error()), h.createReturnKeyboard())
	}

	// Store username in state
	err := h.stateService.WithPayload(c.Sender().ID, username)
	if err != nil {
		h.logger.Errorf("Failed to set payload: %v", err)
		return err
	}

	// Set state to awaiting duration
	err = h.stateService.WithConversationState(c.Sender().ID, models.AwaitingDuration)
	if err != nil {
		h.logger.Errorf("Failed to set state: %v", err)
		return err
	}

	// Create keyboard with Infinite option
	markup := &telebot.ReplyMarkup{
		ResizeKeyboard: true,
	}
	markup.Reply(
		telebot.Row{
			telebot.Btn{Text: "∞ " + commands.Infinite},
		},
		telebot.Row{
			telebot.Btn{Text: "↩️ " + commands.ReturnToMainMenu},
		},
	)

	return h.sendTextMessage(c, fmt.Sprintf("⏰ <b>Set Duration for %s</b>\n\n📅 Enter subscription duration in days:\n\n<i>• Example: 30 (for 30 days)\n• Maximum: 3650 days\n• Or choose Infinite for unlimited time</i>", username), markup)
}

// processDuration processes the duration input
func (h *AdminHandler) processDuration(c telebot.Context) error {
	// Get duration from message
	durationStr := c.Text()

	// Check for return to main menu
	if h.getButtonCommand(durationStr) == commands.ReturnToMainMenu {
		return h.handleStart(c)
	}

	// Extract command from button text
	durationStr = h.getButtonCommand(durationStr)

	// Get user state
	userState, err := h.stateService.GetState(c.Sender().ID)
	if err != nil {
		h.logger.Errorf("Failed to get user state: %v", err)
		return err
	}

	// Get username from state
	if userState.Payload == nil {
		return h.sendTextMessage(c, "❌ <b>Session Error</b>\n\nUsername data was lost. Please start over.", h.createReturnKeyboard())
	}

	baseUsername := *userState.Payload

	// Get enabled inbounds
	enabledInbounds, err := h.getEnabledInbounds(context.Background())
	if err != nil {
		h.logger.Errorf("Failed to get enabled inbounds: %v", err)
		return h.sendTextMessage(c, "❌ <b>Server Configuration Error</b>\n\nNo enabled inbound connections found. Please check your server configuration or contact the administrator.", h.createReturnKeyboard())
	}

	// Calculate expiry time
	expiryTime, err := calculateExpiryTime(durationStr)
	if err != nil {
		return h.sendTextMessage(c, fmt.Sprintf("❌ <b>Invalid Duration</b>\n\n%s\n\n💡 <b>Valid formats:</b>\n• Number: 30 (for 30 days)\n• Range: 1-3650 days\n• Or use the Infinite button\n\nPlease try again:", err.Error()), h.createReturnKeyboard())
	}

	// Create client creation parameters
	params := ClientCreationParams{
		BaseUsername:    baseUsername,
		DurationStr:     durationStr,
		ExpiryTime:      expiryTime,
		CommonSubId:     models.GenerateSubID(),
		BaseFingerprint: fmt.Sprintf("%x", time.Now().UnixNano()),
		SenderID:        c.Sender().ID,
	}

	// Send loading message
	loadingMsg, _ := h.sendTextMessageWithReturn(c, "⏳ <b>Creating User...</b>\n\nPlease wait while we set up the new user configuration across all servers.", nil)

	// Create clients for all enabled inbounds
	createdEmails, addErrors, addedToAny := h.createClientsForAllInbounds(context.Background(), params, enabledInbounds)

	// Delete loading message
	if loadingMsg != nil {
		c.Bot().Delete(loadingMsg)
	}

	if !addedToAny {
		return h.sendTextMessage(c, fmt.Sprintf("❌ <b>User Creation Failed</b>\n\nCouldn't create user '%s' in any server configuration.\n\n<b>Errors:</b>\n%s\n\nPlease check server configuration or try again later.", baseUsername, strings.Join(addErrors, "\n")), h.createReturnKeyboard())
	}

	// Send subscription information and QR code
	return h.sendSubscriptionInfo(c, params, createdEmails, addErrors)
}

// processSelectUser processes the user selection
func (h *AdminHandler) processSelectUser(c telebot.Context) error {
	// Get username from message
	username := c.Text()

	// Check for return to main menu
	if h.getButtonCommand(username) == commands.ReturnToMainMenu {
		return h.handleStart(c)
	}

	// Store username in state
	err := h.stateService.WithPayload(c.Sender().ID, username)
	if err != nil {
		h.logger.Errorf("Failed to set payload: %v", err)
		return err
	}

	// Set state to awaiting member action
	err = h.stateService.WithConversationState(c.Sender().ID, models.AwaitMemberAction)
	if err != nil {
		h.logger.Errorf("Failed to set state: %v", err)
		return err
	}

	// Create action keyboard
	markup := h.createUserActionKeyboard()

	return h.sendTextMessage(c, fmt.Sprintf("👤 <b>Managing User: %s</b>\n\n🎛️ Choose an action:", username), markup)
}

// processMemberAction processes the member action selection
func (h *AdminHandler) processMemberAction(c telebot.Context) error {
	// Get action from message
	action := c.Text()

	// Check for return to main menu first
	if h.getButtonCommand(action) == commands.ReturnToMainMenu {
		return h.handleStart(c)
	}

	// Проверяем доступность сервиса
	userState, err := h.stateService.GetState(c.Sender().ID)
	if err != nil {
		h.logger.Errorf("Failed to get user state: %v", err)
		return err
	}

	// Get username from state
	if userState.Payload == nil {
		return h.sendTextMessage(c, "❌ <b>Session Error</b>\n\nUser data was lost. Please start over.", h.createReturnKeyboard())
	}

	username := *userState.Payload

	// Extract command from button text
	command := h.getButtonCommand(action)

	// Handle action
	switch command {
	case commands.ViewConfig:
		return h.handleViewConfig(c, username)
	case commands.ResetTraffic:
		return h.handleResetTraffic(c, username)
	case commands.Delete:
		return h.handleConfirmDelete(c, username)
	default:
		return h.sendTextMessage(c, "❌ <b>Invalid Action</b>\n\nPlease select one of the available options from the menu.", h.createUserActionKeyboard())
	}
}

// createUserActionKeyboard creates a keyboard for user actions
func (h *AdminHandler) createUserActionKeyboard() *telebot.ReplyMarkup {
	markup := &telebot.ReplyMarkup{
		ResizeKeyboard: true,
	}

	markup.Reply(
		telebot.Row{
			telebot.Btn{Text: "🔗 " + commands.ViewConfig},
		},
		telebot.Row{
			telebot.Btn{Text: "🔄 " + commands.ResetTraffic},
			telebot.Btn{Text: "🗑️ " + commands.Delete},
		},
		telebot.Row{
			telebot.Btn{Text: "↩️ " + commands.ReturnToMainMenu},
		},
	)

	return markup
}

// handleViewConfig handles the View Config action
func (h *AdminHandler) handleViewConfig(c telebot.Context, username string) error {
	h.logger.Infof("Starting view config for user: %s", username)

	// Get all inbounds
	inbounds, err := h.xrayService.GetInbounds(context.Background())
	if err != nil {
		h.logger.Errorf("Failed to get inbounds: %v", err)
		return h.sendTextMessage(c, fmt.Sprintf("Failed to get inbounds: %v", err), h.createUserActionKeyboard())
	}

	// Find first client with the base username to get SubID
	var foundClientSubID string

	for _, inbound := range inbounds {
		// Parse inbound settings to get client details
		var settings models.InboundSettings
		if err := json.Unmarshal([]byte(inbound.Settings), &settings); err != nil {
			h.logger.Errorf("Failed to parse settings for inbound %d: %v", inbound.ID, err)
			continue
		}

		// Find client in settings
		for _, client := range settings.Clients {
			// Check if client email matches the base username using helper function
			if helpers.IsEmailMatchingBaseUsername(client.Email, username) {
				h.logger.Infof("Found matching client: %s in inbound %d", client.Email, inbound.ID)
				foundClientSubID = client.SubID
				break
			}
		}
		if foundClientSubID != "" {
			break
		}
	}

	if foundClientSubID == "" {
		return h.sendTextMessage(c, fmt.Sprintf("❌ <b>User Not Found</b>\n\nNo configuration found for user '%s'. The user may have been deleted or never existed.", username), h.createUserActionKeyboard())
	}

	// Get subscription URL using SubID (same format as when adding user)
	subURL := fmt.Sprintf("https://iris.xele.one:2096/sub/%s?name=%s", foundClientSubID, foundClientSubID)

	// Send subscription URL with user action keyboard (stays in same state)
	err = h.sendTextMessage(c, fmt.Sprintf("🔗 <b>Configuration for %s</b>\n\n📋 <b>Subscription URL:</b>\n<code>%s</code>\n\n<i>Copy this link to your VPN client or scan the QR code below</i>", username, subURL), h.createUserActionKeyboard())
	if err != nil {
		return err
	}

	// Send QR code
	return h.sendQRCode(c, subURL)
}

// handleResetTraffic handles the Reset Traffic action
func (h *AdminHandler) handleResetTraffic(c telebot.Context, username string) error {
	h.logger.Infof("Starting reset traffic for user: %s", username)

	// Send loading message
	loadingMsg, _ := h.sendTextMessageWithReturn(c, fmt.Sprintf("⏳ <b>Resetting Traffic...</b>\n\nResetting traffic statistics for user '%s'. Please wait...", username), nil)

	// Get all inbounds
	inbounds, err := h.xrayService.GetInbounds(context.Background())
	if err != nil {
		h.logger.Errorf("Failed to get inbounds: %v", err)
		return h.sendTextMessage(c, "❌ <b>Connection Error</b>\n\nCouldn't retrieve server data. Please check your connection and try again.", h.createUserActionKeyboard())
	}

	// Find all clients with the base username and reset their traffic
	var resetErrors []string
	successfullyReset := 0

	for _, inbound := range inbounds {
		for _, clientStat := range inbound.ClientStats {
			// Check if client email matches the base username using helper function
			if helpers.IsEmailMatchingBaseUsername(clientStat.Email, username) {
				h.logger.Infof("Found matching client: %s in inbound %d", clientStat.Email, inbound.ID)

				err := h.xrayService.ResetUserTraffic(context.Background(), inbound.ID, clientStat.Email)
				if err != nil {
					h.logger.Errorf("Failed to reset traffic for %s in inbound %d: %v", clientStat.Email, inbound.ID, err)
					resetErrors = append(resetErrors, fmt.Sprintf("Failed to reset %s in inbound %d: %v", clientStat.Email, inbound.ID, err))
				} else {
					h.logger.Infof("Successfully reset traffic for %s in inbound %d", clientStat.Email, inbound.ID)
					successfullyReset++
				}
			}
		}
	}

	// Send result message
	var message string
	if successfullyReset > 0 {
		message = fmt.Sprintf("✅ <b>Traffic Reset Complete</b>\n\n🔄 Successfully reset traffic for user <b>%s</b> (%d configurations)", username, successfullyReset)
		if len(resetErrors) > 0 {
			message += fmt.Sprintf("\n\n⚠️ <b>Some errors occurred:</b>\n%s", strings.Join(resetErrors, "\n"))
		}
	} else {
		message = fmt.Sprintf("❌ <b>Reset Failed</b>\n\nNo active configurations found for user '%s'.", username)
		if len(resetErrors) > 0 {
			message += fmt.Sprintf("\n\n<b>Errors:</b>\n%s", strings.Join(resetErrors, "\n"))
		}
	}

	// Delete loading message
	if loadingMsg != nil {
		c.Bot().Delete(loadingMsg)
	}

	return h.sendTextMessage(c, message, h.createUserActionKeyboard())
}

// handleConfirmDelete handles the Delete action
func (h *AdminHandler) handleConfirmDelete(c telebot.Context, username string) error {
	// Установить состояние подтверждения удаления
	err := h.stateService.WithConversationState(c.Sender().ID, models.AwaitConfirmMemberDeletion)
	if err != nil {
		h.logger.Errorf("Failed to set state: %v", err)
		return err
	}
	// Показать клавиатуру подтверждения
	markup := h.createConfirmKeyboard()
	return h.sendTextMessage(c, fmt.Sprintf("🗑️ <b>Confirm User Deletion</b>\n\n⚠️ You are about to permanently delete user <b>%s</b>\n\n<b>This action will:</b>\n• Remove user from all server configurations\n• Delete all associated data\n• Cannot be undone\n\nAre you absolutely sure?", username), markup)
}

// processConfirmDeletion processes the deletion confirmation
func (h *AdminHandler) processConfirmDeletion(c telebot.Context) error {
	// Get confirmation from message
	confirmation := c.Text()

	// Check for return to main menu
	if h.getButtonCommand(confirmation) == commands.ReturnToMainMenu {
		return h.handleStart(c)
	}

	// Check if user confirmed
	if h.getButtonCommand(confirmation) != commands.Confirm {
		return h.sendTextMessage(c, "❌ <b>Invalid Selection</b>\n\nPlease click Confirm to proceed with deletion or use the Return button to cancel.", h.createConfirmKeyboard())
	}

	// Get user state to get the username we want to delete
	userState, err := h.stateService.GetState(c.Sender().ID)
	if err != nil {
		h.logger.Errorf("Failed to get user state: %v", err)
		return err
	}

	if userState.Payload == nil {
		return h.sendTextMessage(c, "❌ <b>Session Error</b>\n\nUser data was lost. Please start the deletion process again.", h.createReturnKeyboard())
	}

	username := *userState.Payload

	// Send loading message
	loadingMsg, _ := h.sendTextMessageWithReturn(c, fmt.Sprintf("⏳ <b>Deleting User...</b>\n\nRemoving user '%s' from all server configurations. Please wait...", username), nil)

	// Delete client using email
	err = h.xrayService.RemoveClients(context.Background(), []string{username})
	// Delete loading message
	if loadingMsg != nil {
		c.Bot().Delete(loadingMsg)
	}

	if err != nil {
		h.logger.Errorf("Failed to delete client: %v", err)
		return h.sendTextMessage(c, fmt.Sprintf("❌ <b>Deletion Failed</b>\n\nCouldn't delete user '%s'. Please try again or contact administrator.\n\n<b>Error:</b> %v", username, err), h.createReturnKeyboard())
	}

	return h.sendTextMessage(c, fmt.Sprintf("✅ <b>User Deleted Successfully</b>\n\n🗑️ User '%s' has been permanently removed from all server configurations.", username), h.createReturnKeyboard())
}

// handleGetDetailedUsersInfo handles the Detailed Usage command
func (h *AdminHandler) handleGetDetailedUsersInfo(c telebot.Context) error {

	// Get inbounds
	inbounds, err := h.xrayService.GetInbounds(context.Background())
	if err != nil {
		h.logger.Errorf("Failed to get inbounds: %v", err)
		return h.sendTextMessage(c, "❌ <b>Connection Error</b>\n\nCouldn't retrieve detailed usage data. Please check your server connection and try again.", h.createMainKeyboard(permissions.Admin))
	}

	// Format detailed user information report
	message := helpers.FormatDetailedUsersReport(inbounds)

	return h.sendTextMessage(c, message, h.createMainKeyboard(permissions.Admin))
}

// createConfirmKeyboard creates a keyboard for confirmation
func (h *AdminHandler) createConfirmKeyboard() *telebot.ReplyMarkup {
	markup := &telebot.ReplyMarkup{
		ResizeKeyboard: true,
	}

	markup.Reply(
		telebot.Row{
			telebot.Btn{Text: "✅ " + commands.Confirm},
		},
		telebot.Row{
			telebot.Btn{Text: "↩️ " + commands.ReturnToMainMenu},
		},
	)

	return markup
}

// processConfirmResetUsersNetworkUsage processes the confirmation for resetting network usage
func (h *AdminHandler) processConfirmResetUsersNetworkUsage(c telebot.Context) error {
	// Get confirmation from message
	confirmation := c.Text()

	// Check for return to main menu
	if h.getButtonCommand(confirmation) == commands.ReturnToMainMenu {
		return h.handleStart(c)
	}

	// Check if user confirmed
	if h.getButtonCommand(confirmation) != commands.Confirm {
		return h.sendTextMessage(c, "❌ <b>Invalid Selection</b>\n\nPlease click Confirm to proceed with reset or use the Return button to cancel.", h.createConfirmKeyboard())
	}

	h.logger.Infof("Starting reset network usage for all users")

	// Send loading message
	loadingMsg, _ := h.sendTextMessageWithReturn(c, "⏳ <b>Resetting All Traffic...</b>\n\nThis may take a few moments. Resetting traffic statistics for all users across all servers...", nil)

	// Get all inbounds
	inbounds, err := h.xrayService.GetInbounds(context.Background())
	if err != nil {
		h.logger.Errorf("Failed to get inbounds: %v", err)
		return h.sendTextMessage(c, "❌ <b>Connection Error</b>\n\nCouldn't retrieve server data for reset operation. Please check your connection and try again.", h.createMainKeyboard(permissions.Admin))
	}

	// Collect all user emails from all inbounds
	var userEmails []struct {
		inboundID int
		email     string
	}

	for _, inbound := range inbounds {
		for _, clientStat := range inbound.ClientStats {
			userEmails = append(userEmails, struct {
				inboundID int
				email     string
			}{
				inboundID: inbound.ID,
				email:     clientStat.Email,
			})
		}
	}

	if len(userEmails) == 0 {
		return h.sendTextMessage(c, "📭 <b>No Users Found</b>\n\nThere are no users in the system to reset traffic for.", h.createMainKeyboard(permissions.Admin))
	}

	h.logger.Infof("Found %d users to reset traffic", len(userEmails))

	// Reset traffic for all users
	var resetErrors []string
	successfullyReset := 0

	for _, user := range userEmails {
		err := h.xrayService.ResetUserTraffic(context.Background(), user.inboundID, user.email)
		if err != nil {
			h.logger.Errorf("Failed to reset traffic for %s in inbound %d: %v", user.email, user.inboundID, err)
			resetErrors = append(resetErrors, fmt.Sprintf("Failed to reset %s in inbound %d: %v", user.email, user.inboundID, err))
		} else {
			h.logger.Infof("Successfully reset traffic for %s in inbound %d", user.email, user.inboundID)
			successfullyReset++
		}
	}

	// Send result message
	var message string
	if successfullyReset > 0 {
		message = fmt.Sprintf("✅ <b>Mass Traffic Reset Complete</b>\n\n🔄 Successfully reset traffic for <b>%d users</b>\n\n<i>All user traffic counters have been set to zero</i>", successfullyReset)
		if len(resetErrors) > 0 {
			message += fmt.Sprintf("\n\n⚠️ <b>Some errors occurred:</b>\n%s", strings.Join(resetErrors, "\n"))
		}
	} else {
		message = fmt.Sprintf("❌ <b>Mass Reset Failed</b>\n\nCouldn't reset traffic for any users.\n\n<b>Errors:</b>\n%s", strings.Join(resetErrors, "\n"))
	}

	// Delete loading message
	if loadingMsg != nil {
		c.Bot().Delete(loadingMsg)
	}

	// Clear user state and return to main menu
	err = h.stateService.ClearState(c.Sender().ID)
	if err != nil {
		h.logger.Errorf("Failed to clear user state: %v", err)
	}

	return h.sendTextMessage(c, message, h.createMainKeyboard(permissions.Admin))
}

// showSortTypeMenu показывает меню выбора типа сортировки
func (h *AdminHandler) showSortTypeMenu(c telebot.Context, messageText string) error {
	markup := &telebot.ReplyMarkup{
		ResizeKeyboard: true,
	}

	sortTypes := []models.SortType{
		models.SortByCreationOrder,
		models.SortByExpiryDate,
		models.SortByTrafficTotal,
		models.SortByStatus,
		models.SortByName,
	}

	var rows []telebot.Row
	for _, sortType := range sortTypes {
		rows = append(rows, telebot.Row{telebot.Btn{Text: sortType.GetSortName()}})
	}

	// Add return button
	rows = append(rows, telebot.Row{telebot.Btn{Text: "↩️ " + commands.ReturnToMainMenu}})

	markup.Reply(rows...)

	return h.sendTextMessage(c, messageText, markup)
}

// processSortTypeSelection обрабатывает выбор типа сортировки
func (h *AdminHandler) processSortTypeSelection(c telebot.Context, actionType string) error {
	// Get selected sort type from message
	selectedText := c.Text()

	// Check for return to main menu
	if h.getButtonCommand(selectedText) == commands.ReturnToMainMenu {
		return h.handleStart(c)
	}

	// Find matching sort type
	var selectedSortType models.SortType
	sortTypes := []models.SortType{
		models.SortByCreationOrder,
		models.SortByExpiryDate,
		models.SortByTrafficTotal,
		models.SortByStatus,
		models.SortByName,
	}

	found := false
	for _, sortType := range sortTypes {
		if sortType.GetSortName() == selectedText {
			selectedSortType = sortType
			found = true
			break
		}
	}

	if !found {
		return h.sendTextMessage(c, "❌ <b>Invalid Selection</b>\n\nPlease select one of the sorting options from the menu.", h.createReturnKeyboard())
	}

	// Save sort type
	err := h.stateService.WithSortType(c.Sender().ID, selectedSortType)
	if err != nil {
		h.logger.Errorf("Failed to set sort type: %v", err)
		return err
	}

	// Show members list with selected sorting
	return h.showMembersWithSort(c, selectedSortType, actionType)
}

// showMembersWithSort показывает список пользователей с указанной сортировкой
func (h *AdminHandler) showMembersWithSort(c telebot.Context, sortType models.SortType, actionType string) error {
	// Get all members with detailed info
	members, err := h.xrayService.GetAllMembersWithInfo(context.Background(), sortType)
	if err != nil {
		h.logger.Errorf("Failed to get members with info: %v", err)
		return h.sendTextMessage(c, "❌ <b>Connection Error</b>\n\nCouldn't retrieve user list. Please check your server connection and try again.", h.createReturnKeyboard())
	}

	if len(members) == 0 {
		message := "📭 <b>No Users Found</b>\n\nThere are no users in the system yet."
		if actionType == "edit" {
			message += " Use <b>Add Member</b> to create your first user."
		}
		return h.sendTextMessage(c, message, h.createReturnKeyboard())
	}

	// Create keyboard with member names and additional info
	markup := &telebot.ReplyMarkup{
		ResizeKeyboard: true,
	}

	var rows []telebot.Row
	for _, member := range members {
		// Format button text with additional info based on sort type
		buttonText := h.formatMemberButtonText(member, sortType)
		rows = append(rows, telebot.Row{telebot.Btn{Text: buttonText}})
	}

	// Add return button
	rows = append(rows, telebot.Row{telebot.Btn{Text: "↩️ " + commands.ReturnToMainMenu}})

	markup.Reply(rows...)

	// Set appropriate state
	var nextState models.ConversationState
	var messageText string

	if actionType == "edit" {
		nextState = models.AwaitSelectUserName
		messageText = "✏️ <b>Edit User</b>\n\n👥 Select a user to manage:"
	} else if actionType == "delete" {
		nextState = models.AwaitConfirmMemberDeletion
		messageText = "🗑️ <b>Delete User</b>\n\n⚠️ Select a user to permanently delete:"
	}

	err = h.stateService.WithConversationState(c.Sender().ID, nextState)
	if err != nil {
		h.logger.Errorf("Failed to set state: %v", err)
		return err
	}

	return h.sendTextMessage(c, messageText, markup)
}

// formatMemberButtonText форматирует текст кнопки пользователя с дополнительной информацией
func (h *AdminHandler) formatMemberButtonText(member models.MemberInfo, sortType models.SortType) string {
	baseText := member.BaseUsername

	switch sortType {
	case models.SortByCreationOrder:
		return baseText // По дате добавления показываем только имя
	case models.SortByExpiryDate:
		return fmt.Sprintf("%s (%s)", baseText, member.GetExpiryStatus())
	case models.SortByTrafficTotal:
		if member.TotalTraffic > 0 {
			totalGB := float64(member.TotalTraffic) / (1024 * 1024 * 1024)
			return fmt.Sprintf("%s (%.1f GB)", baseText, totalGB)
		}
		return fmt.Sprintf("%s (0 GB)", baseText)
	case models.SortByStatus:
		status := "❌"
		if member.Enable {
			status = "✅"
		}
		return fmt.Sprintf("%s %s", status, baseText)
	default:
		return baseText
	}
}
