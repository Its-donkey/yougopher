package streaming

import "time"

// Message types returned by the YouTube Live Chat API.
const (
	MessageTypeText                   = "textMessageEvent"
	MessageTypeSuperChat              = "superChatEvent"
	MessageTypeSuperSticker           = "superStickerEvent"
	MessageTypeMembership             = "newSponsorEvent"
	MessageTypeMemberMilestone        = "memberMilestoneChatEvent"
	MessageTypeGiftMembershipReceived = "giftMembershipReceivedEvent"
	MessageTypeMembershipGifting      = "membershipGiftingEvent"
	MessageTypeChatEnded      = "chatEndedEvent"
	MessageTypeMessageDeleted = "messageDeletedEvent"
	MessageTypeUserBanned     = "userBannedEvent"
	// MessageTypePoll covers all poll states. To distinguish poll states,
	// check msg.Snippet.PollDetails.Status (PollStatusOpen, PollStatusClosed).
	MessageTypePoll      = "pollEvent"
	MessageTypeTombstone = "tombstone" // Deleted/moderated message
)

// LiveChatMessage represents a message from the YouTube Live Chat API.
type LiveChatMessage struct {
	// Kind identifies the resource type.
	Kind string `json:"kind,omitempty"`

	// ETag for caching.
	ETag string `json:"etag,omitempty"`

	// ID is the unique identifier for this message.
	ID string `json:"id,omitempty"`

	// Snippet contains message metadata and content.
	Snippet *MessageSnippet `json:"snippet,omitempty"`

	// AuthorDetails contains information about the message author.
	AuthorDetails *AuthorDetails `json:"authorDetails,omitempty"`
}

// MessageSnippet contains the core message data.
type MessageSnippet struct {
	// Type indicates the message type (e.g., "textMessageEvent", "superChatEvent").
	Type string `json:"type,omitempty"`

	// LiveChatID is the ID of the live chat this message belongs to.
	LiveChatID string `json:"liveChatId,omitempty"`

	// AuthorChannelID is the channel ID of the author.
	AuthorChannelID string `json:"authorChannelId,omitempty"`

	// PublishedAt is when the message was published.
	PublishedAt time.Time `json:"publishedAt,omitempty"`

	// HasDisplayContent indicates if the message has displayable content.
	HasDisplayContent bool `json:"hasDisplayContent,omitempty"`

	// DisplayMessage is the text shown to viewers.
	DisplayMessage string `json:"displayMessage,omitempty"`

	// TextMessageDetails contains details for text messages.
	TextMessageDetails *TextMessageDetails `json:"textMessageDetails,omitempty"`

	// SuperChatDetails contains details for Super Chat messages.
	SuperChatDetails *SuperChatDetails `json:"superChatDetails,omitempty"`

	// SuperStickerDetails contains details for Super Sticker messages.
	SuperStickerDetails *SuperStickerDetails `json:"superStickerDetails,omitempty"`

	// MemberMilestoneChatDetails contains details for member milestone messages.
	MemberMilestoneChatDetails *MemberMilestoneChatDetails `json:"memberMilestoneChatDetails,omitempty"`

	// NewSponsorDetails contains details for new membership events.
	NewSponsorDetails *NewSponsorDetails `json:"newSponsorDetails,omitempty"`

	// MembershipGiftingDetails contains details for membership gifting events.
	MembershipGiftingDetails *MembershipGiftingDetails `json:"membershipGiftingDetails,omitempty"`

	// GiftMembershipReceivedDetails contains details for received gift memberships.
	GiftMembershipReceivedDetails *GiftMembershipReceivedDetails `json:"giftMembershipReceivedDetails,omitempty"`

	// MessageDeletedDetails contains details for deleted message events.
	MessageDeletedDetails *MessageDeletedDetails `json:"messageDeletedDetails,omitempty"`

	// UserBannedDetails contains details for user ban events.
	UserBannedDetails *UserBannedDetails `json:"userBannedDetails,omitempty"`

	// PollDetails contains details for poll events.
	PollDetails *PollDetails `json:"pollDetails,omitempty"`
}

