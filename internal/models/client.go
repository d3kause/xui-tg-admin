package models

import (
	"crypto/rand"
	"encoding/hex"
)

// Client represents an X-ray client
type Client struct {
	ID         string  `json:"id"`
	Enable     bool    `json:"enable"`
	Flow       *string `json:"flow,omitempty"`
	Email      string  `json:"email"`
	TotalGB    int     `json:"totalGB"`
	LimitIP    int     `json:"limitIp"`
	ExpiryTime *int64  `json:"expiryTime,omitempty"`
	Fingerprint string `json:"fingerprint"`
	TgID       string  `json:"tgId"`
	SubID      string  `json:"subId"`
}

// ToDictionary converts the client to a map for API requests
func (c *Client) ToDictionary() map[string]interface{} {
	result := map[string]interface{}{
		"id":          c.ID,
		"enable":      c.Enable,
		"email":       c.Email,
		"totalGB":     c.TotalGB,
		"limitIp":     c.LimitIP,
		"fingerprint": c.Fingerprint,
		"tgId":        c.TgID,
		"subId":       c.SubID,
	}

	// Add optional fields if they exist
	if c.Flow != nil {
		result["flow"] = *c.Flow
	}

	if c.ExpiryTime != nil {
		result["expiryTime"] = *c.ExpiryTime
	}

	return result
}

// GenerateSubID generates a random subscription ID
func GenerateSubID() string {
	bytes := make([]byte, 16)
	_, err := rand.Read(bytes)
	if err != nil {
		return "sub_" + hex.EncodeToString([]byte("fallback"))
	}
	return "sub_" + hex.EncodeToString(bytes)
}