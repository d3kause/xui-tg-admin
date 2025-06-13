package handlers

import (
	"context"
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
		return h.processConfirmReset(c)
	case models.AwaitExtendDuration:
		return h.processExtendDuration(c)
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

// handleDefaultState handles the default state
func (h *AdminHandler) handleDefaultState(c telebot.Context) error {
	// Check if we have a command handler for this text
	if handler, ok := h.commandHandlers[c.Text()]; ok {
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

	// Show main menu
	markup := h.createMainKeyboard(permissions.Admin)
	return h.sendTextMessage(c, "Welcome to X-UI Admin Bot!", markup)
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
	return h.sendTextMessage(c, "Please enter the username for the new member:", markup)
}

// handleEditMember handles the Edit Member command
func (h *AdminHandler) handleEditMember(c telebot.Context) error {

	// Проверяем доступность сервиса
	_, err := h.stateService.GetState(c.Sender().ID)
	if err != nil {
		h.logger.Errorf("Failed to get user state: %v", err)
		return err
	}

	// Get all members
	members, err := h.xrayService.GetAllMembers(context.Background())
	if err != nil {
		h.logger.Errorf("Failed to get members: %v", err)
		return h.sendTextMessage(c, fmt.Sprintf("Failed to get members: %v", err), nil)
	}

	if len(members) == 0 {
		return h.sendTextMessage(c, "No members found.", h.createReturnKeyboard())
	}

	// Create keyboard with member names
	markup := &telebot.ReplyMarkup{
		ResizeKeyboard: true,
	}

	var rows []telebot.Row
	for _, name := range members {
		rows = append(rows, telebot.Row{telebot.Btn{Text: name}})
	}

	// Add return button
	rows = append(rows, telebot.Row{telebot.Btn{Text: commands.ReturnToMainMenu}})

	markup.Reply(rows...)

	// Set state to awaiting user selection
	err = h.stateService.WithConversationState(c.Sender().ID, models.AwaitSelectUserName)
	if err != nil {
		h.logger.Errorf("Failed to set state: %v", err)
		return err
	}

	return h.sendTextMessage(c, "Please select a member to edit:", markup)
}

// handleDeleteMember handles the Delete Member command
func (h *AdminHandler) handleDeleteMember(c telebot.Context) error {

	// Проверяем доступность сервиса
	_, err := h.stateService.GetState(c.Sender().ID)
	if err != nil {
		h.logger.Errorf("Failed to get user state: %v", err)
		return err
	}

	// Get all members
	members, err := h.xrayService.GetAllMembers(context.Background())
	if err != nil {
		h.logger.Errorf("Failed to get members: %v", err)
		return h.sendTextMessage(c, fmt.Sprintf("Failed to get members: %v", err), nil)
	}

	if len(members) == 0 {
		return h.sendTextMessage(c, "No members found.", h.createReturnKeyboard())
	}

	// Create keyboard with member names
	markup := &telebot.ReplyMarkup{
		ResizeKeyboard: true,
	}

	var rows []telebot.Row
	for _, name := range members {
		rows = append(rows, telebot.Row{telebot.Btn{Text: name}})
	}

	// Add return button
	rows = append(rows, telebot.Row{telebot.Btn{Text: commands.ReturnToMainMenu}})

	markup.Reply(rows...)

	// Set state to awaiting user selection for deletion
	err = h.stateService.WithConversationState(c.Sender().ID, models.AwaitConfirmMemberDeletion)
	if err != nil {
		h.logger.Errorf("Failed to set state: %v", err)
		return err
	}

	return h.sendTextMessage(c, "Please select a member to delete:", markup)
}

// handleGetOnlineMembers handles the Online Members command
func (h *AdminHandler) handleGetOnlineMembers(c telebot.Context) error {

	// Get online users
	onlineUsers, err := h.xrayService.GetOnlineUsers(context.Background())
	if err != nil {
		h.logger.Errorf("Failed to get online users: %v", err)
		return h.sendTextMessage(c, fmt.Sprintf("Failed to get online users: %v", err), nil)
	}

	// Format message
	var message string
	if len(onlineUsers) == 0 {
		message = "No users are currently online."
	} else {
		message = "Online users:\n\n"
		for _, user := range onlineUsers {
			message += fmt.Sprintf("・%s\n", user)
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
		return h.sendTextMessage(c, fmt.Sprintf("Failed to get inbounds: %v", err), nil)
	}

	// Format beautiful network usage report
	message := helpers.FormatNetworkUsageReport(inbounds)

	return h.sendTextMessage(c, message, h.createReturnKeyboard())
}

// handleResetUsersNetworkUsage handles the Reset Network Usage command
func (h *AdminHandler) handleResetUsersNetworkUsage(c telebot.Context) error {

	// Проверяем доступность сервиса
	_, err := h.stateService.GetState(c.Sender().ID)
	if err != nil {
		h.logger.Errorf("Failed to get user state: %v", err)
		return err
	}

	// Get all members
	members, err := h.xrayService.GetAllMembers(context.Background())
	if err != nil {
		h.logger.Errorf("Failed to get members: %v", err)
		return h.sendTextMessage(c, fmt.Sprintf("Failed to get members: %v", err), nil)
	}

	if len(members) == 0 {
		return h.sendTextMessage(c, "No members found.", h.createReturnKeyboard())
	}

	// Create keyboard with member names
	markup := &telebot.ReplyMarkup{
		ResizeKeyboard: true,
	}

	var rows []telebot.Row
	for _, name := range members {
		rows = append(rows, telebot.Row{telebot.Btn{Text: name}})
	}

	// Add return button
	rows = append(rows, telebot.Row{telebot.Btn{Text: commands.ReturnToMainMenu}})

	markup.Reply(rows...)

	// Set state to awaiting user selection for reset
	err = h.stateService.WithConversationState(c.Sender().ID, models.AwaitConfirmResetUsersNetworkUsage)
	if err != nil {
		h.logger.Errorf("Failed to set state: %v", err)
		return err
	}

	return h.sendTextMessage(c, "Please select a member to reset network usage:", markup)
}

// processUserName processes the username input
func (h *AdminHandler) processUserName(c telebot.Context) error {
	// Get username from message
	username := c.Text()

	// Validate username
	if username == commands.ReturnToMainMenu {
		return h.handleStart(c)
	}

	// Validate username format
	if err := validation.ValidateUsername(username); err != nil {
		return h.sendTextMessage(c, fmt.Sprintf("%s. Please try again:", err.Error()), nil)
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
			telebot.Btn{Text: commands.Infinite},
		},
		telebot.Row{
			telebot.Btn{Text: commands.ReturnToMainMenu},
		},
	)

	return h.sendTextMessage(c, "Enter the duration in days (e.g., 30) or choose infinite duration:", markup)
}

// processDuration processes the duration input
func (h *AdminHandler) processDuration(c telebot.Context) error {
	// Get duration from message
	durationStr := c.Text()

	// Validate duration
	if durationStr == commands.ReturnToMainMenu {
		return h.handleStart(c)
	}

	// Get user state
	userState, err := h.stateService.GetState(c.Sender().ID)
	if err != nil {
		h.logger.Errorf("Failed to get user state: %v", err)
		return err
	}

	// Get username from state
	if userState.Payload == nil {
		return h.sendTextMessage(c, "Username not found. Please try again.", nil)
	}

	baseUsername := *userState.Payload

	// Get enabled inbounds
	enabledInbounds, err := h.getEnabledInbounds(context.Background())
	if err != nil {
		h.logger.Errorf("Failed to get enabled inbounds: %v", err)
		return h.sendTextMessage(c, fmt.Sprintf("Failed to get inbounds: %v. Please contact administrator.", err), nil)
	}

	// Calculate expiry time
	expiryTime, err := calculateExpiryTime(durationStr)
	if err != nil {
		return h.sendTextMessage(c, fmt.Sprintf("%s. Please try again:", err.Error()), nil)
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

	// Create clients for all enabled inbounds
	createdEmails, addErrors, addedToAny := h.createClientsForAllInbounds(context.Background(), params, enabledInbounds)

	if !addedToAny {
		return h.sendTextMessage(c, fmt.Sprintf("Failed to add client to any inbound:\n%s", strings.Join(addErrors, "\n")), nil)
	}

	// Send subscription information and QR code
	return h.sendSubscriptionInfo(c, params, createdEmails, addErrors)
}

// processSelectUser processes the user selection
func (h *AdminHandler) processSelectUser(c telebot.Context) error {
	// Get username from message
	username := c.Text()

	// Validate username
	if username == commands.ReturnToMainMenu {
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
	markup := &telebot.ReplyMarkup{
		ResizeKeyboard: true,
	}

	markup.Reply(
		telebot.Row{
			telebot.Btn{Text: commands.ViewConfig},
			telebot.Btn{Text: commands.ExtendDuration},
		},
		telebot.Row{
			telebot.Btn{Text: commands.ResetTraffic},
			telebot.Btn{Text: commands.Delete},
		},
		telebot.Row{
			telebot.Btn{Text: commands.ReturnToMainMenu},
		},
	)

	return h.sendTextMessage(c, fmt.Sprintf("Selected user: %s\nWhat would you like to do?", username), markup)
}

// processMemberAction processes the member action selection
func (h *AdminHandler) processMemberAction(c telebot.Context) error {
	// Get action from message
	action := c.Text()

	// Проверяем доступность сервиса
	userState, err := h.stateService.GetState(c.Sender().ID)
	if err != nil {
		h.logger.Errorf("Failed to get user state: %v", err)
		return err
	}

	// Get username from state
	if userState.Payload == nil {
		return h.sendTextMessage(c, "Username not found. Please try again.", nil)
	}

	username := *userState.Payload

	// Handle action
	switch action {
	case commands.ViewConfig:
		return h.handleViewConfig(c, username)
	case commands.ExtendDuration:
		return h.handleExtendDuration(c, username)
	case commands.ResetTraffic:
		return h.handleResetTraffic(c, username)
	case commands.Delete:
		return h.handleConfirmDelete(c, username)
	case commands.ReturnToMainMenu:
		return h.handleStart(c)
	default:
		return h.sendTextMessage(c, "Invalid action. Please try again.", nil)
	}
}

// handleViewConfig handles the View Config action
func (h *AdminHandler) handleViewConfig(c telebot.Context, username string) error {
	// Get subscription URL
	subURL, err := h.xrayService.GetSubscriptionURL(context.Background(), username)
	if err != nil {
		h.logger.Errorf("Failed to get subscription URL: %v", err)
		return h.sendTextMessage(c, fmt.Sprintf("Failed to get subscription URL: %v", err), nil)
	}

	// Send subscription URL
	err = h.sendTextMessage(c, fmt.Sprintf("Subscription URL for %s:\n\n%s", username, subURL), h.createReturnKeyboard())
	if err != nil {
		return err
	}

	// Send QR code
	return h.sendQRCode(c, subURL)
}

// handleExtendDuration handles the Extend Duration action
func (h *AdminHandler) handleExtendDuration(c telebot.Context, username string) error {
	// Set state to awaiting extend duration
	err := h.stateService.WithConversationState(c.Sender().ID, models.AwaitExtendDuration)
	if err != nil {
		h.logger.Errorf("Failed to set state: %v", err)
		return err
	}

	// Show return keyboard
	markup := h.createReturnKeyboard()
	return h.sendTextMessage(c, fmt.Sprintf("Please enter the number of days to extend for %s:", username), markup)
}

// processExtendDuration processes the extend duration input
func (h *AdminHandler) processExtendDuration(c telebot.Context) error {
	// Get duration from message
	durationStr := c.Text()

	// Validate duration
	if durationStr == commands.ReturnToMainMenu {
		return h.handleStart(c)
	}

	// Parse duration
	days, err := validation.ValidateDuration(durationStr)
	if err != nil {
		return h.sendTextMessage(c, fmt.Sprintf("%s. Please try again:", err.Error()), nil)
	}

	// Get user state
	userState, err := h.stateService.GetState(c.Sender().ID)
	if err != nil {
		h.logger.Errorf("Failed to get user state: %v", err)
		return err
	}

	// Get username from state
	if userState.Payload == nil {
		return h.sendTextMessage(c, "Username not found. Please try again.", nil)
	}

	username := *userState.Payload
	return h.extendClientDuration(context.Background(), c, username, days)
}

// handleResetTraffic handles the Reset Traffic action
func (h *AdminHandler) handleResetTraffic(c telebot.Context, username string) error {
	return h.resetClientTraffic(context.Background(), c, username)
}

// handleConfirmDelete handles the Delete action
func (h *AdminHandler) handleConfirmDelete(c telebot.Context, username string) error {
	// Show confirm keyboard
	markup := h.createConfirmKeyboard()
	return h.sendTextMessage(c, fmt.Sprintf("Are you sure you want to delete %s?", username), markup)
}

// processConfirmDeletion processes the deletion confirmation
func (h *AdminHandler) processConfirmDeletion(c telebot.Context) error {
	// Get confirmation from message
	confirmation := c.Text()

	// If user wants to cancel, return to main menu
	if confirmation == commands.ReturnToMainMenu {
		return h.handleStart(c)
	}

	// Check if user confirmed
	if confirmation != commands.Confirm {
		return h.sendTextMessage(c, "Invalid action. Please confirm deletion or return to main menu.", nil)
	}

	// Get user state to get the username we want to delete
	userState, err := h.stateService.GetState(c.Sender().ID)
	if err != nil {
		h.logger.Errorf("Failed to get user state: %v", err)
		return err
	}

	if userState.Payload == nil {
		return h.sendTextMessage(c, "Error: No user selected for deletion", h.createReturnKeyboard())
	}

	username := *userState.Payload

	// Delete client using email
	err = h.xrayService.RemoveClients(context.Background(), []string{username})
	if err != nil {
		h.logger.Errorf("Failed to delete client: %v", err)
		return h.sendTextMessage(c, fmt.Sprintf("Failed to delete client: %v", err), nil)
	}

	return h.sendTextMessage(c, fmt.Sprintf("Client %s deleted successfully.", username), h.createReturnKeyboard())
}

// processConfirmReset processes the reset confirmation
func (h *AdminHandler) processConfirmReset(c telebot.Context) error {
	username := c.Text()
	if username == commands.ReturnToMainMenu {
		return h.handleStart(c)
	}
	return h.resetClientTraffic(context.Background(), c, username)
}

// handleGetDetailedUsersInfo handles the Detailed Usage command
func (h *AdminHandler) handleGetDetailedUsersInfo(c telebot.Context) error {

	// Get inbounds
	inbounds, err := h.xrayService.GetInbounds(context.Background())
	if err != nil {
		h.logger.Errorf("Failed to get inbounds: %v", err)
		return h.sendTextMessage(c, fmt.Sprintf("Failed to get inbounds: %v", err), nil)
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
			telebot.Btn{Text: commands.Confirm},
		},
		telebot.Row{
			telebot.Btn{Text: commands.ReturnToMainMenu},
		},
	)

	return markup
}
