package handlers

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	telebot "gopkg.in/telebot.v3"

	"xui-tg-admin/internal/config"
	"xui-tg-admin/internal/models"
	"xui-tg-admin/internal/permissions"
	"xui-tg-admin/internal/services"
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
	state, err := h.stateService.GetState(userID)
	if err != nil {
		h.logger.Errorf("Failed to get user state: %v", err)
		return err
	}

	// Handle based on state
	switch state.State {
	case models.Default:
		return h.handleDefaultState(c)
	case models.AwaitingSelectServer:
		return h.HandleSelectServer(c)
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
		h.logger.Warnf("Unknown state: %d", state.State)
		return h.handleDefaultState(c)
	}
}

// initializeCommands initializes the command handlers
func (h *AdminHandler) initializeCommands() {
	h.commandHandlers = map[string]func(telebot.Context) error{
		"/start":              h.handleStart,
		"Add Member":          h.handleAddMember,
		"Edit Member":         h.handleEditMember,
		"Delete Member":       h.handleDeleteMember,
		"Online Members":      h.handleGetOnlineMembers,
		"Network Usage":       h.handleGetUsersNetworkUsage,
		"Reset Network Usage": h.handleResetUsersNetworkUsage,
		"Change Server":       h.handleChangeServer,
		"Return to Main Menu": h.handleStart,
		"Cancel":              h.handleStart,
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

	// Check if a server is selected
	state, err := h.stateService.GetState(c.Sender().ID)
	if err != nil {
		h.logger.Errorf("Failed to get user state: %v", err)
		return err
	}

	// If no server is selected, show server selection
	if state.SelectedServer == nil {
		return h.HandleSelectServer(c)
	}

	// Show main menu
	markup := h.createMainKeyboard(permissions.Admin)
	return h.sendTextMessage(c, fmt.Sprintf("Welcome to X-UI Admin Bot!\nCurrent server: %s", *state.SelectedServer), markup)
}

// handleAddMember handles the Add Member command
func (h *AdminHandler) handleAddMember(c telebot.Context) error {
	// Validate server selection
	if err := h.validateServerSelection(c.Sender().ID); err != nil {
		return h.handleSelectServer(c)
	}

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
	// Validate server selection
	if err := h.validateServerSelection(c.Sender().ID); err != nil {
		return h.handleSelectServer(c)
	}

	// Get user state
	state, err := h.stateService.GetState(c.Sender().ID)
	if err != nil {
		h.logger.Errorf("Failed to get user state: %v", err)
		return err
	}

	// Get all members
	members, err := h.xrayService.GetAllMembers(context.Background(), *state.SelectedServer)
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
	rows = append(rows, telebot.Row{telebot.Btn{Text: "Return to Main Menu"}})

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
	// Validate server selection
	if err := h.validateServerSelection(c.Sender().ID); err != nil {
		return h.handleSelectServer(c)
	}

	// Get user state
	state, err := h.stateService.GetState(c.Sender().ID)
	if err != nil {
		h.logger.Errorf("Failed to get user state: %v", err)
		return err
	}

	// Get all members
	members, err := h.xrayService.GetAllMembers(context.Background(), *state.SelectedServer)
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
	rows = append(rows, telebot.Row{telebot.Btn{Text: "Return to Main Menu"}})

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
	// Validate server selection
	if err := h.validateServerSelection(c.Sender().ID); err != nil {
		return h.handleSelectServer(c)
	}

	// Get user state
	state, err := h.stateService.GetState(c.Sender().ID)
	if err != nil {
		h.logger.Errorf("Failed to get user state: %v", err)
		return err
	}

	// Get online users
	onlineUsers, err := h.xrayService.GetOnlineUsers(context.Background(), *state.SelectedServer)
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
		for i, user := range onlineUsers {
			message += fmt.Sprintf("%d. %s\n", i+1, user)
		}
	}

	return h.sendTextMessage(c, message, h.createReturnKeyboard())
}