// AuthorDetails contains information about a message author.
type AuthorDetails struct {
	// ChannelID is the YouTube channel ID of the author.
	ChannelID string `json:"channelId,omitempty"`

	// ChannelURL is the URL to the author's channel.
	ChannelURL string `json:"channelUrl,omitempty"`

	// DisplayName is the author's display name.
	DisplayName string `json:"displayName,omitempty"`

	// ProfileImageURL is the URL to the author's profile image.
	ProfileImageURL string `json:"profileImageUrl,omitempty"`

	// IsVerified indicates if the author is a verified YouTube user.
	IsVerified bool `json:"isVerified,omitempty"`

	// IsChatOwner indicates if the author owns the live chat.
	IsChatOwner bool `json:"isChatOwner,omitempty"`

	// IsChatSponsor indicates if the author is a channel member.
	IsChatSponsor bool `json:"isChatSponsor,omitempty"`

	// IsChatModerator indicates if the author is a moderator.
	IsChatModerator bool `json:"isChatModerator,omitempty"`
}

// TextMessageDetails contains details for a text message.
type TextMessageDetails struct {
	// MessageText is the text content of the message.
	MessageText string `json:"messageText,omitempty"`
}

// SuperChatDetails contains details for a Super Chat donation.
type SuperChatDetails struct {
	// AmountMicros is the donation amount in micros (1/1,000,000 of currency unit).
	AmountMicros int64 `json:"amountMicros,omitempty,string"`

	// Currency is the ISO 4217 currency code.
	Currency string `json:"currency,omitempty"`

	// AmountDisplayString is the formatted donation amount (e.g., "$5.00").
	AmountDisplayString string `json:"amountDisplayString,omitempty"`

	// UserComment is the optional message from the donor.
	UserComment string `json:"userComment,omitempty"`

	// Tier is the Super Chat tier (1-7).
	Tier int `json:"tier,omitempty"`
}

// SuperStickerDetails contains details for a Super Sticker.
type SuperStickerDetails struct {
	// SuperStickerID is the ID of the sticker.
	SuperStickerID string `json:"superStickerId,omitempty"`

	// SuperStickerMetadata contains sticker display information.
	SuperStickerMetadata *SuperStickerMetadata `json:"superStickerMetadata,omitempty"`

	// AmountMicros is the sticker cost in micros.
	AmountMicros int64 `json:"amountMicros,omitempty,string"`

	// Currency is the ISO 4217 currency code.
	Currency string `json:"currency,omitempty"`

	// AmountDisplayString is the formatted sticker cost.
	AmountDisplayString string `json:"amountDisplayString,omitempty"`

	// Tier is the Super Sticker tier.
	Tier int `json:"tier,omitempty"`
}

// SuperStickerMetadata contains display information for a Super Sticker.
type SuperStickerMetadata struct {
	// StickerID is the unique sticker identifier.
	StickerID string `json:"stickerId,omitempty"`

	// AltText is alternative text describing the sticker.
	AltText string `json:"altText,omitempty"`

	// AltTextLanguage is the language of the alt text.
	AltTextLanguage string `json:"altTextLanguage,omitempty"`
}

// MemberMilestoneChatDetails contains details for a member milestone message.
type MemberMilestoneChatDetails struct {
	// MemberLevelName is the name of the membership level.
	MemberLevelName string `json:"memberLevelName,omitempty"`

	// MemberMonth is the number of months the user has been a member.
	MemberMonth int `json:"memberMonth,omitempty"`

	// UserComment is the optional milestone message.
	UserComment string `json:"userComment,omitempty"`
}

// NewSponsorDetails contains details for a new channel membership.
type NewSponsorDetails struct {
	// MemberLevelName is the name of the membership level.
	MemberLevelName string `json:"memberLevelName,omitempty"`

	// IsUpgrade indicates if this is an upgrade from a lower tier.
	IsUpgrade bool `json:"isUpgrade,omitempty"`
}

// MembershipGiftingDetails contains details for gifted memberships.
type MembershipGiftingDetails struct {
	// GiftMembershipsCount is the number of memberships gifted.
	GiftMembershipsCount int `json:"giftMembershipsCount,omitempty"`

	// MemberLevelName is the membership level being gifted.
	MemberLevelName string `json:"memberLevelName,omitempty"`
}

