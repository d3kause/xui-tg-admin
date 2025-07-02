package handlers

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	telebot "gopkg.in/telebot.v3"

	"xui-tg-admin/internal/commands"
	"xui-tg-admin/internal/helpers"
	"xui-tg-admin/internal/models"
	"xui-tg-admin/internal/permissions"
	"xui-tg-admin/internal/services"
)

// TrustedHandler handles trusted user operations
type TrustedHandler struct {
	BaseHandler
	storageService  *services.StorageService
	commandHandlers map[string]func(telebot.Context) error
}

// NewTrustedHandler creates a new trusted handler
func NewTrustedHandler(base *BaseHandler, storageService *services.StorageService) *TrustedHandler {
	handler := &TrustedHandler{
		BaseHandler:    *base,
		storageService: storageService,
	}

	handler.initializeCommands()
	return handler
}

// CanHandle checks if the handler can handle the given access type
func (h *TrustedHandler) CanHandle(accessType permissions.AccessType) bool {
	return accessType == permissions.Trusted
}

// Handle handles incoming updates for trusted users
func (h *TrustedHandler) Handle(ctx context.Context, c telebot.Context) error {
	// Handle callback queries
	if c.Callback() != nil {
		return h.handleCallback(ctx, c)
	}

	// Get user ID
	userID := c.Sender().ID

	// Check account limit before any operation
	accountCount := h.storageService.GetUserAccountCount(userID)
	if accountCount >= 3 && c.Text() == "➕ "+commands.AddMember {
		return c.Send("You can create maximum 3 accounts.")
	}

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
	case models.AwaitConfirmMemberDeletion:
		return h.processConfirmDeletion(c)
	default:
		h.logger.Warnf("Unknown state: %d", userState.State)
		return h.handleDefaultState(c)
	}
}

// initializeCommands initializes the command handlers
func (h *TrustedHandler) initializeCommands() {
	h.commandHandlers = map[string]func(telebot.Context) error{
		commands.Start:            h.handleStart,
		commands.AddMember:        h.handleAddMember,
		commands.DeleteMember:     h.handleDeleteMember,
		commands.ReturnToMainMenu: h.handleStart,
		commands.Cancel:           h.handleStart,
	}
}

