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
	// AwaitSortTypeForEdit is the state when the user is selecting sort type for edit member list
	AwaitSortTypeForEdit
	// AwaitSortTypeForDelete is the state when the user is selecting sort type for delete member list
	AwaitSortTypeForDelete
)

// UserState represents the state of a user's conversation
type UserState struct {
	State      ConversationState
	Payload    *string
	SortType   *SortType // Хранит выбранный тип сортировки
	ActionType *string   // Хранит тип действия (edit/delete)
}
