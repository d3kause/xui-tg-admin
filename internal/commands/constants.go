package commands

// TelegramCommands contains all commands for the Telegram bot
const (
	// Main commands
	Start  = "/start"
	Cancel = "Cancel"

	// Navigation commands
	ReturnToMainMenu = "Return to Main Menu"

	// Administrator commands
	AddMember         = "Add Member"
	EditMember        = "Edit Member"
	DeleteMember      = "Delete Member"
	OnlineMembers     = "Online Members"
	NetworkUsage      = "Network Usage"
	DetailedUsage     = "Detailed Usage"
	ResetNetworkUsage = "Reset Network Usage"

	// Member commands
	CreateNewConfig = "Create New Config"
	ViewConfigsInfo = "View Configs Info"

	// Demo user commands
	About = "About"
	Help  = "Help"

	// Member action commands
	ViewConfig   = "View Config"
	ResetTraffic = "Reset Traffic"
	Delete       = "Delete"

	// Confirmation commands
	Confirm = "Confirm"

	// Duration options
	Infinite = "∞ Infinite"
)
