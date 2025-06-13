package models

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"strings"
)

// Client represents an X-ray client
type Client struct {
	ID          string  `json:"id"`
	Enable      bool    `json:"enable"`
	Flow        *string `json:"flow,omitempty"`
	Email       string  `json:"email"`
	TotalGB     int     `json:"totalGB"`
	LimitIP     int     `json:"limitIp"` // Note: using limitIp to match C# implementation
	ExpiryTime  *int64  `json:"expiryTime,omitempty"`
	Fingerprint string  `json:"fingerprint"`
	TgID        string  `json:"tgId"`
	SubID       string  `json:"subId"`
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
	// Generate UUID bytes
	uuid := make([]byte, 16)
	_, err := rand.Read(uuid)
	if err != nil {
		return "sub_" + hex.EncodeToString([]byte("fallback"))
	}

	// Convert to base64 and clean up
	b64 := base64.StdEncoding.EncodeToString(uuid)
	b64 = strings.ReplaceAll(b64, "=", "")
	b64 = strings.ReplaceAll(b64, "+", "")
	b64 = strings.ReplaceAll(b64, "/", "")

	// Take first 16 characters
	if len(b64) > 16 {
		b64 = b64[:16]
	}

	return b64
}
