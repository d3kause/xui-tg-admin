package models

// ConversationState represents the state of a conversation with a user
type ConversationState int

const (
	// Default is the initial state
	Default ConversationState = iota
	// AwaitingInputUserName is the state when the user is inputting a username
	AwaitingInputUserName
	// AwaitingDuration is the state when the user is inputting a duration
	AwaitingDuration
	// AwaitSelectUserName is the state when the user is selecting a username
	AwaitSelectUserName
	// AwaitMemberAction is the state when the user is selecting an action for a member
	AwaitMemberAction
	// AwaitConfirmMemberDeletion is the state when the user is confirming member deletion
	AwaitConfirmMemberDeletion
	// AwaitConfirmResetUsersNetworkUsage is the state when the user is confirming network usage reset
	AwaitConfirmResetUsersNetworkUsage
	// StateAwaitingTrustedUsername is the state when admin is inputting trusted username
	StateAwaitingTrustedUsername
	// StateAwaitingVpnUsername is the state when trusted user is inputting VPN username
	StateAwaitingVpnUsername
	// StateAwaitingVpnPassword is the state when trusted user is inputting VPN password
	StateAwaitingVpnPassword
)

// Additional state constants for trusted user functionality
const (
	StateDefault = Default
)

// UserState represents the state of a user's conversation
type UserState struct {
	State      ConversationState
	Payload    *string
	SortType   *SortType // Хранит выбранный тип сортировки
	ActionType *string   // Хранит тип действия (edit/delete)
}
