# 🤖 Claude AI Agent Guide - X-UI Telegram Admin Bot

## 📋 Project Overview

**X-UI Telegram Admin Bot** is a Go application for managing X-UI panel through a Telegram bot with role-based access system.

### 🎯 Main Purpose
Automation of VPN server management through a convenient Telegram interface with role support (Admin/User/Demo).

---

## 🏗️ Project Architecture

### 📁 Directory Structure

```
xui-tg-admin/
├── cmd/bot/                    # Application entry point
│   └── main.go                # Main file with initialization
├── internal/                   # Internal application logic
│   ├── commands/              # Telegram command constants
│   ├── config/                # Configuration management
│   ├── constants/             # Application constants
│   ├── errors/                # Custom error types
│   ├── handlers/              # Telegram message handlers
│   │   ├── admin.go           # Administrator logic
│   │   ├── admin_client_operations.go # Client operations
│   │   ├── base.go            # Base handler
│   │   ├── demo.go            # Demo mode
│   │   ├── factory.go         # Handler factory
│   │   └── member.go          # User logic
│   ├── helpers/               # Helper functions
│   │   ├── grouping.go        # Data grouping
│   │   ├── subscription.go    # Subscription handling
│   │   └── traffic.go         # Traffic formatting
│   ├── models/                # Data models
│   │   ├── client.go          # Client model
│   │   ├── inbound.go         # Inbound model
│   │   └── userstate.go       # User state
│   ├── permissions/           # Access control system
│   │   └── controller.go      # Permission controller
│   ├── services/              # Business logic
│   │   ├── qr.go              # QR code generation
│   │   ├── userstate.go       # State management
│   │   ├── validator.go       # Data validation
│   │   └── xray.go            # X-UI service
│   └── validation/            # Validation
│       └── validation.go      # Validation rules
├── pkg/                       # Reusable packages
│   ├── telegrambot/           # Telegram bot
│   │   └── bot.go             # Main bot
│   └── xrayclient/            # X-UI API client
│       └── client.go          # HTTP client for X-UI
└── Configuration files
    ├── config.example.env     # Configuration example
    ├── docker-compose.yml     # Docker Compose
    └── Dockerfile             # Docker image
```

---

## 🔧 Key Components

### 1. **Handler System (handlers/)**

#### `admin.go` - Main administrator logic
- **States**: `AwaitingInputUserName`, `AwaitingDuration`, `AwaitSelectUserName`, `AwaitMemberAction`, `AwaitConfirmMemberDeletion`, `AwaitConfirmResetUsersNetworkUsage`
- **Main methods**:
  - `handleStart()` - Main menu
  - `handleAddMember()` - Add user
  - `handleEditMember()` - Edit user
  - `handleDeleteMember()` - Delete user
  - `handleViewConfig()` - View configuration
  - `handleResetTraffic()` - Reset traffic
  - `processConfirmDeletion()` - Deletion confirmation

#### `admin_client_operations.go` - Client operations
- **Main methods**:
  - `createClientsForAllInbounds()` - Create clients in all inbounds
  - `sendSubscriptionInfo()` - Send subscription information
  - `findClientInInbounds()` - Find client in inbounds

#### `base.go` - Base handler
- **Common methods**:
  - `createMainKeyboard()` - Create main keyboard
  - `createReturnKeyboard()` - Return keyboard
  - `createConfirmKeyboard()` - Confirmation keyboard
  - `sendTextMessage()` - Send text message
  - `sendQRCode()` - Send QR code

### 2. **X-UI API Client (pkg/xrayclient/)**

#### `client.go` - HTTP client for X-UI
- **Main methods**:
  - `Login()` - Authentication
  - `GetInbounds()` - Get inbounds
  - `AddClientToInbound()` - Add client
  - `RemoveClients()` - Remove clients
  - `ResetUserTraffic()` - Reset traffic
  - `GetOnlineUsers()` - Get online users

### 3. **Data Models (internal/models/)**

#### `client.go` - Client model
```go
type Client struct {
    ID          string  `json:"id"`
    Enable      bool    `json:"enable"`
    Email       string  `json:"email"`
    TotalGB     int     `json:"totalGB"`
    LimitIP     int     `json:"limitIp"`
    ExpiryTime  *int64  `json:"expiryTime,omitempty"`
    Fingerprint string  `json:"fingerprint"`
    TgID        string  `json:"tgId"`
    SubID       string  `json:"subId"`
}
```

