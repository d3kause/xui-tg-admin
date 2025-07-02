package services

import (
	"context"
	"encoding/json"

	"github.com/sirupsen/logrus"

	"xui-tg-admin/internal/config"
	"xui-tg-admin/internal/helpers"
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

// GetAllMembersWithInfo получает детальную информацию о всех пользователях с поддержкой сортировки
func (s *XrayService) GetAllMembersWithInfo(ctx context.Context, sortType models.SortType) ([]models.MemberInfo, error) {
	inbounds, err := s.GetInbounds(ctx)
	if err != nil {
		return nil, err
	}

	// Создаем карту для группировки пользователей по базовому имени
	memberMap := make(map[string]*models.MemberInfo)

	// Собираем информацию из ClientStats
	for _, inbound := range inbounds {
		for _, clientStat := range inbound.ClientStats {
			baseUsername := helpers.ExtractBaseUsername(clientStat.Email)

			if memberInfo, exists := memberMap[baseUsername]; exists {
				// Обновляем существующую запись
				memberInfo.FullEmails = append(memberInfo.FullEmails, clientStat.Email)
				memberInfo.TotalUp += clientStat.Up
				memberInfo.TotalDown += clientStat.Down
				memberInfo.TotalTraffic += clientStat.Up + clientStat.Down

				// Обновляем статус и время истечения
				if clientStat.Enable {
					memberInfo.Enable = true
				}
				if clientStat.ExpiryTime > memberInfo.ExpiryTime {
					memberInfo.ExpiryTime = clientStat.ExpiryTime
				}
				// Используем наименьший ID для сортировки по порядку создания
				if clientStat.ID < memberInfo.ID {
					memberInfo.ID = clientStat.ID
				}
			} else {
				// Создаем новую запись
				memberInfo := &models.MemberInfo{
					BaseUsername: baseUsername,
					FullEmails:   []string{clientStat.Email},
					ID:           clientStat.ID,
					Enable:       clientStat.Enable,
					ExpiryTime:   clientStat.ExpiryTime,
					TotalUp:      clientStat.Up,
					TotalDown:    clientStat.Down,
					TotalTraffic: clientStat.Up + clientStat.Down,
				}
				memberMap[baseUsername] = memberInfo
			}
		}
	}

	// Получаем дополнительную информацию из InboundSettings для каждого пользователя
	for _, inbound := range inbounds {
		if inbound.Settings == "" {
			continue
		}

		var settings models.InboundSettings
		if err := json.Unmarshal([]byte(inbound.Settings), &settings); err != nil {
			continue
		}

		for _, client := range settings.Clients {
			baseUsername := helpers.ExtractBaseUsername(client.Email)
			if memberInfo, exists := memberMap[baseUsername]; exists {
				// Обновляем время истечения из настроек, если оно больше
				if client.ExpiryTime > memberInfo.ExpiryTime {
					memberInfo.ExpiryTime = client.ExpiryTime
				}
			}
		}
	}

	// Преобразуем карту в срез
	var members []models.MemberInfo
	for _, memberInfo := range memberMap {
		memberInfo.IsExpired = memberInfo.IsExpiredMember()
		members = append(members, *memberInfo)
	}

	// Сортируем по указанному типу
	models.SortMembers(members, sortType)

	return members, nil
}
