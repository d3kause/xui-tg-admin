package services

import (
	"encoding/json"
	"os"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"xui-tg-admin/internal/models"
)

// StorageData represents the JSON structure stored in data.json
type StorageData struct {
	TrustedUsers []models.TrustedUser `json:"trusted_users"`
	VpnAccounts  []models.VpnAccount  `json:"vpn_accounts"`
	NextID       int                  `json:"next_id"`
}

// StorageService handles JSON file operations for trusted users and VPN accounts
type StorageService struct {
	filename string
	data     *StorageData
	mu       sync.RWMutex
	logger   *logrus.Logger
}

// NewStorageService creates a new storage service
func NewStorageService(filename string, logger *logrus.Logger) *StorageService {
	s := &StorageService{
		filename: filename,
		data: &StorageData{
			TrustedUsers: make([]models.TrustedUser, 0),
			VpnAccounts:  make([]models.VpnAccount, 0),
			NextID:       1,
		},
		logger: logger,
	}

	if err := s.Load(); err != nil {
		logger.Warnf("Failed to load storage file: %v", err)
	}

	return s
}

// Load reads data from JSON file
func (s *StorageService) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.filename)
	if os.IsNotExist(err) {
		s.logger.Info("Storage file does not exist, starting with empty data")
		return nil
	}
	if err != nil {
		return err
	}

	return json.Unmarshal(data, s.data)
}

// Save writes data to JSON file atomically
func (s *StorageService) Save() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return err
	}

	tmpFile := s.filename + ".tmp"
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		return err
	}

	return os.Rename(tmpFile, s.filename)
}

// IsTrusted checks if a user is in the trusted list
func (s *StorageService) IsTrusted(telegramID int64) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, user := range s.data.TrustedUsers {
		if user.TelegramID == telegramID {
			return true
		}
	}
	return false
}

// IsTrustedByUsername checks if a username is in the trusted list and returns the stored telegram ID
func (s *StorageService) IsTrustedByUsername(username string) (bool, int64) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, user := range s.data.TrustedUsers {
		if user.Username == username {
			return true, user.TelegramID
		}
	}
	return false, 0
}

// UpdateTrustedUserTelegramID updates the telegram ID for a trusted user by username
func (s *StorageService) UpdateTrustedUserTelegramID(username string, realTelegramID int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, user := range s.data.TrustedUsers {
		if user.Username == username {
			s.data.TrustedUsers[i].TelegramID = realTelegramID
			return s.save()
		}
	}
	return nil
}

// AddTrusted adds a user to the trusted list
func (s *StorageService) AddTrusted(telegramID int64, username string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if already exists
	for _, user := range s.data.TrustedUsers {
		if user.TelegramID == telegramID {
			return nil // Already exists
		}
	}

	s.data.TrustedUsers = append(s.data.TrustedUsers, models.TrustedUser{
		TelegramID: telegramID,
		Username:   username,
		AddedAt:    time.Now().Unix(),
	})

	return s.save()
}

// RemoveTrusted removes a user from the trusted list
func (s *StorageService) RemoveTrusted(telegramID int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, user := range s.data.TrustedUsers {
		if user.TelegramID == telegramID {
			s.data.TrustedUsers = append(s.data.TrustedUsers[:i], s.data.TrustedUsers[i+1:]...)
			return s.save()
		}
	}
	return nil
}

// GetTrustedUsers returns all trusted users
func (s *StorageService) GetTrustedUsers() []models.TrustedUser {
	s.mu.RLock()
	defer s.mu.RUnlock()

	users := make([]models.TrustedUser, len(s.data.TrustedUsers))
	copy(users, s.data.TrustedUsers)
	return users
}

// GetUserAccountCount returns the number of VPN accounts created by a user
func (s *StorageService) GetUserAccountCount(telegramID int64) int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	count := 0
	for _, account := range s.data.VpnAccounts {
		if account.AddedBy == telegramID {
			count++
		}
	}
	return count
}

// AddVpnAccount adds a new VPN account
func (s *StorageService) AddVpnAccount(username, password string, addedBy int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data.VpnAccounts = append(s.data.VpnAccounts, models.VpnAccount{
		ID:        s.data.NextID,
		Username:  username,
		Password:  password,
		AddedBy:   addedBy,
		CreatedAt: time.Now().Unix(),
	})
	s.data.NextID++

	return s.save()
}

// RemoveVpnAccount removes a VPN account if it belongs to the specified user
func (s *StorageService) RemoveVpnAccount(id int, telegramID int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, account := range s.data.VpnAccounts {
		if account.ID == id && account.AddedBy == telegramID {
			s.data.VpnAccounts = append(s.data.VpnAccounts[:i], s.data.VpnAccounts[i+1:]...)
			return s.save()
		}
	}
	return nil
}

// GetUserAccounts returns all VPN accounts created by a specific user
func (s *StorageService) GetUserAccounts(telegramID int64) []models.VpnAccount {
	s.mu.RLock()
	defer s.mu.RUnlock()

	accounts := make([]models.VpnAccount, 0)
	for _, account := range s.data.VpnAccounts {
		if account.AddedBy == telegramID {
			accounts = append(accounts, account)
		}
	}
	return accounts
}

// save is an internal method that assumes the mutex is already locked
func (s *StorageService) save() error {
	data, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return err
	}

	tmpFile := s.filename + ".tmp"
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		return err
	}

	return os.Rename(tmpFile, s.filename)
}
