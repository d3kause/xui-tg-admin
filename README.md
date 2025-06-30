# 🚀 X-UI Telegram Admin Bot

<div align="center">

![Go Version](https://img.shields.io/badge/Go-1.24+-blue.svg)
![License](https://img.shields.io/badge/License-MIT-green.svg)
![Telegram](https://img.shields.io/badge/Telegram-Bot-blue.svg)
![X-Ray](https://img.shields.io/badge/X--Ray-Panel-orange.svg)

**Powerful Telegram bot for managing X-UI panel with role-based access and advanced features**

[🚀 Quick Start](#quick-start) • [📋 Features](#features) • [⚙️ Installation](#installation) • [🔧 Configuration](#configuration) • [📖 Usage](#usage)

</div>

---

## 🎯 What is this?

**X-UI Telegram Admin Bot** is a modern solution for managing VPN servers through Telegram. The bot provides full control over the X-UI panel directly from the messenger with an intuitive interface and role-based access system.

### 🌟 Key advantages

- **🔐 Role-based system**: Admin, User, Demo mode
- **📱 User-friendly interface**: Intuitive buttons and menus with proper HTML formatting
- **⚡ Fast operation**: Session caching and optimized requests
- **🔄 Automation**: Bulk operations and automatic management
- **📊 Monitoring**: Real-time traffic statistics and connection status
- **🔒 Security**: Access control verification and data validation
- **🎯 Smart navigation**: Universal button command handling with emoji support

---

## 📋 Features

### 👑 Administrator
- ✅ **User creation** with expiration time settings (including infinite duration)
- 🔄 **Traffic management** (reset individual or all users)
- 👥 **Online users view** with real-time connection status
- 📊 **Detailed usage statistics** with aggregated data
- 🗑️ **User deletion** with confirmation dialogs
- 🔗 **QR code generation** for configurations
- ⚙️ **Bulk operations** (reset traffic for all users)
- 🎯 **Smart navigation** with universal return buttons

### 👤 User
- 🔗 **View own configurations**
- 📱 **Get QR codes** for connection
- 📊 **Traffic monitoring**

### 🎭 Demo mode
- ℹ️ **Bot information**
- ❓ **Usage help**

---

## 🚀 Quick Start

### Requirements
- **Go 1.24+** or **Docker**
- **X-UI panel** with API access
- **Telegram Bot Token**

### ⚡ Quick installation with Docker

```bash
# 1. Clone repository
git clone https://github.com/yourusername/xui-tg-admin.git
cd xui-tg-admin

# 2. Configure settings
cp config.example.env .env
nano .env

# 3. Start
docker-compose up -d
```

### 🔧 Manual installation

```bash
# 1. Clone and build
git clone https://github.com/yourusername/xui-tg-admin.git
cd xui-tg-admin
go mod download
go build -o xui-tg-admin ./cmd/bot

# 2. Set environment variables
export TG_TOKEN=your_telegram_bot_token
export TG_ADMIN_IDS=123456789,987654321
export XRAY_SERVER=my-server
export XRAY_USER=admin
export XRAY_PASSWORD=password123
export XRAY_API_URL=http://localhost:8080/api

# 3. Run
./xui-tg-admin
```

---

## ⚙️ Configuration

### 📝 Configuration example

```env
# Telegram Bot Configuration
TG_TOKEN=1234567890:ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890
TG_ADMIN_IDS=123456789,987654321

# X-ray Server Configuration
XRAY_SERVER=my-server
XRAY_USER=admin
XRAY_PASSWORD=secure_password_123
XRAY_API_URL=http://your-server.com:54321/api
XRAY_SUB_URL_PREFIX=http://your-server.com:54321/sub

# Logging Configuration
LOG_LEVEL=info
```

### 🔑 Environment variables

| Variable | Description | Required | Example |
|----------|-------------|----------|---------|
| `TG_TOKEN` | Telegram Bot Token | ✅ | `1234567890:ABCdef...` |
| `TG_ADMIN_IDS` | Admin IDs (comma-separated) | ✅ | `123456789,987654321` |
| `XRAY_SERVER` | X-UI server identifier | ✅ | `my-server` |
| `XRAY_USER` | X-UI panel login | ✅ | `admin` |
| `XRAY_PASSWORD` | X-UI panel password | ✅ | `password123` |
| `XRAY_API_URL` | X-UI panel API URL | ✅ | `http://server.com:54321/api` |
| `XRAY_SUB_URL_PREFIX` | Subscription URL prefix | ❌ | `http://server.com:54321/sub` |
| `LOG_LEVEL` | Logging level | ❌ | `info` |

---

## 📖 Usage

### 🎮 Administrator interface

#### Main menu
```
┌─────────────────────────┐
│    🏠 Main Menu         │
├─────────────────────────┤
│  👤 Add Member  │ 🟢 Online │
│  ✏️ Edit Member │ 📈 Detailed│
│  🔄 Reset Network Usage │
└─────────────────────────┘
```

#### User management
```
┌─────────────────────────┐
│  👤 vasya_pupkin        │
├─────────────────────────┤
│  🔗 View Config         │
│  🔄 Reset │ 🗑️ Delete   │
│  ↩️ Return to Main Menu │
└─────────────────────────┘
```

### 📱 Administrator commands

| Command | Description | Example |
|---------|-------------|---------|
| `/start` | Start the bot | `/start` |
| `Add Member` | Add user | Creates user with expiration settings |
| `Edit Member` | Edit user | View, reset traffic, delete |
| `Online Members` | Online users | List of active connections |
| `Detailed Usage` | Detailed statistics | Traffic by users and inbounds |
| `Reset Network Usage` | Reset all traffic | Bulk operation with confirmation |

### 🔄 Workflow

1. **User creation**:
   ```
   Add Member → Enter name → Choose duration (∞ Infinite available) → ✅ Done!
   ```

2. **User management**:
   ```
   Edit Member → Select user → Action → Result
   ```

3. **Monitoring**:
   ```
   Online Members → Active list with real-time status
   Detailed Usage → Traffic statistics with aggregation
   ```

### 🎯 Smart Navigation

The bot features universal button handling:
- **↩️ Return to Main Menu** - Works from any state
- **∞ Infinite** - For unlimited duration subscriptions
- **✅ Confirm** - For confirmation dialogs
- **❌ Cancel** - For cancellation

---

## 🏗️ Architecture

### 📁 Project structure

```
xui-tg-admin/
├── 📂 cmd/bot/           # Application entry point
├── 📂 internal/          # Internal logic
│   ├── 📂 commands/      # Command constants
│   ├── 📂 config/        # Configuration
│   ├── 📂 constants/     # Application constants
│   ├── 📂 errors/        # Error handling
│   ├── 📂 handlers/      # Telegram handlers
│   ├── 📂 helpers/       # Helper functions
│   ├── 📂 models/        # Data models
│   ├── 📂 permissions/   # Access control system
│   ├── 📂 services/      # Business logic
│   └── 📂 validation/    # Data validation
├── 📂 pkg/               # Reusable packages
│   ├── 📂 telegrambot/   # Telegram bot
│   └── 📂 xrayclient/    # X-UI API client
└── 📄 Configuration files
```

### 🔧 Main components

- **`handlers/`** - Telegram message handlers with role system and smart button handling
- **`services/`** - Business logic and X-UI API integration
- **`xrayclient/`** - HTTP client for X-UI API with session management
- **`permissions/`** - Role and access control system
- **`commands/`** - Centralized command constants

---

## 🛠️ Development

### 🔨 Building

```bash
# Development build
go build -o xui-tg-admin ./cmd/bot

# Production build
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o xui-tg-admin ./cmd/bot
```

### 🧪 Testing

```bash
# Run tests
go test ./...

# Tests with coverage
go test -cover ./...
```

### 📝 Logging

```bash
# Log levels
LOG_LEVEL=debug  # Detailed logs
LOG_LEVEL=info   # Information messages
LOG_LEVEL=warn   # Warnings only
LOG_LEVEL=error  # Errors only
```

---

## 🆕 Recent Updates

### ✅ Fixed Issues
- **Smart button handling**: Universal command extraction from emoji buttons
- **HTML formatting**: Proper `<b>` tags rendering in all messages
- **Navigation**: Return to Main Menu works from any state
- **User experience**: Improved error messages and confirmation dialogs

### 🎯 Key Improvements
- **Universal button processing**: Single function handles all emoji buttons
- **Better error handling**: More informative error messages
- **Consistent UI**: All messages use proper HTML formatting
- **Robust navigation**: Return buttons work reliably across all states

---

## 🤝 Contributing

We welcome contributions to the project!

### 📋 How to help

1. 🍴 Fork the repository
2. 🌿 Create a branch for new feature
3. 💾 Commit your changes
4. 🔀 Create a Pull Request

### 📝 Code standards

- Follow [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Use `gofmt` for formatting
- Add tests for new functionality
- Update documentation when changing API

---

## 📄 License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

---

## 🙏 Acknowledgments

- [X-UI](https://github.com/vaxilu/x-ui) - Excellent X-Ray management panel
- [Telegram Bot API](https://core.telegram.org/bots/api) - Telegram Bot API
- [Go](https://golang.org/) - Go programming language

---

<div align="center">

**⭐ If you liked the project, give it a star!**

[🚀 Start using](#quick-start) • [📖 Documentation](#usage) • [🐛 Report bug](https://github.com/yourusername/xui-tg-admin/issues)

</div>