// GiftMembershipReceivedDetails contains details for a received gift membership.
type GiftMembershipReceivedDetails struct {
	// MemberLevelName is the membership level received.
	MemberLevelName string `json:"memberLevelName,omitempty"`

	// GifterChannelID is the channel ID of the gifter.
	GifterChannelID string `json:"gifterChannelId,omitempty"`

	// AssociatedMembershipGiftingMessageID links to the gifting event.
	AssociatedMembershipGiftingMessageID string `json:"associatedMembershipGiftingMessageId,omitempty"`
}

// MessageDeletedDetails contains details for a deleted message event.
type MessageDeletedDetails struct {
	// DeletedMessageID is the ID of the deleted message.
	DeletedMessageID string `json:"deletedMessageId,omitempty"`
}

// UserBannedDetails contains details for a user ban event.
type UserBannedDetails struct {
	// BannedUserDetails contains information about the banned user.
	BannedUserDetails *BannedUserDetails `json:"bannedUserDetails,omitempty"`

	// BanType indicates the type of ban.
	BanType string `json:"banType,omitempty"`

	// BanDurationSeconds is the ban duration (0 for permanent bans).
	BanDurationSeconds int64 `json:"banDurationSeconds,omitempty,string"`
}

// BannedUserDetails contains information about a banned user.
type BannedUserDetails struct {
	// ChannelID is the banned user's channel ID.
	ChannelID string `json:"channelId,omitempty"`

	// ChannelURL is the banned user's channel URL.
	ChannelURL string `json:"channelUrl,omitempty"`

	// DisplayName is the banned user's display name.
	DisplayName string `json:"displayName,omitempty"`

	// ProfileImageURL is the banned user's profile image URL.
	ProfileImageURL string `json:"profileImageUrl,omitempty"`
}

// BanType constants.
const (
	BanTypePermanent = "permanent"
	BanTypeTemporary = "temporary"
)

// PollDetails contains details for poll events.
type PollDetails struct {
	// PollID is the unique poll identifier.
	PollID string `json:"pollId,omitempty"`

	// Question is the poll question.
	Question string `json:"question,omitempty"`

	// Choices contains the poll options.
	Choices []PollChoice `json:"choices,omitempty"`

	// Status indicates the poll state.
	Status string `json:"status,omitempty"`
}

// PollChoice represents a single poll option.
type PollChoice struct {
	// ChoiceID is the unique choice identifier.
	ChoiceID string `json:"choiceId,omitempty"`

	// Text is the choice text.
	Text string `json:"text,omitempty"`

	// NumVotes is the vote count (available in closed polls).
	NumVotes int64 `json:"numVotes,omitempty,string"`
}

// Poll status constants.
const (
	PollStatusOpen   = "open"
	PollStatusClosed = "closed"
)

// LiveChatMessageListResponse is the response from liveChatMessages.list.
type LiveChatMessageListResponse struct {
	// Kind identifies the resource type.
	Kind string `json:"kind,omitempty"`

	// ETag for caching.
	ETag string `json:"etag,omitempty"`

	// NextPageToken for pagination.
	NextPageToken string `json:"nextPageToken,omitempty"`

	// PollingIntervalMillis indicates how long to wait before polling again.
	PollingIntervalMillis int `json:"pollingIntervalMillis,omitempty"`

	// OfflineAt is set when the chat has ended.
	OfflineAt *time.Time `json:"offlineAt,omitempty"`

	// PageInfo contains pagination metadata.
	PageInfo *PageInfo `json:"pageInfo,omitempty"`

	// Items contains the messages.
	Items []*LiveChatMessage `json:"items,omitempty"`
}

// PageInfo contains pagination metadata.
type PageInfo struct {
	TotalResults   int `json:"totalResults,omitempty"`
	ResultsPerPage int `json:"resultsPerPage,omitempty"`
}

// PollingInterval returns the recommended polling interval as a Duration.
func (r *LiveChatMessageListResponse) PollingInterval() time.Duration {
	if r.PollingIntervalMillis > 0 {
		return time.Duration(r.PollingIntervalMillis) * time.Millisecond
	}
	return 0
}

