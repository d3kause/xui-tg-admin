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
- **📱 User-friendly interface**: Intuitive buttons and menus
- **⚡ Fast operation**: Session caching and optimized requests
- **🔄 Automation**: Bulk operations and automatic management
- **📊 Monitoring**: Real-time traffic statistics
- **🔒 Security**: Access control verification and data validation

---

## 📋 Features

### 👑 Administrator
- ✅ **User creation** with expiration time settings
- 🔄 **Traffic management** (reset, monitoring)
- 👥 **Online users view**
- 📊 **Detailed usage statistics**
- 🗑️ **User deletion** with confirmation
- 🔗 **QR code generation** for configurations
- ⚙️ **Bulk operations** (reset traffic for all users)

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
go build -o bot ./cmd/bot

# 2. Set environment variables
export TG_TOKEN=your_telegram_bot_token
export TG_ADMIN_IDS=123456789,987654321
export XRAY_USER=admin
export XRAY_PASSWORD=password123
export XRAY_API_URL=http://localhost:8080/api

# 3. Run
./bot
```

---

## ⚙️ Configuration

### 📝 Configuration example

```env
# Telegram Bot
TG_TOKEN=123456789:ABCdefGHIjklMNOpqrSTUvwxYZ
TG_ADMIN_IDS=123456789,987654321

# X-UI Panel
XRAY_USER=admin
XRAY_PASSWORD=secure_password_123
XRAY_API_URL=http://your-server.com:54321/api
XRAY_SUB_URL_PREFIX=http://your-server.com:54321/sub

# Logging
LOG_LEVEL=info
```

### 🔑 Environment variables

| Variable | Description | Required | Example |
|----------|-------------|----------|---------|
| `TG_TOKEN` | Telegram Bot Token | ✅ | `123456789:ABCdef...` |
| `TG_ADMIN_IDS` | Admin IDs (comma-separated) | ✅ | `123456789,987654321` |
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
│    ⚙️ Main Menu         │
├─────────────────────────┤
│  👤 Add Member  │ 📊 Online │
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
   Add Member → Enter name → Choose duration → ✅ Done!
   ```

2. **User management**:
   ```
   Edit Member → Select user → Action → Result
   ```

3. **Monitoring**:
   ```
   Online Members → Active list
   Detailed Usage → Traffic statistics
   ```

---

## 🏗️ Architecture

### 📁 Project structure

```
xui-tg-admin/
├── 📂 cmd/bot/           # Application entry point
├── 📂 internal/          # Internal logic
│   ├── 📂 config/        # Configuration
│   ├── 📂 handlers/      # Telegram handlers
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

- **`handlers/`** - Telegram message handlers with role system
- **`services/`** - Business logic and X-UI API integration
- **`xrayclient/`** - HTTP client for X-UI API
- **`permissions/`** - Role and access control system

---

## 🛠️ Development

### 🔨 Building

```bash
# Development build
go build -o bot ./cmd/bot

# Production build
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bot ./cmd/bot
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
