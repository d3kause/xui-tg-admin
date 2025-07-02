package handlers

import (
	"context"
	"fmt"
	"hash/fnv"
	"strconv"
	"strings"

	telebot "gopkg.in/telebot.v3"

	"xui-tg-admin/internal/models"
	"xui-tg-admin/internal/services"
)

// AdminTrustedHandler handles admin operations for trusted user management
type AdminTrustedHandler struct {
	*BaseHandler
	storageService *services.StorageService
}

// NewAdminTrustedHandler creates a new admin trusted handler
func NewAdminTrustedHandler(base *BaseHandler, storageService *services.StorageService) *AdminTrustedHandler {
	return &AdminTrustedHandler{
		BaseHandler:    base,
		storageService: storageService,
	}
}

// HandleAddTrustedRequest handles the request to add a trusted user
func (h *AdminTrustedHandler) HandleAddTrustedRequest(ctx context.Context, c telebot.Context) error {
	state := models.UserState{
		State: models.StateAwaitingTrustedUsername,
	}
	h.stateService.SetState(c.Sender().ID, state)

	msg := "Send @username to add to trusted list:"
	return c.Send(msg)
}

// HandleRevokeTrustedRequest handles the request to show revoke menu
func (h *AdminTrustedHandler) HandleRevokeTrustedRequest(ctx context.Context, c telebot.Context) error {
	trustedUsers := h.storageService.GetTrustedUsers()

	if len(trustedUsers) == 0 {
		return c.Send("No trusted users found.")
	}

	keyboard := h.createRevokeTrustedKeyboard(trustedUsers)
	return c.Send("Select user to revoke:", &telebot.ReplyMarkup{InlineKeyboard: keyboard})
}

// HandleRevokeTrusted handles revoking a trusted user
func (h *AdminTrustedHandler) HandleRevokeTrusted(ctx context.Context, c telebot.Context, telegramID int64) error {
	if err := h.storageService.RemoveTrusted(telegramID); err != nil {
		h.logger.Errorf("Failed to remove trusted user: %v", err)
		return c.Send("Failed to revoke user.")
	}

	return c.Send("User revoked from trusted list.")
}

// HandleTrustedUsernameInput handles username input for adding trusted user
func (h *AdminTrustedHandler) HandleTrustedUsernameInput(ctx context.Context, c telebot.Context, text string) error {
	if !strings.HasPrefix(text, "@") {
		return c.Send("Please send a valid @username:")
	}

	username := strings.TrimPrefix(text, "@")

	// Generate pseudo telegram ID from username hash for consistency
	telegramID := generatePseudoTelegramID(username)

	if err := h.storageService.AddTrusted(telegramID, username); err != nil {
		h.logger.Errorf("Failed to add trusted user: %v", err)
		return c.Send("Failed to add user to trusted list.")
	}

	state := models.UserState{
		State: models.Default,
	}
	h.stateService.SetState(c.Sender().ID, state)
	return c.Send(fmt.Sprintf("@%s added to trusted list.", username))
}

// createRevokeTrustedKeyboard creates keyboard for revoking trusted users
func (h *AdminTrustedHandler) createRevokeTrustedKeyboard(trustedUsers []models.TrustedUser) [][]telebot.InlineButton {
	var keyboard [][]telebot.InlineButton

	for _, user := range trustedUsers {
		row := []telebot.InlineButton{
			{
				Text: fmt.Sprintf("❌ @%s", user.Username),
				Data: fmt.Sprintf("revoke_trusted_%d", user.TelegramID),
			},
		}
		keyboard = append(keyboard, row)
	}

	return keyboard
}

// ParseRevokeTrustedCallback parses the revoke trusted callback data
func ParseRevokeTrustedCallback(data string) (int64, error) {
	if !strings.HasPrefix(data, "revoke_trusted_") {
		return 0, fmt.Errorf("invalid callback data")
	}

	idStr := strings.TrimPrefix(data, "revoke_trusted_")
	return strconv.ParseInt(idStr, 10, 64)
}

// generatePseudoTelegramID generates a consistent pseudo telegram ID from username
func generatePseudoTelegramID(username string) int64 {
	h := fnv.New64a()
	h.Write([]byte(username))
	hash := h.Sum64()
	// Convert to int64 and ensure it's positive (Telegram IDs are positive)
	id := int64(hash & 0x7FFFFFFFFFFFFFFF)
	// Ensure it's not 0 (which we used as placeholder)
	if id == 0 {
		id = 1
	}
	return id
}
