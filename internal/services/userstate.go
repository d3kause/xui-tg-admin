package services

import (
	"fmt"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/sirupsen/logrus"

	"xui-tg-admin/internal/models"
)

// UserStateService manages user conversation states
type UserStateService struct {
	cache  *cache.Cache
	logger *logrus.Logger
}

// NewUserStateService creates a new user state service
func NewUserStateService(logger *logrus.Logger) *UserStateService {
	return &UserStateService{
		cache:  cache.New(30*time.Minute, 10*time.Minute),
		logger: logger,
	}
}

// GetState gets a user's state
func (s *UserStateService) GetState(userID int64) (*models.UserState, error) {
	key := fmt.Sprintf("user_state_%d", userID)
	
	if data, found := s.cache.Get(key); found {
		if state, ok := data.(*models.UserState); ok {
			return state, nil
		}
		return nil, fmt.Errorf("invalid state type for user %d", userID)
	}
	
	// Return default state if not found
	return &models.UserState{
		State:          models.Default,
		Payload:        nil,
		SelectedServer: nil,
	}, nil
}

// SetState sets a user's state
func (s *UserStateService) SetState(userID int64, state models.UserState) error {
	key := fmt.Sprintf("user_state_%d", userID)
	s.cache.Set(key, &state, cache.DefaultExpiration)
	s.logger.Debugf("Set state for user %d: %+v", userID, state)
	return nil
}

// ClearState clears a user's state
func (s *UserStateService) ClearState(userID int64) error {
	key := fmt.Sprintf("user_state_%d", userID)
	s.cache.Delete(key)
	s.logger.Debugf("Cleared state for user %d", userID)
	return nil
}

// WithSelectedServer updates a user's selected server
func (s *UserStateService) WithSelectedServer(userID int64, server string) error {
	state, err := s.GetState(userID)
	if err != nil {
		return err
	}
	
	state.SelectedServer = &server
	return s.SetState(userID, *state)
}

// WithConversationState updates a user's conversation state
func (s *UserStateService) WithConversationState(userID int64, conversationState models.ConversationState) error {
	state, err := s.GetState(userID)
	if err != nil {
		return err
	}
	
	state.State = conversationState
	return s.SetState(userID, *state)
}

// WithPayload updates a user's payload
func (s *UserStateService) WithPayload(userID int64, payload string) error {
	state, err := s.GetState(userID)
	if err != nil {
		return err
	}
	
	state.Payload = &payload
	return s.SetState(userID, *state)
}