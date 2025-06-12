package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	telebot "gopkg.in/telebot.v3"

	"xui-tg-admin/internal/commands"
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
	message := h.formatNetworkUsageReport(inbounds)

	return h.sendTextMessage(c, message, h.createReturnKeyboard())
}

// formatNetworkUsageReport formats a beautiful network usage report
func (h *AdminHandler) formatNetworkUsageReport(inbounds []models.Inbound) string {
	var sb strings.Builder
	sb.WriteString("<b>Network Usage Report:</b>\n")
	sb.WriteString("<pre>\n")
	sb.WriteString("Email             | ↓ (GB) | ↑ (GB)\n")
	sb.WriteString("------------------|--------|--------\n")

	var totalUploadGB int64 = 0
	var totalDownloadGB int64 = 0

	for _, inbound := range inbounds {
		if len(inbound.ClientStats) == 0 {
			continue
		}

		sb.WriteString("\n")
		sb.WriteString(fmt.Sprintf("Inbound: %s\n", inbound.Remark))

		inboundDownloadTotal, inboundUploadTotal := h.calculateInboundTraffic(inbound.ClientStats)
		totalDownloadGB += inboundDownloadTotal
		totalUploadGB += inboundUploadTotal

		for _, client := range inbound.ClientStats {
			sb.WriteString(h.formatTableLine(client.Email, client.Down, client.Up))
		}

		sb.WriteString("-----------\n")
		sb.WriteString(h.formatTableLine("Total:", inboundDownloadTotal*1024*1024*1024, inboundUploadTotal*1024*1024*1024))
	}

	sb.WriteString("\n")
	sb.WriteString(h.formatTableLine("Grand Total:", totalDownloadGB*1024*1024*1024, totalUploadGB*1024*1024*1024))
	sb.WriteString("</pre>")

	return sb.String()
}

// calculateInboundTraffic calculates total traffic for an inbound (in GB)
func (h *AdminHandler) calculateInboundTraffic(clientStats []models.ClientStat) (downloadGB int64, uploadGB int64) {
	for _, client := range clientStats {
		downloadGB += client.Down / (1024 * 1024 * 1024)
		uploadGB += client.Up / (1024 * 1024 * 1024)
	}
	return
}

// formatTableLine formats a single line of the traffic table
func (h *AdminHandler) formatTableLine(email string, downBytes int64, upBytes int64) string {
	downGB := float64(downBytes) / (1024 * 1024 * 1024)
	upGB := float64(upBytes) / (1024 * 1024 * 1024)

	// Truncate email if too long for table formatting
	displayEmail := email
	if len(email) > 17 {
		displayEmail = email[:14] + "..."
	}

	return fmt.Sprintf("%-17s | %6.2f | %6.2f\n", displayEmail, downGB, upGB)
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
	if len(username) < 3 || len(username) > 32 {
		return h.sendTextMessage(c, "Username must be between 3 and 32 characters. Please try again:", nil)
	}

	// Check if username contains only allowed characters (alphanumeric and underscore)
	for _, r := range username {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_') {
			return h.sendTextMessage(c, "Username can only contain letters, numbers, and underscores. Please try again:", nil)
		}
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

	// Show return keyboard
	markup := h.createReturnKeyboard()
	return h.sendTextMessage(c, "Please enter the duration in days (e.g., 30):", markup)
}