// IsChatEnded returns true if the chat has ended.
func (r *LiveChatMessageListResponse) IsChatEnded() bool {
	return r.OfflineAt != nil
}

// Message returns the display message text.
func (m *LiveChatMessage) Message() string {
	if m.Snippet == nil {
		return ""
	}
	return m.Snippet.DisplayMessage
}

// Type returns the message type.
func (m *LiveChatMessage) Type() string {
	if m.Snippet == nil {
		return ""
	}
	return m.Snippet.Type
}

// IsTextMessage returns true if this is a regular text message.
func (m *LiveChatMessage) IsTextMessage() bool {
	return m.Type() == MessageTypeText
}

// IsSuperChat returns true if this is a Super Chat message.
func (m *LiveChatMessage) IsSuperChat() bool {
	return m.Type() == MessageTypeSuperChat
}

// IsSuperSticker returns true if this is a Super Sticker message.
func (m *LiveChatMessage) IsSuperSticker() bool {
	return m.Type() == MessageTypeSuperSticker
}

// IsMembership returns true if this is a new membership event.
func (m *LiveChatMessage) IsMembership() bool {
	return m.Type() == MessageTypeMembership
}

// IsMemberMilestone returns true if this is a member milestone message.
func (m *LiveChatMessage) IsMemberMilestone() bool {
	return m.Type() == MessageTypeMemberMilestone
}

// IsGiftMembership returns true if this is a gift membership event.
func (m *LiveChatMessage) IsGiftMembership() bool {
	return m.Type() == MessageTypeMembershipGifting
}

// LiveChatBan represents a ban in the live chat.
type LiveChatBan struct {
	// Kind identifies the resource type.
	Kind string `json:"kind,omitempty"`

	// ETag for caching.
	ETag string `json:"etag,omitempty"`

	// ID is the unique identifier for this ban.
	ID string `json:"id,omitempty"`

	// Snippet contains ban details.
	Snippet *BanSnippet `json:"snippet,omitempty"`
}

// BanSnippet contains ban details.
type BanSnippet struct {
	// LiveChatID is the ID of the live chat.
	LiveChatID string `json:"liveChatId,omitempty"`

	// BanType is "permanent" or "temporary".
	BanType string `json:"type,omitempty"`

	// BanDurationSeconds is the duration for temporary bans.
	BanDurationSeconds int64 `json:"banDurationSeconds,omitempty,string"`

	// BannedUserDetails contains information about the banned user.
	BannedUserDetails *BannedUserDetails `json:"bannedUserDetails,omitempty"`
}

// LiveChatModerator represents a moderator in the live chat.
type LiveChatModerator struct {
	// Kind identifies the resource type.
	Kind string `json:"kind,omitempty"`

	// ETag for caching.
	ETag string `json:"etag,omitempty"`

	// ID is the unique identifier for this moderator entry.
	ID string `json:"id,omitempty"`

	// Snippet contains moderator details.
	Snippet *ModeratorSnippet `json:"snippet,omitempty"`
}

// ModeratorSnippet contains moderator details.
type ModeratorSnippet struct {
	// LiveChatID is the ID of the live chat.
	LiveChatID string `json:"liveChatId,omitempty"`

	// ModeratorDetails contains information about the moderator.
	ModeratorDetails *ModeratorDetails `json:"moderatorDetails,omitempty"`
}

// ModeratorDetails contains information about a moderator.
type ModeratorDetails struct {
	// ChannelID is the moderator's channel ID.
	ChannelID string `json:"channelId,omitempty"`

	// ChannelURL is the moderator's channel URL.
	ChannelURL string `json:"channelUrl,omitempty"`

	// DisplayName is the moderator's display name.
	DisplayName string `json:"displayName,omitempty"`

	// ProfileImageURL is the moderator's profile image URL.
	ProfileImageURL string `json:"profileImageUrl,omitempty"`
}

// InsertMessageRequest is the request body for sending a chat message.
type InsertMessageRequest struct {
	Snippet *InsertMessageSnippet `json:"snippet"`
}

