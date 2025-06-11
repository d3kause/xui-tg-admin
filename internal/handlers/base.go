package handlers

import (
	"bytes"
	"fmt"

	"github.com/sirupsen/logrus"
	telebot "gopkg.in/telebot.v3"

	"xui-tg-admin/internal/config"
	"xui-tg-admin/internal/models"
	"xui-tg-admin/internal/permissions"
	"xui-tg-admin/internal/services"
)

// BaseHandler provides common functionality for all handlers
type BaseHandler struct {
	xrayService  *services.XrayService
	stateService *services.UserStateService
	qrService    *services.QRService
	config       *config.Config
	logger       *logrus.Logger
}

// NewBaseHandler creates a new base handler
func NewBaseHandler(
	xrayService *services.XrayService,
	stateService *services.UserStateService,
	qrService *services.QRService,
	config *config.Config,
	logger *logrus.Logger,
) BaseHandler {
	return BaseHandler{
		xrayService:  xrayService,
		stateService: stateService,
		qrService:    qrService,
		config:       config,
		logger:       logger,
	}
}

// CanHandle checks if the handler can handle the given access type
func (h *BaseHandler) CanHandle(accessType permissions.AccessType) bool {
	// Base handler can't handle any access type directly
	return false
}

// sendTextMessage sends a text message with optional markup
func (h *BaseHandler) sendTextMessage(c telebot.Context, text string, markup *telebot.ReplyMarkup) error {
	opts := &telebot.SendOptions{
		ParseMode: telebot.ModeHTML,
	}

	if markup != nil {
		opts.ReplyMarkup = markup
	}

	_, err := c.Bot().Send(c.Recipient(), text, opts)
	if err != nil {
		h.logger.Errorf("Failed to send message: %v", err)
	}
	return err
}

// sendQRCode sends a QR code for the given URL
func (h *BaseHandler) sendQRCode(c telebot.Context, url string) error {
	// Generate QR code
	qrBytes, err := h.qrService.GenerateQR(url)
	if err != nil {
		h.logger.Errorf("Failed to generate QR code: %v", err)
		return err
	}

	// Create photo from bytes
	reader := bytes.NewReader(qrBytes)
	photo := &telebot.Photo{File: telebot.FromReader(reader)}

	// Send photo
	_, err = c.Bot().Send(c.Recipient(), photo)
	if err != nil {
		h.logger.Errorf("Failed to send QR code: %v", err)
	}
	return err
}

// createMainKeyboard creates the main keyboard for the given access type
func (h *BaseHandler) createMainKeyboard(accessType permissions.AccessType) *telebot.ReplyMarkup {
	markup := &telebot.ReplyMarkup{
		ResizeKeyboard: true,
	}

	var rows []telebot.Row

	switch accessType {
	case permissions.Admin:
		rows = []telebot.Row{
			{
				telebot.Btn{Text: "Add Member"},
				telebot.Btn{Text: "Edit Member"},
			},
			{
				telebot.Btn{Text: "Delete Member"},
				telebot.Btn{Text: "Online Members"},
			},
			{
				telebot.Btn{Text: "Network Usage"},
				telebot.Btn{Text: "Reset Network Usage"},
			},
			{
				telebot.Btn{Text: "Change Server"},
			},
		}
	case permissions.Member:
		rows = []telebot.Row{
			{
				telebot.Btn{Text: "Create New Config"},
				telebot.Btn{Text: "View Configs Info"},
			},
		}
	case permissions.Demo:
		rows = []telebot.Row{
			{
				telebot.Btn{Text: "About"},
				telebot.Btn{Text: "Help"},
			},
		}
	}

	markup.Reply(rows...)
	return markup
}

// createReturnKeyboard creates a keyboard with a return button
func (h *BaseHandler) createReturnKeyboard() *telebot.ReplyMarkup {
	markup := &telebot.ReplyMarkup{
		ResizeKeyboard: true,
	}

	markup.Reply(
		telebot.Row{
			telebot.Btn{Text: "Return to Main Menu"},
		},
	)

	return markup
}

// createConfirmKeyboard creates a keyboard with confirm/cancel buttons
func (h *BaseHandler) createConfirmKeyboard() *telebot.ReplyMarkup {
	markup := &telebot.ReplyMarkup{
		ResizeKeyboard: true,
	}

	markup.Reply(
		telebot.Row{
			telebot.Btn{Text: "Confirm"},
			telebot.Btn{Text: "Cancel"},
		},
	)

	return markup
}