#### `inbound.go` - Inbound model
```go
type Inbound struct {
    ID          int          `json:"id"`
    Up          int64        `json:"up"`
    Down        int64        `json:"down"`
    Total       int64        `json:"total"`
    Remark      string       `json:"remark"`
    Enable      bool         `json:"enable"`
    ExpiryTime  int64        `json:"expiryTime"`
    ClientStats []ClientStat `json:"clientStats"`
    Settings    string       `json:"settings"`
}
```

### 4. **Permission System (internal/permissions/)**

#### `controller.go` - Permission controller
- **Roles**: `Admin`, `Member`, `Demo`
- **Methods**:
  - `GetAccessType()` - Determine access type
  - `IsAdmin()` - Check if admin

---

## 🔄 Workflows

### 1. **User Creation**
```
Add Member → Enter name → Validation → Choose duration → Create in inbounds → Send QR code
```

### 2. **User Management**
```
Edit Member → Select user → Action menu → Execute action → Result
```

### 3. **User Deletion**
```
Delete → Select user → Confirmation → Delete from all inbounds → Result
```

### 4. **Traffic Reset**
```
Reset Traffic → Select user → Reset in all inbounds → Result
```

---

## 🎮 User Interface

### Keyboards

#### Administrator main menu
```
┌─────────────────────────┐
│  👤 Add Member  │ 📊 Online │
│  ✏️ Edit Member │ 📈 Detailed│
│  🔄 Reset Network Usage │
└─────────────────────────┘
```

#### User action menu
```
┌─────────────────────────┐
│  🔗 View Config         │
│  🔄 Reset │ 🗑️ Delete   │
│  ↩️ Return to Main Menu │
└─────────────────────────┘
```

### User States
- `Default` - Default state
- `AwaitingInputUserName` - Waiting for username input
- `AwaitingDuration` - Waiting for duration input
- `AwaitSelectUserName` - Waiting for user selection
- `AwaitMemberAction` - Waiting for user action
- `AwaitConfirmMemberDeletion` - Waiting for deletion confirmation
- `AwaitConfirmResetUsersNetworkUsage` - Waiting for traffic reset confirmation

---

## 🔧 API Integration

### X-UI API Endpoints
- `POST /login` - Authentication
- `GET /xui/API/inbounds` - Get inbounds
- `POST /xui/API/inbounds/addClient` - Add client
- `POST /xui/API/inbounds/{inboundId}/delClient/{clientUuid}` - Delete client
- `POST /xui/API/inbounds/{inboundId}/resetClientTraffic/{clientEmail}` - Reset traffic
- `POST /xui/API/inbounds/onlines` - Get online users

### Data Format
- **Clients**: Created with suffixes (`username-1`, `username-2`, etc.)
- **SubID**: Common for all clients of one user
- **Traffic**: Unlimited (TotalGB: 0)
- **Expiration**: Unix timestamp in milliseconds

---

## 🛠️ Development

### Main Dependencies
- `gopkg.in/telebot.v3` - Telegram Bot API
- `github.com/go-resty/resty/v2` - HTTP client
- `github.com/sirupsen/logrus` - Logging
- `github.com/patrickmn/go-cache` - Caching
- `github.com/skip2/go-qrcode` - QR codes

### Configuration
```env
TG_TOKEN=your_telegram_bot_token
TG_ADMIN_IDS=123456789,987654321
XRAY_USER=admin
XRAY_PASSWORD=password123
XRAY_API_URL=http://localhost:8080/api
XRAY_SUB_URL_PREFIX=http://localhost:8080/sub
LOG_LEVEL=info
```

### Build and Run
```bash
# Build
go build -o bot ./cmd/bot

# Run
./bot
```

---

## 🐛 Debugging and Logging

### Log Levels
- `debug` - Detailed logs for development
- `info` - Information messages
- `warn` - Warnings
- `error` - Errors

### Key Logs
- X-UI API authentication
- Client creation/deletion
- API request errors
- User states

---

## 🔒 Security

### Data Validation
- Username format validation
- Duration validation
- Access control verification

### Role System
- **Admin**: Full access to all functions
- **Member**: Limited access to own configurations
- **Demo**: Information viewing only

---

## 📝 Development Notes

### Important Points
1. **States**: Always check current user state
2. **Caching**: X-UI API sessions are cached for optimization
3. **Error Handling**: All API calls handle authentication errors
4. **Bulk Operations**: Deletion and traffic reset work with all inbounds
5. **QR Codes**: Generated for each user with common SubID

### Recommendations
- Use existing methods for creating keyboards
- Follow state handling pattern
- Add logging for debugging
- Check access rights before performing operations 