// InsertMessageSnippet contains the message to send.
type InsertMessageSnippet struct {
	// LiveChatID is the ID of the live chat to send to.
	LiveChatID string `json:"liveChatId"`

	// Type is always "textMessageEvent" for sending messages.
	Type string `json:"type"`

	// TextMessageDetails contains the message text.
	TextMessageDetails *TextMessageDetails `json:"textMessageDetails"`
}

// InsertBanRequest is the request body for banning a user.
type InsertBanRequest struct {
	Snippet *InsertBanSnippet `json:"snippet"`
}

// InsertBanSnippet contains the ban details.
type InsertBanSnippet struct {
	// LiveChatID is the ID of the live chat.
	LiveChatID string `json:"liveChatId"`

	// BanType is "permanent" or "temporary".
	Type string `json:"type"`

	// BanDurationSeconds is required for temporary bans.
	BanDurationSeconds int64 `json:"banDurationSeconds,omitempty"`

	// BannedUserDetails identifies the user to ban.
	BannedUserDetails *BannedUserDetails `json:"bannedUserDetails"`
}

// InsertModeratorRequest is the request body for adding a moderator.
type InsertModeratorRequest struct {
	Snippet *InsertModeratorSnippet `json:"snippet"`
}

// InsertModeratorSnippet contains the moderator details.
type InsertModeratorSnippet struct {
	// LiveChatID is the ID of the live chat.
	LiveChatID string `json:"liveChatId"`

	// ModeratorDetails identifies the user to make a moderator.
	ModeratorDetails *ModeratorDetails `json:"moderatorDetails"`
}

// Clone creates a deep copy of the LiveChatMessage.
// Returns nil if the receiver is nil.
func (m *LiveChatMessage) Clone() *LiveChatMessage {
	if m == nil {
		return nil
	}
	clone := *m
	clone.Snippet = cloneMessageSnippet(m.Snippet)
	clone.AuthorDetails = cloneAuthorDetails(m.AuthorDetails)
	return &clone
}

func cloneMessageSnippet(snippet *MessageSnippet) *MessageSnippet {
	if snippet == nil {
		return nil
	}
	clone := *snippet
	clone.TextMessageDetails = cloneTextMessageDetails(snippet.TextMessageDetails)
	clone.SuperChatDetails = cloneSuperChatDetails(snippet.SuperChatDetails)
	clone.SuperStickerDetails = cloneSuperStickerDetails(snippet.SuperStickerDetails)
	clone.MemberMilestoneChatDetails = cloneMemberMilestoneChatDetails(snippet.MemberMilestoneChatDetails)
	clone.NewSponsorDetails = cloneNewSponsorDetails(snippet.NewSponsorDetails)
	clone.MembershipGiftingDetails = cloneMembershipGiftingDetails(snippet.MembershipGiftingDetails)
	clone.GiftMembershipReceivedDetails = cloneGiftMembershipReceivedDetails(snippet.GiftMembershipReceivedDetails)
	clone.MessageDeletedDetails = cloneMessageDeletedDetails(snippet.MessageDeletedDetails)
	clone.UserBannedDetails = cloneUserBannedDetails(snippet.UserBannedDetails)
	clone.PollDetails = clonePollDetails(snippet.PollDetails)
	return &clone
}

func cloneAuthorDetails(details *AuthorDetails) *AuthorDetails {
	if details == nil {
		return nil
	}
	clone := *details
	return &clone
}

func cloneTextMessageDetails(details *TextMessageDetails) *TextMessageDetails {
	if details == nil {
		return nil
	}
	clone := *details
	return &clone
}

func cloneSuperChatDetails(details *SuperChatDetails) *SuperChatDetails {
	if details == nil {
		return nil
	}
	clone := *details
	return &clone
}

func cloneSuperStickerDetails(details *SuperStickerDetails) *SuperStickerDetails {
	if details == nil {
		return nil
	}
	clone := *details
	clone.SuperStickerMetadata = cloneSuperStickerMetadata(details.SuperStickerMetadata)
	return &clone
}

func cloneSuperStickerMetadata(metadata *SuperStickerMetadata) *SuperStickerMetadata {
	if metadata == nil {
		return nil
	}
	clone := *metadata
	return &clone
}

func cloneMemberMilestoneChatDetails(details *MemberMilestoneChatDetails) *MemberMilestoneChatDetails {
	if details == nil {
		return nil
	}
	clone := *details
	return &clone
}

