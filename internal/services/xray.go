package services

import (
	"context"

	"github.com/sirupsen/logrus"

	"xui-tg-admin/internal/config"
	"xui-tg-admin/internal/models"
	"xui-tg-admin/pkg/xrayclient"
)

// XrayService manages X-ray API client for a single server
type XrayService struct {
	client *xrayclient.Client
	config *config.Config
	logger *logrus.Logger
}

// NewXrayService creates a new X-ray service
func NewXrayService(cfg *config.Config, logger *logrus.Logger) *XrayService {
	client := xrayclient.NewClient(cfg.Server, logger)

	return &XrayService{
		client: client,
		config: cfg,
		logger: logger,
	}
}

// GetInbounds gets the inbounds from the server
func (s *XrayService) GetInbounds(ctx context.Context) ([]models.Inbound, error) {
	return s.client.GetInbounds(ctx)
}

// AddClient adds a client to an inbound on the server
func (s *XrayService) AddClient(ctx context.Context, inboundID int, client models.Client) error {
	return s.client.AddClientToInbound(ctx, inboundID, client)
}

// RemoveClients removes clients from the server
func (s *XrayService) RemoveClients(ctx context.Context, emails []string) error {
	return s.client.RemoveClients(ctx, emails)
}

// GetOnlineUsers gets the online users from the server
func (s *XrayService) GetOnlineUsers(ctx context.Context) ([]string, error) {
	return s.client.GetOnlineUsers(ctx)
}

// ResetUserTraffic resets a user's traffic on the server
func (s *XrayService) ResetUserTraffic(ctx context.Context, inboundID int, email string) error {
	return s.client.ResetUserTraffic(ctx, inboundID, email)
}

// GetSubscriptionURL gets a user's subscription URL from the server
func (s *XrayService) GetSubscriptionURL(ctx context.Context, email string) (string, error) {
	return s.client.GetSubscriptionURL(ctx, email)
}

// GetAllMembers gets all members from the server
func (s *XrayService) GetAllMembers(ctx context.Context) ([]string, error) {
	inbounds, err := s.GetInbounds(ctx)
	if err != nil {
		return nil, err
	}

	var members []string
	for _, inbound := range inbounds {
		for _, clientStat := range inbound.ClientStats {
			members = append(members, clientStat.Email)
		}
	}

	return members, nil
}

// GetServerNames gets the name of the configured server
func (s *XrayService) GetServerNames() []string {
	return []string{s.config.Server.Name}
}

// getClient gets the X-ray API client
func (s *XrayService) getClient(serverName string) (*xrayclient.Client, error) {
	// Ignore serverName parameter as we only have one server
	return s.client, nil
}