// processDuration processes the duration input
func (h *AdminHandler) processDuration(c telebot.Context) error {
	// Get duration from message
	durationStr := c.Text()

	// Validate duration
	if durationStr == commands.ReturnToMainMenu {
		return h.handleStart(c)
	}

	// Parse duration
	days, err := strconv.Atoi(durationStr)
	if err != nil {
		return h.sendTextMessage(c, "Invalid duration. Please enter a number of days (e.g., 30):", nil)
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
	err = h.xrayService.AddClient(context.Background(), 1, client)
	if err != nil {
		h.logger.Errorf("Failed to add client: %v", err)
		return h.sendTextMessage(c, fmt.Sprintf("Failed to add client: %v", err), nil)
	}

	// Get subscription URL
	subURL, err := h.xrayService.GetSubscriptionURL(context.Background(), username)
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
	days, err := strconv.Atoi(durationStr)
	if err != nil {
		return h.sendTextMessage(c, "Invalid duration. Please enter a number of days (e.g., 30):", nil)
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

	// Get inbounds to find the client
	inbounds, err := h.xrayService.GetInbounds(context.Background())
	if err != nil {
		h.logger.Errorf("Failed to get inbounds: %v", err)
		return h.sendTextMessage(c, fmt.Sprintf("Failed to get inbounds: %v", err), nil)
	}

	// Find the client
	var foundInbound *models.Inbound
	var foundClient *models.ClientStat
	for _, inbound := range inbounds {
		for _, client := range inbound.ClientStats {
			if client.Email == username {
				foundInbound = &inbound
				foundClient = &client
				break
			}
		}
		if foundInbound != nil {
			break
		}
	}

	if foundInbound == nil || foundClient == nil {
		return h.sendTextMessage(c, fmt.Sprintf("Client %s not found.", username), h.createReturnKeyboard())
	}

	// Calculate new expiry time (extend by the specified days)
	newExpiryTime := foundClient.ExpiryTime + (int64(days) * 24 * 60 * 60 * 1000) // Convert days to milliseconds

	// Create updated client
	updatedClient := models.Client{
		ID:         username,
		Enable:     foundClient.Enable,
		Email:      username,
		TotalGB:    int(foundClient.Total / (1024 * 1024 * 1024)), // Convert bytes to GB
		LimitIP:    0,                                             // Maintain original limit
		ExpiryTime: &newExpiryTime,
		TgID:       fmt.Sprintf("%d", c.Sender().ID),
		SubID:      models.GenerateSubID(),
	}

	// Remove the old client
	err = h.xrayService.RemoveClients(context.Background(), []string{username})
	if err != nil {
		h.logger.Errorf("Failed to remove old client: %v", err)
		return h.sendTextMessage(c, fmt.Sprintf("Failed to remove old client: %v", err), nil)
	}

	// Add the updated client
	err = h.xrayService.AddClient(context.Background(), foundInbound.ID, updatedClient)
	if err != nil {
		h.logger.Errorf("Failed to add updated client: %v", err)
		return h.sendTextMessage(c, fmt.Sprintf("Failed to add updated client: %v", err), nil)
	}

	return h.sendTextMessage(c, fmt.Sprintf("Successfully extended duration for %s by %d days.", username, days), h.createReturnKeyboard())
}

// handleResetTraffic handles the Reset Traffic action
func (h *AdminHandler) handleResetTraffic(c telebot.Context, username string) error {
	// Get inbounds
	inbounds, err := h.xrayService.GetInbounds(context.Background())
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
	err = h.xrayService.ResetUserTraffic(context.Background(), inboundID, username)
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
	if username == commands.ReturnToMainMenu {
		return h.handleStart(c)
	}

	// Delete client
	err := h.xrayService.RemoveClients(context.Background(), []string{username})
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
	if username == commands.ReturnToMainMenu {
		return h.handleStart(c)
	}

	// Get inbounds
	inbounds, err := h.xrayService.GetInbounds(context.Background())
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
	err = h.xrayService.ResetUserTraffic(context.Background(), inboundID, username)
	if err != nil {
		h.logger.Errorf("Failed to reset traffic: %v", err)
		return h.sendTextMessage(c, fmt.Sprintf("Failed to reset traffic: %v", err), nil)
	}

	return h.sendTextMessage(c, fmt.Sprintf("Traffic reset for %s.", username), h.createReturnKeyboard())
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
	message := h.formatDetailedUsersReport(inbounds)

	return h.sendTextMessage(c, message, h.createMainKeyboard(permissions.Admin))
}

// formatDetailedUsersReport formats a detailed users information report
func (h *AdminHandler) formatDetailedUsersReport(inbounds []models.Inbound) string {
	subscriptionSummary := h.aggregateUserDataBySubID(inbounds)

	if len(subscriptionSummary) == 0 {
		return "No active subscriptions found."
	}

	var totalUp, totalDown int64
	activeCount := 0

	for _, data := range subscriptionSummary {
		totalUp += data.TotalUp
		totalDown += data.TotalDown
		if data.Enable {
			activeCount++
		}
	}

	var sb strings.Builder
	sb.WriteString("<b>📊 Detailed Subscription Information</b>\n\n")

	totalUpGB := float64(totalUp) / (1024 * 1024 * 1024)
	totalDownGB := float64(totalDown) / (1024 * 1024 * 1024)

	for _, data := range subscriptionSummary {
		upGB := float64(data.TotalUp) / (1024 * 1024 * 1024)
		downGB := float64(data.TotalDown) / (1024 * 1024 * 1024)

		statusText := "🔴"
		if data.Enable {
			statusText = "🟢"
		}

		sb.WriteString(fmt.Sprintf("%s <b>%s</b>\n", statusText, strings.Join(data.Emails, ", ")))
		sb.WriteString(fmt.Sprintf("├ ⬆️ %.2f GB\n", upGB))
		sb.WriteString(fmt.Sprintf("└ ⬇️ %.2f GB\n\n", downGB))
	}

	sb.WriteString("<b>📈 Summary</b>\n")
	sb.WriteString(fmt.Sprintf("├ 👥 Total: %d subscriptions (%d active)\n", len(subscriptionSummary), activeCount))
	sb.WriteString(fmt.Sprintf("├ ⬆️ Total Upload: %.2f GB\n", totalUpGB))
	sb.WriteString(fmt.Sprintf("└ ⬇️ Total Download: %.2f GB\n\n", totalDownGB))

	return sb.String()
}

// SubscriptionSummary represents aggregated subscription data
type SubscriptionSummary struct {
	TotalUp      int64
	TotalDown    int64
	Enable       bool
	ExpiryTime   int64
	InboundNames []string
	Emails       []string
}

// aggregateUserDataBySubID aggregates user data by subscription ID from all inbounds
func (h *AdminHandler) aggregateUserDataBySubID(inbounds []models.Inbound) map[string]*SubscriptionSummary {
	subscriptionSummary := make(map[string]*SubscriptionSummary)

	// First, create email to subID mapping
	emailToSubID := h.createEmailToSubIDMapping(inbounds)

	for _, inbound := range inbounds {
		for _, clientStat := range inbound.ClientStats {
			// Get subID for this email
			subID, exists := emailToSubID[clientStat.Email]
			if !exists {
				// If no subID found, skip this client
				continue
			}

			if subscriptionSummary[subID] == nil {
				subscriptionSummary[subID] = &SubscriptionSummary{
					TotalUp:      0,
					TotalDown:    0,
					Enable:       clientStat.Enable,
					ExpiryTime:   clientStat.ExpiryTime,
					InboundNames: []string{},
					Emails:       []string{},
				}
			}

			summary := subscriptionSummary[subID]

			// Aggregate traffic data
			summary.TotalUp += clientStat.Up
			summary.TotalDown += clientStat.Down

			// Keep enabled status if any client is enabled
			if clientStat.Enable {
				summary.Enable = true
			}

			// Use the latest expiry time
			if clientStat.ExpiryTime > summary.ExpiryTime {
				summary.ExpiryTime = clientStat.ExpiryTime
			}

			// Add inbound name if not already present
			inboundFound := false
			for _, name := range summary.InboundNames {
				if name == inbound.Remark {
					inboundFound = true
					break
				}
			}
			if !inboundFound {
				summary.InboundNames = append(summary.InboundNames, inbound.Remark)
			}

			// Add email if not already present
			emailFound := false
			for _, email := range summary.Emails {
				if email == clientStat.Email {
					emailFound = true
					break
				}
			}
			if !emailFound {
				summary.Emails = append(summary.Emails, clientStat.Email)
			}
		}
	}

	return subscriptionSummary
}

// createEmailToSubIDMapping creates a mapping of email to subscription ID by parsing inbound settings
func (h *AdminHandler) createEmailToSubIDMapping(inbounds []models.Inbound) map[string]string {
	emailToSubID := make(map[string]string)

	for _, inbound := range inbounds {
		if inbound.Settings == "" {
			continue
		}

		var settings models.InboundSettings
		if err := json.Unmarshal([]byte(inbound.Settings), &settings); err != nil {
			h.logger.Errorf("Failed to parse inbound settings for inbound %d: %v", inbound.ID, err)
			continue
		}

		for _, client := range settings.Clients {
			if client.SubID != "" {
				emailToSubID[client.Email] = client.SubID
			}
		}
	}

	return emailToSubID
}