func cloneNewSponsorDetails(details *NewSponsorDetails) *NewSponsorDetails {
	if details == nil {
		return nil
	}
	clone := *details
	return &clone
}

func cloneMembershipGiftingDetails(details *MembershipGiftingDetails) *MembershipGiftingDetails {
	if details == nil {
		return nil
	}
	clone := *details
	return &clone
}

func cloneGiftMembershipReceivedDetails(details *GiftMembershipReceivedDetails) *GiftMembershipReceivedDetails {
	if details == nil {
		return nil
	}
	clone := *details
	return &clone
}

func cloneMessageDeletedDetails(details *MessageDeletedDetails) *MessageDeletedDetails {
	if details == nil {
		return nil
	}
	clone := *details
	return &clone
}

func cloneUserBannedDetails(details *UserBannedDetails) *UserBannedDetails {
	if details == nil {
		return nil
	}
	clone := *details
	clone.BannedUserDetails = cloneBannedUserDetails(details.BannedUserDetails)
	return &clone
}

func cloneBannedUserDetails(details *BannedUserDetails) *BannedUserDetails {
	if details == nil {
		return nil
	}
	clone := *details
	return &clone
}

func clonePollDetails(details *PollDetails) *PollDetails {
	if details == nil {
		return nil
	}
	clone := *details
	if len(details.Choices) > 0 {
		choices := make([]PollChoice, len(details.Choices))
		copy(choices, details.Choices)
		clone.Choices = choices
	}
	return &clone
}

// LiveChatModeratorListResponse is the response from liveChatModerators.list.
type LiveChatModeratorListResponse struct {
	// Kind identifies the resource type.
	Kind string `json:"kind,omitempty"`

	// ETag for caching.
	ETag string `json:"etag,omitempty"`

	// NextPageToken for pagination.
	NextPageToken string `json:"nextPageToken,omitempty"`

	// PageInfo contains pagination metadata.
	PageInfo *PageInfo `json:"pageInfo,omitempty"`

	// Items contains the moderators.
	Items []*LiveChatModerator `json:"items,omitempty"`
}

// SuperChatEventResource represents a Super Chat or Super Sticker event from the API.
// This is the raw API resource type; for the high-level event used by ChatBotClient,
// see SuperChatEvent in chat.go.
type SuperChatEventResource struct {
	// Kind identifies the resource type.
	Kind string `json:"kind,omitempty"`

	// ETag for caching.
	ETag string `json:"etag,omitempty"`

	// ID is the unique identifier for this event.
	ID string `json:"id,omitempty"`

	// Snippet contains event metadata.
	Snippet *SuperChatEventResourceSnippet `json:"snippet,omitempty"`
}

// SuperChatEventResourceSnippet contains Super Chat event details from the API.
type SuperChatEventResourceSnippet struct {
	// ChannelID is the channel that received the Super Chat.
	ChannelID string `json:"channelId,omitempty"`

	// SupporterDetails contains information about the supporter.
	SupporterDetails *SupporterDetails `json:"supporterDetails,omitempty"`

	// CommentText is the optional message from the supporter.
	CommentText string `json:"commentText,omitempty"`

	// CreatedAt is when the Super Chat was created.
	CreatedAt time.Time `json:"createdAt,omitempty"`

	// AmountMicros is the donation amount in micros.
	AmountMicros int64 `json:"amountMicros,omitempty,string"`

	// Currency is the ISO 4217 currency code.
	Currency string `json:"currency,omitempty"`

	// DisplayString is the formatted donation amount.
	DisplayString string `json:"displayString,omitempty"`

	// MessageType indicates if this is a Super Chat or Super Sticker.
	// Values: "superChatEvent" or "superStickerEvent"
	MessageType string `json:"messageType,omitempty"`

	// SuperStickerMetadata is present for Super Sticker events.
	SuperStickerMetadata *SuperStickerMetadata `json:"superStickerMetadata,omitempty"`

	// IsSuperStickerEvent returns true if this is a Super Sticker.
	IsSuperStickerEvent bool `json:"isSuperStickerEvent,omitempty"`
}