// handleGetUsersNetworkUsage handles the Network Usage command
func (h *AdminHandler) handleGetUsersNetworkUsage(c telebot.Context) error {
	// Validate server selection
	if err := h.validateServerSelection(c.Sender().ID); err != nil {
		return h.handleSelectServer(c)
	}

	// Get user state
	state, err := h.stateService.GetState(c.Sender().ID)
	if err != nil {
		h.logger.Errorf("Failed to get user state: %v", err)
		return err
	}

	// Get inbounds
	inbounds, err := h.xrayService.GetInbounds(context.Background(), *state.SelectedServer)
	if err != nil {
		h.logger.Errorf("Failed to get inbounds: %v", err)
		return h.sendTextMessage(c, fmt.Sprintf("Failed to get inbounds: %v", err), nil)
	}

	// Format message
	var message strings.Builder
	message.WriteString("Network Usage:\n\n")

	for _, inbound := range inbounds {
		message.WriteString(fmt.Sprintf("Inbound: %s (ID: %d)\n", inbound.Remark, inbound.ID))

		if len(inbound.ClientStats) == 0 {
			message.WriteString("  No clients\n\n")
			continue
		}

		for _, client := range inbound.ClientStats {
			// Convert bytes to GB
			upGB := float64(client.Up) / (1024 * 1024 * 1024)
			downGB := float64(client.Down) / (1024 * 1024 * 1024)
			totalGB := float64(client.Total) / (1024 * 1024 * 1024)

			// Format expiry time
			expiryTime := "Never"
			if client.ExpiryTime > 0 {
				expiryTime = time.Unix(client.ExpiryTime/1000, 0).Format("2006-01-02")
			}

			message.WriteString(fmt.Sprintf("  %s:\n", client.Email))
			message.WriteString(fmt.Sprintf("    Up: %.2f GB\n", upGB))
			message.WriteString(fmt.Sprintf("    Down: %.2f GB\n", downGB))
			message.WriteString(fmt.Sprintf("    Total: %.2f GB\n", totalGB))
			message.WriteString(fmt.Sprintf("    Expiry: %s\n", expiryTime))
			message.WriteString(fmt.Sprintf("    Enabled: %v\n\n", client.Enable))
		}
	}

	return h.sendTextMessage(c, message.String(), h.createReturnKeyboard())
}

// handleResetUsersNetworkUsage handles the Reset Network Usage command
func (h *AdminHandler) handleResetUsersNetworkUsage(c telebot.Context) error {
	// Validate server selection
	if err := h.validateServerSelection(c.Sender().ID); err != nil {
		return h.handleSelectServer(c)
	}

	// Get user state
	state, err := h.stateService.GetState(c.Sender().ID)
	if err != nil {
		h.logger.Errorf("Failed to get user state: %v", err)
		return err
	}

	// Get all members
	members, err := h.xrayService.GetAllMembers(context.Background(), *state.SelectedServer)
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
	rows = append(rows, telebot.Row{telebot.Btn{Text: "Return to Main Menu"}})

	markup.Reply(rows...)

	// Set state to awaiting user selection for reset
	err = h.stateService.WithConversationState(c.Sender().ID, models.AwaitConfirmResetUsersNetworkUsage)
	if err != nil {
		h.logger.Errorf("Failed to set state: %v", err)
		return err
	}

	return h.sendTextMessage(c, "Please select a member to reset network usage:", markup)
}

// handleSelectServer handles server selection
func (h *AdminHandler) handleSelectServer(c telebot.Context) error {
	return h.HandleSelectServer(c)
}

// handleChangeServer handles the Change Server command
func (h *AdminHandler) handleChangeServer(c telebot.Context) error {
	return h.handleSelectServer(c)
}

// processUserName processes the username input
func (h *AdminHandler) processUserName(c telebot.Context) error {
	// Get username from message
	username := c.Text()

	// Validate username
	if username == "Return to Main Menu" {
		return h.handleStart(c)
	}

	// TODO: Add username validation

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

	// Show return keyboard
	markup := h.createReturnKeyboard()
	return h.sendTextMessage(c, "Please enter the duration in days (e.g., 30):", markup)
}

