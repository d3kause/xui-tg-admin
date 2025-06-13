package handlers

import (
	"context"
	"fmt"
	"time"

	telebot "gopkg.in/telebot.v3"

	"xui-tg-admin/internal/commands"
	"xui-tg-admin/internal/constants"
	"xui-tg-admin/internal/helpers"
	"xui-tg-admin/internal/models"
	"xui-tg-admin/internal/validation"
)

// ClientCreationParams holds parameters for client creation
type ClientCreationParams struct {
	BaseUsername    string
	DurationStr     string
	ExpiryTime      int64
	CommonSubId     string
	BaseFingerprint string
	SenderID        int64
}

// createClientsForAllInbounds creates clients for all enabled inbounds
func (h *AdminHandler) createClientsForAllInbounds(ctx context.Context, params ClientCreationParams, enabledInbounds []models.Inbound) ([]string, []string, bool) {
	var addErrors []string
	var createdEmails []string
	var addedToAny bool

	for i, inbound := range enabledInbounds {
		email := fmt.Sprintf("%s-%d", params.BaseUsername, i+1)
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
			continue
		}

		addedToAny = true
		createdEmails = append(createdEmails, email)
		h.logger.Infof("Successfully added client %s to inbound %d", email, inbound.ID)
	}

	return createdEmails, addErrors, addedToAny
}

// getEnabledInbounds filters and returns only enabled inbounds
func (h *AdminHandler) getEnabledInbounds(ctx context.Context) ([]models.Inbound, error) {
	inbounds, err := h.xrayService.GetInbounds(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get inbounds: %w", err)
	}

	if len(inbounds) == 0 {
		return nil, fmt.Errorf("no inbounds available")
	}

	var enabledInbounds []models.Inbound
	for _, inbound := range inbounds {
		if inbound.Enable {
			enabledInbounds = append(enabledInbounds, inbound)
		}
	}

	if len(enabledInbounds) == 0 {
		return nil, fmt.Errorf("no enabled inbounds available")
	}

	return enabledInbounds, nil
}

// sendSubscriptionInfo sends subscription information and QR code to user
func (h *AdminHandler) sendSubscriptionInfo(c telebot.Context, params ClientCreationParams, createdEmails []string, addErrors []string) error {
	subscriptionInfo := helpers.FormatSubscriptionInfo(
		params.BaseUsername,
		params.DurationStr,
		params.ExpiryTime,
		createdEmails,
		params.CommonSubId,
		addErrors,
	)

	if err := h.sendTextMessage(c, subscriptionInfo, h.createReturnKeyboard()); err != nil {
		return err
	}

	if len(createdEmails) > 0 {
		subURL := fmt.Sprintf("https://iris.xele.one:2096/sub/%s?name=%s", params.CommonSubId, params.CommonSubId)
		if err := h.sendTextMessage(c, "QR code for subscription:", nil); err != nil {
			h.logger.Errorf("Failed to send QR code message: %v", err)
		} else if err := h.sendQRCode(c, subURL); err != nil {
			h.logger.Errorf("Failed to send QR code: %v", err)
		}
	}

	return nil
}

// calculateExpiryTime calculates expiry time based on duration
func calculateExpiryTime(durationStr string) (int64, error) {
	if durationStr == commands.Infinite {
		return 0, nil
	}

	days, err := validation.ValidateDuration(durationStr)
	if err != nil {
		return 0, err
	}

	return time.Now().Add(time.Duration(days) * 24 * time.Hour).UnixMilli(), nil
}

// findClientInInbounds finds a client by email in all inbounds
func (h *AdminHandler) findClientInInbounds(ctx context.Context, email string) (*models.Inbound, *models.ClientStat, error) {
	inbounds, err := h.xrayService.GetInbounds(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get inbounds: %w", err)
	}

	for _, inbound := range inbounds {
		for _, client := range inbound.ClientStats {
			if client.Email == email {
				return &inbound, &client, nil
			}
		}
	}

	return nil, nil, fmt.Errorf("client %s not found", email)
}

// extendClientDuration extends a client's duration by the specified days
func (h *AdminHandler) extendClientDuration(ctx context.Context, c telebot.Context, username string, days int) error {
	foundInbound, foundClient, err := h.findClientInInbounds(ctx, username)
	if err != nil {
		return h.sendTextMessage(c, err.Error(), h.createReturnKeyboard())
	}

	newExpiryTime := foundClient.ExpiryTime + (int64(days) * constants.MillisecondsInDay)

	updatedClient := models.Client{
		ID:         username,
		Enable:     foundClient.Enable,
		Email:      username,
		TotalGB:    int(foundClient.Total / constants.BytesInGB),
		LimitIP:    0,
		ExpiryTime: &newExpiryTime,
		TgID:       fmt.Sprintf("%d", c.Sender().ID),
		SubID:      models.GenerateSubID(),
	}

	if err := h.xrayService.RemoveClients(ctx, []string{username}); err != nil {
		h.logger.Errorf("Failed to remove old client: %v", err)
		return h.sendTextMessage(c, fmt.Sprintf("Failed to remove old client: %v", err), nil)
	}

	if err := h.xrayService.AddClient(ctx, foundInbound.ID, updatedClient); err != nil {
		h.logger.Errorf("Failed to add updated client: %v", err)
		return h.sendTextMessage(c, fmt.Sprintf("Failed to add updated client: %v", err), nil)
	}

	return h.sendTextMessage(c, fmt.Sprintf("Successfully extended duration for %s by %d days.", username, days), h.createReturnKeyboard())
}

// resetClientTraffic resets traffic for a specific client
func (h *AdminHandler) resetClientTraffic(ctx context.Context, c telebot.Context, username string) error {
	inbounds, err := h.xrayService.GetInbounds(ctx)
	if err != nil {
		return h.sendTextMessage(c, fmt.Sprintf("Failed to get inbounds: %v", err), nil)
	}

	var inboundID int
	var found bool

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

	if err := h.xrayService.ResetUserTraffic(ctx, inboundID, username); err != nil {
		h.logger.Errorf("Failed to reset traffic: %v", err)
		return h.sendTextMessage(c, fmt.Sprintf("Failed to reset traffic: %v", err), nil)
	}

	return h.sendTextMessage(c, fmt.Sprintf("Traffic reset for %s.", username), h.createReturnKeyboard())
}