// getButtonCommand extracts the command from button text with emoji
func (h *TrustedHandler) getButtonCommand(text string) string {
	// Check for specific button patterns
	switch text {
	case "↩️ " + commands.ReturnToMainMenu:
		return commands.ReturnToMainMenu
	case "❌ " + commands.Cancel:
		return commands.Cancel
	case "✅ " + commands.Confirm:
		return commands.Confirm
	case "➕ " + commands.AddMember:
		return commands.AddMember
	case "🗑 " + commands.DeleteMember:
		return commands.DeleteMember
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
func (h *TrustedHandler) handleDefaultState(c telebot.Context) error {
	text := c.Text()
	command := h.getButtonCommand(text)

	// Check if we have a command handler for this command
	if handler, ok := h.commandHandlers[command]; ok {
		return handler(c)
	}

	// If not, show the main menu
	return h.handleStart(c)
}

// handleStart handles the start command
func (h *TrustedHandler) handleStart(c telebot.Context) error {
	// Clear state
	h.stateService.WithConversationState(c.Sender().ID, models.Default)

	// Determine the message based on command
	var message string
	if c.Text() == commands.Start {
		message = "Welcome! You are a trusted user."
	} else {
		message = "Main Menu"
	}

	// Create and send keyboard
	keyboard := h.createMainKeyboard(permissions.Trusted)
	return h.sendTextMessage(c, message, keyboard)
}

// handleAddMember handles adding a new member (VPN account)
func (h *TrustedHandler) handleAddMember(c telebot.Context) error {
	userID := c.Sender().ID

	// Check account limit
	accountCount := h.storageService.GetUserAccountCount(userID)
	if accountCount >= 3 {
		return c.Send("You can create maximum 3 accounts.")
	}

	// Get user's Telegram username
	username := c.Sender().Username
	if username == "" {
		return c.Send("Error: You need to set a Telegram username first. Go to Telegram Settings -> Edit Profile -> Username")
	}

	// Generate auto username based on Telegram username and account count
	autoUsername := fmt.Sprintf("%s-add%d", username, accountCount+1)

	// Send loading message
	loadingMsg := fmt.Sprintf("Creating account '%s'...", autoUsername)
	c.Send(loadingMsg)

	// Create clients for all inbounds with infinite duration
	params := TrustedClientCreationParams{
		Username:    autoUsername,
		ExpiryTime:  0, // Infinite duration
		SenderID:    userID,
		CommonSubId: generateSubID(autoUsername),
	}

	success, errors := h.createClientsForAllInbounds(params)

	// Store VPN account in our storage
	if success {
		if err := h.storageService.AddVpnAccount(autoUsername, "auto-generated", userID); err != nil {
			h.logger.Errorf("Failed to store VPN account: %v", err)
		}
	}

	// Send result
	if success {
		h.sendSubscriptionInfo(c, params)
	} else {
		errorMsg := "Failed to create account:\n" + strings.Join(errors, "\n")
		c.Send(errorMsg)
	}

	// Return to main menu
	return h.handleStart(c)
}

// handleDeleteMember handles showing user's accounts for deletion
func (h *TrustedHandler) handleDeleteMember(c telebot.Context) error {
	userID := c.Sender().ID
	accounts := h.storageService.GetUserAccounts(userID)

	if len(accounts) == 0 {
		return c.Send("You have no accounts to remove.")
	}

	keyboard := h.createRemoveAccountKeyboard(accounts)
	return c.Send("Select account to remove:", &telebot.ReplyMarkup{InlineKeyboard: keyboard})
}

// handleCallback handles callback queries
func (h *TrustedHandler) handleCallback(ctx context.Context, c telebot.Context) error {
	data := c.Callback().Data

	if strings.HasPrefix(data, "remove_vpn_") {
		return h.handleConfirmRemoveVpnAccount(ctx, c, data)
	}

	return c.Send("Unknown action.")
}

// handleConfirmRemoveVpnAccount handles showing confirmation for VPN account removal
func (h *TrustedHandler) handleConfirmRemoveVpnAccount(ctx context.Context, c telebot.Context, data string) error {
	userID := c.Sender().ID

	accountID, err := parseRemoveVpnCallback(data)
	if err != nil {
		return c.Send("Invalid account selection.")
	}

	// Get the account details
	accounts := h.storageService.GetUserAccounts(userID)
	var accountToDelete *models.VpnAccount
	for _, account := range accounts {
		if account.ID == accountID {
			accountToDelete = &account
			break
		}
	}

	if accountToDelete == nil {
		return c.Send("Account not found.")
	}

	// Store account ID in state for confirmation
	accountIDStr := fmt.Sprintf("%d", accountID)
	h.stateService.WithPayload(userID, accountIDStr)
	h.stateService.WithConversationState(userID, models.AwaitConfirmMemberDeletion)

	// Show confirmation keyboard
	markup := h.createConfirmKeyboard()
	return c.Send(fmt.Sprintf("🗑️ **Confirm Account Deletion**\n\n⚠️ You are about to permanently delete account **%s**\n\n**This action will:**\n• Remove account from all server configurations\n• Delete all associated data\n• Cannot be undone\n\nAre you absolutely sure?", accountToDelete.Username), &telebot.SendOptions{
		ParseMode:   telebot.ModeMarkdown,
		ReplyMarkup: markup,
	})
}

// processConfirmDeletion processes the deletion confirmation
func (h *TrustedHandler) processConfirmDeletion(c telebot.Context) error {
	userID := c.Sender().ID
	confirmation := c.Text()

	// Check for return to main menu
	if h.getButtonCommand(confirmation) == commands.ReturnToMainMenu {
		return h.handleStart(c)
	}

	// Check if user confirmed
	if h.getButtonCommand(confirmation) != commands.Confirm {
		return c.Send("❌ **Invalid Selection**\n\nPlease click Confirm to proceed with deletion or use the Return button to cancel.")
	}

	// Get account ID from state
	userState, err := h.stateService.GetState(userID)
	if err != nil || userState.Payload == nil {
		return c.Send("❌ **Session Error**\n\nAccount data was lost. Please start the deletion process again.")
	}

	accountIDStr := *userState.Payload
	accountID, err := strconv.Atoi(accountIDStr)
	if err != nil {
		return c.Send("❌ **Invalid Account ID**\n\nPlease start the deletion process again.")
	}

	// Get the account details before deletion
	accounts := h.storageService.GetUserAccounts(userID)
	var accountToDelete *models.VpnAccount
	for _, account := range accounts {
		if account.ID == accountID {
			accountToDelete = &account
			break
		}
	}

	if accountToDelete == nil {
		return c.Send("❌ **Account Not Found**\n\nThe account may have already been deleted.")
	}

	// Send loading message
	loadingMsg := fmt.Sprintf("⏳ **Deleting Account...**\n\nRemoving account '%s' from all server configurations. Please wait...", accountToDelete.Username)
	c.Send(loadingMsg)

	// First, remove clients from X-Ray server (like admin does)
	ctx := context.Background()
	err = h.xrayService.RemoveClients(ctx, []string{accountToDelete.Username})
	if err != nil {
		h.logger.Errorf("Failed to remove clients from X-Ray server: %v", err)
		// Clear state and return to main menu
		h.stateService.WithConversationState(userID, models.Default)
		return c.Send(fmt.Sprintf("❌ **Deletion Failed**\n\nCouldn't delete account '%s' from server configurations.\n\n**Error:** %v\n\nPlease try again or contact administrator.", accountToDelete.Username, err))
	}

	// Then remove from our database
	if err := h.storageService.RemoveVpnAccount(accountID, userID); err != nil {
		h.logger.Errorf("Failed to remove VPN account from storage: %v", err)
		// Clear state and return to main menu
		h.stateService.WithConversationState(userID, models.Default)
		return c.Send(fmt.Sprintf("⚠️ **Partial Success**\n\nAccount deleted from server but failed to update database:\n%v", err))
	}

	// Clear state and return to main menu
	h.stateService.WithConversationState(userID, models.Default)
	return c.Send(fmt.Sprintf("✅ **Account Deleted Successfully**\n\n🗑️ Account '%s' has been permanently removed from all server configurations.", accountToDelete.Username))
}

// createRemoveAccountKeyboard creates keyboard for removing accounts
func (h *TrustedHandler) createRemoveAccountKeyboard(accounts []models.VpnAccount) [][]telebot.InlineButton {
	var keyboard [][]telebot.InlineButton

	for _, account := range accounts {
		row := []telebot.InlineButton{
			{
				Text: fmt.Sprintf("❌ %s", account.Username),
				Data: fmt.Sprintf("remove_vpn_%d", account.ID),
			},
		}
		keyboard = append(keyboard, row)
	}

	return keyboard
}

// parseRemoveVpnCallback parses the remove VPN callback data
func parseRemoveVpnCallback(data string) (int, error) {
	if !strings.HasPrefix(data, "remove_vpn_") {
		return 0, fmt.Errorf("invalid callback data")
	}

	idStr := strings.TrimPrefix(data, "remove_vpn_")
	return strconv.Atoi(idStr)
}

// TrustedClientCreationParams holds parameters for client creation
type TrustedClientCreationParams struct {
	Username    string
	ExpiryTime  int64
	SenderID    int64
	CommonSubId string
}

// generateSubID generates a subscription ID for the user
func generateSubID(username string) string {
	return models.GenerateSubID()
}

// createClientsForAllInbounds creates clients for all enabled inbounds (simplified version)
func (h *TrustedHandler) createClientsForAllInbounds(params TrustedClientCreationParams) (bool, []string) {
	ctx := context.Background()

	// Get enabled inbounds
	inbounds, err := h.xrayService.GetInbounds(ctx)
	if err != nil {
		h.logger.Errorf("Failed to get inbounds: %v", err)
		return false, []string{"Failed to get server configuration"}
	}

	// Filter enabled inbounds
	var enabledInbounds []models.Inbound
	for _, inbound := range inbounds {
		if inbound.Enable {
			enabledInbounds = append(enabledInbounds, inbound)
		}
	}

	if len(enabledInbounds) == 0 {
		return false, []string{"No enabled inbounds found"}
	}

	// Create client creation params using admin-compatible format
	adminParams := ClientCreationParams{
		BaseUsername:    params.Username,
		DurationStr:     "∞",
		ExpiryTime:      params.ExpiryTime,
		CommonSubId:     params.CommonSubId,
		BaseFingerprint: fmt.Sprintf("%x", time.Now().UnixNano()),
		SenderID:        params.SenderID,
	}

	// Create clients using admin logic
	createdEmails, addErrors, success := h.createClientsForAllInboundsAdmin(ctx, adminParams, enabledInbounds)

	h.logger.Infof("Created %d clients for user %s", len(createdEmails), params.Username)
	return success, addErrors
}

// createClientsForAllInboundsAdmin creates clients using admin logic
func (h *TrustedHandler) createClientsForAllInboundsAdmin(ctx context.Context, params ClientCreationParams, enabledInbounds []models.Inbound) ([]string, []string, bool) {
	var addErrors []string
	var createdEmails []string
	var addedToAny bool

	for i, inbound := range enabledInbounds {
		email := helpers.FormatEmailWithInboundNumber(params.BaseUsername, i+1)
		fingerprint := fmt.Sprintf("%s-%d", params.BaseFingerprint, i+1)

		client := models.Client{
			ID:          email,
			Enable:      true,
			Email:       email,
			TotalGB:     0, // Unlimited traffic
			LimitIP:     0, // No IP limit
			ExpiryTime:  &params.ExpiryTime,
			TgID:        fmt.Sprintf("%d", params.SenderID),
			SubID:       params.CommonSubId,
			Fingerprint: fingerprint,
		}

		if err := h.xrayService.AddClient(ctx, inbound.ID, client); err != nil {
			h.logger.Errorf("Failed to add client to inbound %d: %v", inbound.ID, err)
			addErrors = append(addErrors, fmt.Sprintf("Inbound %d: %v", inbound.ID, err))
		} else {
			h.logger.Infof("Successfully added client %s to inbound %d", email, inbound.ID)
			createdEmails = append(createdEmails, email)
			addedToAny = true
		}
	}

	return createdEmails, addErrors, addedToAny
}

// sendSubscriptionInfo sends subscription information to the user using admin format
func (h *TrustedHandler) sendSubscriptionInfo(c telebot.Context, params TrustedClientCreationParams) error {
	// Create admin-compatible params
	adminParams := ClientCreationParams{
		BaseUsername:    params.Username,
		DurationStr:     "∞",
		ExpiryTime:      params.ExpiryTime,
		CommonSubId:     params.CommonSubId,
		BaseFingerprint: fmt.Sprintf("%x", time.Now().UnixNano()),
		SenderID:        params.SenderID,
	}

	// Get created emails (we need this for the helper function)
	ctx := context.Background()
	inbounds, err := h.xrayService.GetInbounds(ctx)
	if err != nil {
		return err
	}

	var createdEmails []string
	var enabledCount int
	for _, inbound := range inbounds {
		if inbound.Enable {
			enabledCount++
			email := helpers.FormatEmailWithInboundNumber(params.Username, enabledCount)
			createdEmails = append(createdEmails, email)
		}
	}

	// Use admin helper to format subscription info
	subscriptionInfo := helpers.FormatSubscriptionInfo(
		adminParams.BaseUsername,
		adminParams.DurationStr,
		adminParams.ExpiryTime,
		createdEmails,
		adminParams.CommonSubId,
		[]string{}, // No errors for successful creation
		h.config.Server.SubURLPrefix,
	)

	if err := h.sendTextMessage(c, subscriptionInfo, nil); err != nil {
		return err
	}

	// Send QR code with correct URL format (same as admin)
	if len(createdEmails) > 0 {
		subURL := fmt.Sprintf("%s%s?name=%s", h.config.Server.SubURLPrefix, params.CommonSubId, params.CommonSubId)
		if err := h.sendTextMessage(c, "QR code for subscription:", nil); err != nil {
			h.logger.Errorf("Failed to send QR code message: %v", err)
		} else if err := h.sendQRCode(c, subURL); err != nil {
			h.logger.Errorf("Failed to send QR code: %v", err)
		}
	}

	return nil
}

// createConfirmKeyboard creates a keyboard for confirmation
func (h *TrustedHandler) createConfirmKeyboard() *telebot.ReplyMarkup {
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