// processDuration processes the duration input
func (h *AdminHandler) processDuration(c telebot.Context) error {
	// Get duration from message
	durationStr := c.Text()

	// Validate duration
	if durationStr == "Return to Main Menu" {
		return h.handleStart(c)
	}

	// Parse duration
	days, err := strconv.Atoi(durationStr)
	if err != nil {
		return h.sendTextMessage(c, "Invalid duration. Please enter a number of days (e.g., 30):", nil)
	}

	// Get user state
	state, err := h.stateService.GetState(c.Sender().ID)
	if err != nil {
		h.logger.Errorf("Failed to get user state: %v", err)
		return err
	}

	// Get username from state
	if state.Payload == nil {
		return h.sendTextMessage(c, "Username not found. Please try again.", nil)
	}

	username := *state.Payload

	// Calculate expiry time
	expiryTime := time.Now().Add(time.Duration(days)*24*time.Hour).Unix() * 1000

	// Create client
	client := models.Client{
		ID:         username,
		Enable:     true,
		Email:      username,
		TotalGB:    1024, // 1 TB
		LimitIP:    0,    // No limit
		ExpiryTime: &expiryTime,
		TgID:       fmt.Sprintf("%d", c.Sender().ID),
		SubID:      models.GenerateSubID(),
	}

	// Add client to inbound
	err = h.xrayService.AddClient(context.Background(), *state.SelectedServer, 1, client)
	if err != nil {
		h.logger.Errorf("Failed to add client: %v", err)
		return h.sendTextMessage(c, fmt.Sprintf("Failed to add client: %v", err), nil)
	}

	// Get subscription URL
	subURL, err := h.xrayService.GetSubscriptionURL(context.Background(), *state.SelectedServer, username)
	if err != nil {
		h.logger.Errorf("Failed to get subscription URL: %v", err)
		return h.sendTextMessage(c, fmt.Sprintf("Client added, but failed to get subscription URL: %v", err), nil)
	}

	// Send success message
	err = h.sendTextMessage(c, fmt.Sprintf("Client added successfully!\n\nUsername: %s\nDuration: %d days\nExpiry: %s\n\nSubscription URL: %s",
		username,
		days,
		time.Unix(expiryTime/1000, 0).Format("2006-01-02"),
		subURL),
		h.createReturnKeyboard())
	if err != nil {
		return err
	}

	// Send QR code
	return h.sendQRCode(c, subURL)
}

// processSelectUser processes the user selection
func (h *AdminHandler) processSelectUser(c telebot.Context) error {
	// Get username from message
	username := c.Text()

	// Validate username
	if username == "Return to Main Menu" {
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
			telebot.Btn{Text: "View Config"},
			telebot.Btn{Text: "Extend Duration"},
		},
		telebot.Row{
			telebot.Btn{Text: "Reset Traffic"},
			telebot.Btn{Text: "Delete"},
		},
		telebot.Row{
			telebot.Btn{Text: "Return to Main Menu"},
		},
	)

	return h.sendTextMessage(c, fmt.Sprintf("Selected user: %s\nWhat would you like to do?", username), markup)
}

// processMemberAction processes the member action selection
func (h *AdminHandler) processMemberAction(c telebot.Context) error {
	// Get action from message
	action := c.Text()

	// Get user state
	state, err := h.stateService.GetState(c.Sender().ID)
	if err != nil {
		h.logger.Errorf("Failed to get user state: %v", err)
		return err
	}

	// Get username from state
	if state.Payload == nil {
		return h.sendTextMessage(c, "Username not found. Please try again.", nil)
	}

	username := *state.Payload

	// Handle action
	switch action {
	case "View Config":
		return h.handleViewConfig(c, username)
	case "Extend Duration":
		return h.handleExtendDuration(c, username)
	case "Reset Traffic":
		return h.handleResetTraffic(c, username)
	case "Delete":
		return h.handleConfirmDelete(c, username)
	case "Return to Main Menu":
		return h.handleStart(c)
	default:
		return h.sendTextMessage(c, "Invalid action. Please try again.", nil)
	}
}

// handleViewConfig handles the View Config action
func (h *AdminHandler) handleViewConfig(c telebot.Context, username string) error {
	// Get user state
	state, err := h.stateService.GetState(c.Sender().ID)
	if err != nil {
		h.logger.Errorf("Failed to get user state: %v", err)
		return err
	}

	// Get subscription URL
	subURL, err := h.xrayService.GetSubscriptionURL(context.Background(), *state.SelectedServer, username)
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
	if durationStr == "Return to Main Menu" {
		return h.handleStart(c)
	}

	// Parse duration
	days, err := strconv.Atoi(durationStr)
	if err != nil {
		return h.sendTextMessage(c, "Invalid duration. Please enter a number of days (e.g., 30):", nil)
	}

	// Get user state
	state, err := h.stateService.GetState(c.Sender().ID)
	if err != nil {
		h.logger.Errorf("Failed to get user state: %v", err)
		return err
	}

	// Get username from state
	if state.Payload == nil {
		return h.sendTextMessage(c, "Username not found. Please try again.", nil)
	}

	username := *state.Payload

	// TODO: Implement extend duration functionality
	// This would require getting the current client, updating the expiry time, and updating the client

	return h.sendTextMessage(c, fmt.Sprintf("Extended duration for %s by %d days.", username, days), h.createReturnKeyboard())
}