// SupporterDetails contains information about a Super Chat supporter.
type SupporterDetails struct {
	// ChannelID is the supporter's channel ID.
	ChannelID string `json:"channelId,omitempty"`

	// ChannelURL is the supporter's channel URL.
	ChannelURL string `json:"channelUrl,omitempty"`

	// DisplayName is the supporter's display name.
	DisplayName string `json:"displayName,omitempty"`

	// ProfileImageURL is the supporter's profile image URL.
	ProfileImageURL string `json:"profileImageUrl,omitempty"`
}

// SuperChatEventResourceListResponse is the response from superChatEvents.list.
type SuperChatEventResourceListResponse struct {
	// Kind identifies the resource type.
	Kind string `json:"kind,omitempty"`

	// ETag for caching.
	ETag string `json:"etag,omitempty"`

	// NextPageToken for pagination.
	NextPageToken string `json:"nextPageToken,omitempty"`

	// PageInfo contains pagination metadata.
	PageInfo *PageInfo `json:"pageInfo,omitempty"`

	// Items contains the Super Chat events.
	Items []*SuperChatEventResource `json:"items,omitempty"`
}

// Cuepoint represents an ad break cuepoint in a live broadcast.
type Cuepoint struct {
	// Kind identifies the resource type.
	Kind string `json:"kind,omitempty"`

	// ETag for caching.
	ETag string `json:"etag,omitempty"`

	// ID is the unique identifier for this cuepoint.
	ID string `json:"id,omitempty"`

	// CueType indicates the type of cuepoint.
	// Value is always "cueTypeAd" for ad breaks.
	CueType string `json:"cueType,omitempty"`

	// DurationSecs is the ad break duration in seconds.
	// Default is 30 seconds if not specified.
	DurationSecs int `json:"durationSecs,omitempty"`

	// InsertionOffsetTimeMs is when the cuepoint should be inserted.
	// Use -1 for immediate insertion.
	InsertionOffsetTimeMs int64 `json:"insertionOffsetTimeMs,omitempty,string"`

	// WalltimeMs is the wall clock time when the cuepoint should be inserted.
	// Alternative to InsertionOffsetTimeMs.
	WalltimeMs int64 `json:"walltimeMs,omitempty,string"`
}

// CuepointRequest is the request body for inserting a cuepoint.
type CuepointRequest struct {
	// CueType is always "cueTypeAd" for ad breaks.
	CueType string `json:"cueType"`

	// DurationSecs is the ad break duration (default 30).
	DurationSecs int `json:"durationSecs,omitempty"`

	// InsertionOffsetTimeMs is when to insert (-1 for immediate).
	InsertionOffsetTimeMs int64 `json:"insertionOffsetTimeMs,omitempty,string"`

	// WalltimeMs is alternative wall clock time for insertion.
	WalltimeMs int64 `json:"walltimeMs,omitempty,string"`
}

// CueType constants.
const (
	CueTypeAd = "cueTypeAd"
)

// Chat mode constants for liveChatMessages.transition.
const (
	// ChatModeSubscribersOnly restricts chat to channel subscribers.
	ChatModeSubscribersOnly = "subscribersOnlyMode"

	// ChatModeMembersOnly restricts chat to channel members.
	ChatModeMembersOnly = "membersOnlyMode"

	// ChatModeSlowMode enables slow mode with a delay between messages.
	ChatModeSlowMode = "slowMode"

	// ChatModeNormal is normal chat mode with no restrictions.
	ChatModeNormal = "normal"
)

// TransitionChatModeRequest is the request body for changing chat mode.
type TransitionChatModeRequest struct {
	Snippet *TransitionChatModeSnippet `json:"snippet"`
}

// TransitionChatModeSnippet contains chat mode transition details.
type TransitionChatModeSnippet struct {
	// LiveChatID is the ID of the live chat.
	LiveChatID string `json:"liveChatId"`

	// Type specifies the new chat mode.
	Type string `json:"type"`

	// SlowModeDelayMs is the delay between messages in slow mode (milliseconds).
	// Only used when Type is ChatModeSlowMode.
	SlowModeDelayMs int64 `json:"slowModeDelayMs,omitempty"`
}
