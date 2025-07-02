package models

// TrustedUser represents a trusted user who can manage VPN accounts
type TrustedUser struct {
	TelegramID int64  `json:"telegram_id"`
	Username   string `json:"username"`
	AddedAt    int64  `json:"added_at"`
}

// VpnAccount represents a VPN account created by a trusted user
type VpnAccount struct {
	ID        int    `json:"id"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	AddedBy   int64  `json:"added_by"`
	CreatedAt int64  `json:"created_at"`
}