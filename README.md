# X-UI Telegram Bot

A Telegram bot for managing X-ray VPN panel users with role-based access control and comprehensive user management features.

## Features

- X-ray panel management via REST API
- Telegram bot with role-based access (Admin/Member/Demo)
- User creation, deletion, and traffic management
- QR code generation for VPN configurations
- Session-based conversation state management
- Detailed traffic usage statistics
- Automatic authentication and session caching

## Requirements

- Go 1.24 or higher
- Docker and Docker Compose (for containerized deployment)
- X-ray panel with API access

## Installation

### Using Docker (Recommended)

1. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/xui-tg-admin.git
   cd xui-tg-admin
   ```

2. Create a `.env` file with your configuration:
   ```env
   TG_TOKEN=your_telegram_bot_token
   TG_ADMIN_IDS=123456789,987654321
   XRAY_USER=admin
   XRAY_PASSWORD=password123
   XRAY_API_URL=http://localhost:8080/api
   XRAY_SUB_URL_PREFIX=http://localhost:8080/sub
   LOG_LEVEL=info
   ```

3. Build and start the container:
   ```bash
   docker-compose up -d
   ```

### Manual Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/xui-tg-admin.git
   cd xui-tg-admin
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Build the application:
   ```bash
   go build -o bot ./cmd/bot
   ```

4. Set environment variables:
   ```bash
   export TG_TOKEN=your_telegram_bot_token
   export TG_ADMIN_IDS=123456789,987654321
   export XRAY_USER=admin
   export XRAY_PASSWORD=password123
   export XRAY_API_URL=http://localhost:8080/api
   export XRAY_SUB_URL_PREFIX=http://localhost:8080/sub
   export LOG_LEVEL=info
   ```

5. Run the application:
   ```bash
   ./bot
   ```

## Configuration

The application is configured using environment variables:

### Example Configuration

An example configuration file is provided in `config.example.env`. You can use this as a template for your own configuration:

```bash
# Copy the example configuration
cp config.example.env .env

# Edit the configuration with your values
nano .env
```

| Variable | Description | Required | Example |
|----------|-------------|----------|---------|
| TG_TOKEN | Telegram Bot Token | Yes | `123456789:ABCdefGHIjklMNOpqrSTUvwxYZ` |
| TG_ADMIN_IDS | Comma-separated list of Telegram user IDs with admin access | Yes | `123456789,987654321` |
| XRAY_USER | X-ray panel username | Yes | `admin` |
| XRAY_PASSWORD | X-ray panel password | Yes | `password123` |
| XRAY_API_URL | X-ray panel API URL | Yes | `http://localhost:8080/api` |
| XRAY_SUB_URL_PREFIX | Subscription URL prefix | No | `http://localhost:8080/sub` |
| LOG_LEVEL | Log level (debug, info, warn, error) | No | `info` |

## Usage

### Admin Commands

- `/start` - Start the bot and show the main menu
- `Add Member` - Add a new member
- `Edit Member` - Edit an existing member
- `Delete Member` - Delete a member
- `Online Members` - View online members
- `Network Usage` - View network usage statistics
- `Detailed Usage` - View detailed user statistics
- `Reset Network Usage` - Reset network usage for a member

### Member Commands

- `/start` - Start the bot and show the main menu
- `Create New Config` - Create a new VPN configuration
- `View Configs Info` - View information about your configurations

### Demo Commands

- `/start` - Start the bot and show the main menu
- `About` - Show information about the bot
- `Help` - Show help information

## Architecture

The application follows a clean architecture approach with the following components:

- **cmd/bot**: Main application entry point
- **internal/config**: Configuration management
- **internal/errors**: Custom error types
- **internal/handlers**: Telegram message handlers
- **internal/models**: Data models
- **internal/permissions**: Access control
- **internal/services**: Business logic services
- **pkg/telegrambot**: Telegram bot framework
- **pkg/xrayclient**: X-ray API client

## Development

### Project Structure

```
.
├── cmd
│   └── bot
│       └── main.go
├── internal
│   ├── config
│   │   ├── config.go
│   │   └── loader.go
│   ├── errors
│   │   └── errors.go
│   ├── handlers
│   │   ├── admin.go
│   │   ├── base.go
│   │   ├── demo.go
│   │   ├── factory.go
│   │   └── member.go
│   ├── models
│   │   ├── client.go
│   │   ├── inbound.go
│   │   └── userstate.go
│   ├── permissions
│   │   └── controller.go
│   └── services
│       ├── qr.go
│       ├── userstate.go
│       ├── validator.go
│       └── xray.go
├── pkg
│   ├── telegrambot
│   │   └── bot.go
│   └── xrayclient
│       └── client.go
├── config.example.env
├── Dockerfile
├── docker-compose.yml
├── go.mod
├── go.sum
└── README.md
```

### Building and Testing

To build the application:

```bash
go build -o bot ./cmd/bot
```

To run tests:

```bash
go test ./...
```

### Key Dependencies

- **telebot.v3** - Telegram Bot API framework
- **logrus** - Structured logging
- **resty** - HTTP client for API requests
- **viper** - Configuration management
- **go-cache** - In-memory caching
- **go-qrcode** - QR code generation

## Features

### User Management

- Create new users with automatic configuration generation
- Edit existing users (extend expiration time)
- Delete users with confirmation
- View detailed traffic usage statistics

### Monitoring

- View online users
- Network usage statistics by inbounds
- Detailed subscription statistics with grouping
- Reset traffic counters

### Security

- Role-based access control (Admin/Member/Demo)
- Automatic authentication with session caching
- Input data validation
- Graceful shutdown with proper signal handling

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgements

- [telebot](https://github.com/tucnak/telebot) - Telegram Bot API framework
- [X-UI](https://github.com/vaxilu/x-ui) - X-ray panel
- [resty](https://github.com/go-resty/resty) - HTTP client for Go
- [logrus](https://github.com/sirupsen/logrus) - Structured logging
