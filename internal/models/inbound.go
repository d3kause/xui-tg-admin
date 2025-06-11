package models

// Inbound represents an X-ray inbound configuration
type Inbound struct {
	ID          int          `json:"id"`
	Up          int64        `json:"up"`
	Down        int64        `json:"down"`
	Total       int64        `json:"total"`
	Remark      string       `json:"remark"`
	Enable      bool         `json:"enable"`
	ExpiryTime  int64        `json:"expiryTime"`
	ClientStats []ClientStat `json:"clientStats"`
	Listen      string       `json:"listen"`
	Port        int          `json:"port"`
	Protocol    string       `json:"protocol"`
	Settings    string       `json:"settings"`
}

// ClientStat represents statistics for a client
type ClientStat struct {
	ID         int    `json:"id"`
	InboundID  int    `json:"inboundId"`
	Enable     bool   `json:"enable"`
	Email      string `json:"email"`
	Up         int64  `json:"up"`
	Down       int64  `json:"down"`
	ExpiryTime int64  `json:"expiryTime"`
	Total      int64  `json:"total"`
	Reset      int64  `json:"reset"`
}

// InboundSettings represents the parsed settings of an inbound
type InboundSettings struct {
	Clients []InboundClient `json:"clients"`
}

// InboundClient represents a client in inbound settings
type InboundClient struct {
	Email      string `json:"email"`
	Enable     bool   `json:"enable"`
	ExpiryTime int64  `json:"expiryTime"`
	SubID      string `json:"subId"`
	TgID       string `json:"tgId"`
}