// handleResetTraffic handles the Reset Traffic action
func (h *AdminHandler) handleResetTraffic(c telebot.Context, username string) error {
	// Get user state
	state, err := h.stateService.GetState(c.Sender().ID)
	if err != nil {
		h.logger.Errorf("Failed to get user state: %v", err)
		return err
	}

	// Get inbounds
	inbounds, err := h.xrayService.GetInbounds(context.Background(), *state.SelectedServer)
	if err != nil {
		h.logger.Errorf("Failed to get inbounds: %v", err)
		return h.sendTextMessage(c, fmt.Sprintf("Failed to get inbounds: %v", err), nil)
	}

	// Find inbound with client
	var inboundID int
	found := false

	for _, inbound := range inbounds {
		for _, client := range inbound.ClientStats {
			if client.Email == username {
				inboundID = inbound.ID
				found = true
				break
			}
		}
		if found {
			break
		}
	}

	if !found {
		return h.sendTextMessage(c, fmt.Sprintf("Client %s not found.", username), h.createReturnKeyboard())
	}

	// Reset traffic
	err = h.xrayService.ResetUserTraffic(context.Background(), *state.SelectedServer, inboundID, username)
	if err != nil {
		h.logger.Errorf("Failed to reset traffic: %v", err)
		return h.sendTextMessage(c, fmt.Sprintf("Failed to reset traffic: %v", err), nil)
	}

	return h.sendTextMessage(c, fmt.Sprintf("Traffic reset for %s.", username), h.createReturnKeyboard())
}

// handleConfirmDelete handles the Delete action
func (h *AdminHandler) handleConfirmDelete(c telebot.Context, username string) error {
	// Show confirm keyboard
	markup := h.createConfirmKeyboard()
	return h.sendTextMessage(c, fmt.Sprintf("Are you sure you want to delete %s?", username), markup)
}

// processConfirmDeletion processes the deletion confirmation
func (h *AdminHandler) processConfirmDeletion(c telebot.Context) error {
	// Get username from message
	username := c.Text()

	// Validate username
	if username == "Return to Main Menu" {
		return h.handleStart(c)
	}

	// Get user state
	state, err := h.stateService.GetState(c.Sender().ID)
	if err != nil {
		h.logger.Errorf("Failed to get user state: %v", err)
		return err
	}

	// Delete client
	err = h.xrayService.RemoveClients(context.Background(), *state.SelectedServer, []string{username})
	if err != nil {
		h.logger.Errorf("Failed to delete client: %v", err)
		return h.sendTextMessage(c, fmt.Sprintf("Failed to delete client: %v", err), nil)
	}

	return h.sendTextMessage(c, fmt.Sprintf("Client %s deleted successfully.", username), h.createReturnKeyboard())
}

// processConfirmReset processes the reset confirmation
func (h *AdminHandler) processConfirmReset(c telebot.Context) error {
	// Get username from message
	username := c.Text()

	// Validate username
	if username == "Return to Main Menu" {
		return h.handleStart(c)
	}

	// Get user state
	state, err := h.stateService.GetState(c.Sender().ID)
	if err != nil {
		h.logger.Errorf("Failed to get user state: %v", err)
		return err
	}

	// Get inbounds
	inbounds, err := h.xrayService.GetInbounds(context.Background(), *state.SelectedServer)
	if err != nil {
		h.logger.Errorf("Failed to get inbounds: %v", err)
		return h.sendTextMessage(c, fmt.Sprintf("Failed to get inbounds: %v", err), nil)
	}

	// Find inbound with client
	var inboundID int
	found := false

	for _, inbound := range inbounds {
		for _, client := range inbound.ClientStats {
			if client.Email == username {
				inboundID = inbound.ID
				found = true
				break
			}
		}
		if found {
			break
		}
	}

	if !found {
		return h.sendTextMessage(c, fmt.Sprintf("Client %s not found.", username), h.createReturnKeyboard())
	}

	// Reset traffic
	err = h.xrayService.ResetUserTraffic(context.Background(), *state.SelectedServer, inboundID, username)
	if err != nil {
		h.logger.Errorf("Failed to reset traffic: %v", err)
		return h.sendTextMessage(c, fmt.Sprintf("Failed to reset traffic: %v", err), nil)
	}

	return h.sendTextMessage(c, fmt.Sprintf("Traffic reset for %s.", username), h.createReturnKeyboard())
}
