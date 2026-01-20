package streaming

import (
	"context"
	"fmt"
	"slices"
	"sync"
	"time"

	"github.com/Its-donkey/yougopher/youtube/core"
)

// TokenProvider provides access tokens for authentication.
// This interface allows ChatBotClient to work with any auth implementation.
type TokenProvider interface {
	// AccessToken returns a valid access token, refreshing if necessary.
	AccessToken(ctx context.Context) (string, error)
}

// ChatMessage represents a parsed chat message with author information.
type ChatMessage struct {
	// ID is the unique message identifier.
	ID string

	// Message is the display text of the message.
	Message string

	// Author contains information about the message author.
	Author *Author

	// PublishedAt is when the message was sent.
	// May be zero if the API doesn't provide a publish time.
	PublishedAt time.Time

	// Raw is the underlying LiveChatMessage if needed.
	Raw *LiveChatMessage
}

// Author represents the author of a chat message.
type Author struct {
	// ChannelID is the YouTube channel ID.
	ChannelID string

	// DisplayName is the author's display name.
	DisplayName string

	// ProfileImageURL is the URL to the author's profile image.
	ProfileImageURL string

	// IsModerator indicates if the author is a chat moderator.
	IsModerator bool

	// IsOwner indicates if the author owns the channel.
	IsOwner bool

	// IsMember indicates if the author is a channel member.
	IsMember bool

	// IsVerified indicates if the author is verified.
	IsVerified bool
}

// SuperChatEvent represents a Super Chat donation.
type SuperChatEvent struct {
	// ID is the unique event identifier.
	ID string

	// Author is the user who sent the Super Chat.
	Author *Author

	// Message is the optional message from the donor.
	Message string

	// Amount is the formatted donation amount (e.g., "$5.00").
	Amount string

	// AmountMicros is the donation amount in micros.
	AmountMicros int64

	// Currency is the ISO 4217 currency code.
	Currency string

	// Tier is the Super Chat tier (1-7).
	Tier int

	// Raw is the underlying LiveChatMessage.
	Raw *LiveChatMessage
}

// SuperStickerEvent represents a Super Sticker.
type SuperStickerEvent struct {
	// ID is the unique event identifier.
	ID string

	// Author is the user who sent the Super Sticker.
	Author *Author

	// StickerID is the unique sticker identifier.
	StickerID string

	// AltText is alternative text describing the sticker.
	AltText string

	// Amount is the formatted sticker cost.
	Amount string

	// AmountMicros is the sticker cost in micros.
	AmountMicros int64

	// Currency is the ISO 4217 currency code.
	Currency string

	// Tier is the Super Sticker tier.
	Tier int

	// Raw is the underlying LiveChatMessage.
	Raw *LiveChatMessage
}

// MembershipEvent represents a new channel membership.
type MembershipEvent struct {
	// ID is the unique event identifier.
	ID string

	// Author is the new member.
	Author *Author

	// LevelName is the membership level name.
	LevelName string

	// IsUpgrade indicates if this is an upgrade from a lower tier.
	IsUpgrade bool

	// Raw is the underlying LiveChatMessage.
	Raw *LiveChatMessage
}

// MemberMilestoneEvent represents a member milestone message.
type MemberMilestoneEvent struct {
	// ID is the unique event identifier.
	ID string

	// Author is the member celebrating the milestone.
	Author *Author

	// Message is the optional milestone message.
	Message string

	// LevelName is the membership level name.
	LevelName string

	// Months is the number of months as a member.
	Months int

	// Raw is the underlying LiveChatMessage.
	Raw *LiveChatMessage
}

// GiftMembershipEvent represents gifted memberships.
type GiftMembershipEvent struct {
	// ID is the unique event identifier.
	ID string

	// Author is the user gifting memberships.
	Author *Author

	// LevelName is the membership level being gifted.
	LevelName string

	// Count is the number of memberships gifted.
	Count int

	// Raw is the underlying LiveChatMessage.
	Raw *LiveChatMessage
}

// GiftMembershipReceivedEvent represents a received gifted membership.
type GiftMembershipReceivedEvent struct {
	// ID is the unique event identifier.
	ID string

	// Author is the user who received the gifted membership.
	Author *Author

	// LevelName is the membership level received.
	LevelName string

	// GifterChannelID is the channel ID of the gifter.
	GifterChannelID string

	// AssociatedGiftingMessageID links to the gifting event.
	AssociatedGiftingMessageID string

	// Raw is the underlying LiveChatMessage.
	Raw *LiveChatMessage
}

// BanEvent represents a user ban in chat.
type BanEvent struct {
	// BannedUser contains information about the banned user.
	BannedUser *Author

	// BanType is "permanent" or "temporary".
	BanType string

	// Duration is the ban duration for temporary bans.
	Duration time.Duration

	// Raw is the underlying UserBannedDetails.
	Raw *UserBannedDetails
}

// Handler wrapper types for pointer identity.
type (
	chatMessageHandler            struct{ fn func(*ChatMessage) }
	superChatHandler              struct{ fn func(*SuperChatEvent) }
	superStickerHandler           struct{ fn func(*SuperStickerEvent) }
	membershipHandler             struct{ fn func(*MembershipEvent) }
	memberMilestoneHandler        struct{ fn func(*MemberMilestoneEvent) }
	giftMembershipHandler         struct{ fn func(*GiftMembershipEvent) }
	giftMembershipReceivedHandler struct {
		fn func(*GiftMembershipReceivedEvent)
	}
	messageDeletedHandler struct{ fn func(string) }
	userBannedHandler     struct{ fn func(*BanEvent) }
	chatConnectHandler    struct{ fn func() }
	chatDisconnectHandler struct{ fn func() }
	chatErrorHandler      struct{ fn func(error) }
)

// ChatBotClient is a high-level client for building YouTube chat bots.
type ChatBotClient struct {
	poller        *LiveChatPoller
	tokenProvider TokenProvider
	client        *core.Client
	liveChatID    string

	// Composable event handlers
	mu                             sync.RWMutex
	messageHandlers                []*chatMessageHandler
	superChatHandlers              []*superChatHandler
	superStickerHandlers           []*superStickerHandler
	membershipHandlers             []*membershipHandler
	memberMilestoneHandlers        []*memberMilestoneHandler
	giftMembershipHandlers         []*giftMembershipHandler
	giftMembershipReceivedHandlers []*giftMembershipReceivedHandler
	messageDeletedHandlers         []*messageDeletedHandler
	userBannedHandlers             []*userBannedHandler
	connectHandlers                []*chatConnectHandler
	disconnectHandlers             []*chatDisconnectHandler
	errorHandlers                  []*chatErrorHandler

	// Internal state
	pollerUnsub      func()        // Unsubscribe from poller events
	tokenRefreshStop chan struct{} // Signal to stop token refresh loop
	tokenRefreshDone chan struct{} // Token refresh loop completed
	refreshInterval  time.Duration // How often to refresh token (default 45 minutes)
}

// ChatBotOption configures a ChatBotClient.
type ChatBotOption func(*ChatBotClient)

// DefaultTokenRefreshInterval is the default interval for refreshing the access token.
const DefaultTokenRefreshInterval = 45 * time.Minute

// NewChatBotClient creates a new high-level chat bot client.
// The tokenProvider can be nil if the client already has an access token set.
// Returns an error if client is nil or liveChatID is empty.
func NewChatBotClient(client *core.Client, tokenProvider TokenProvider, liveChatID string, opts ...ChatBotOption) (*ChatBotClient, error) {
	if client == nil {
		return nil, fmt.Errorf("client cannot be nil")
	}
	if liveChatID == "" {
		return nil, fmt.Errorf("liveChatID cannot be empty")
	}

	c := &ChatBotClient{
		client:          client,
		tokenProvider:   tokenProvider,
		liveChatID:      liveChatID,
		refreshInterval: DefaultTokenRefreshInterval,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c, nil
}

// WithPoller sets a custom LiveChatPoller (useful for testing).
func WithPoller(poller *LiveChatPoller) ChatBotOption {
	return func(c *ChatBotClient) { c.poller = poller }
}

// WithTokenRefreshInterval sets the interval for refreshing the access token.
// Default is 45 minutes. Set to 0 to disable auto-refresh.
func WithTokenRefreshInterval(d time.Duration) ChatBotOption {
	return func(c *ChatBotClient) { c.refreshInterval = d }
}

// Connect starts the chat bot and begins listening for messages.
func (c *ChatBotClient) Connect(ctx context.Context) error {
	// Update access token from token provider
	if c.tokenProvider != nil {
		token, err := c.tokenProvider.AccessToken(ctx)
		if err != nil {
			return err
		}
		c.client.SetAccessToken(token)
	}

	// Create poller if not provided
	if c.poller == nil {
		c.poller = NewLiveChatPoller(c.client, c.liveChatID)
	}

	// Clean up any existing subscription to prevent handler duplication
	if c.pollerUnsub != nil {
		c.pollerUnsub()
		c.pollerUnsub = nil
	}

	// Stop any existing token refresh loop
	c.stopTokenRefresh()

	// Subscribe to poller events
	c.subscribeToPoller()

	// Start token refresh loop if we have a token provider and refresh interval
	if c.tokenProvider != nil && c.refreshInterval > 0 {
		c.tokenRefreshStop = make(chan struct{})
		c.tokenRefreshDone = make(chan struct{})
		go c.tokenRefreshLoop(ctx)
	}

	// Start polling
	return c.poller.Start(ctx)
}

// Close stops the chat bot.
func (c *ChatBotClient) Close() error {
	// Stop token refresh loop first
	c.stopTokenRefresh()

	// Stop poller (this will trigger disconnect handlers)
	if c.poller != nil {
		c.poller.Stop()
	}
	// Then unsubscribe from poller events
	if c.pollerUnsub != nil {
		c.pollerUnsub()
		c.pollerUnsub = nil
	}
	return nil
}

// stopTokenRefresh stops the token refresh loop if running.
func (c *ChatBotClient) stopTokenRefresh() {
	if c.tokenRefreshStop != nil {
		close(c.tokenRefreshStop)
		<-c.tokenRefreshDone
		c.tokenRefreshStop = nil
		c.tokenRefreshDone = nil
	}
}

// tokenRefreshLoop periodically refreshes the access token.
func (c *ChatBotClient) tokenRefreshLoop(ctx context.Context) {
	defer close(c.tokenRefreshDone)

	ticker := time.NewTicker(c.refreshInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-c.tokenRefreshStop:
			return
		case <-ticker.C:
			token, err := c.tokenProvider.AccessToken(ctx)
			if err != nil {
				c.dispatchError(fmt.Errorf("token refresh failed: %w", err))
				continue
			}
			c.client.SetAccessToken(token)
		}
	}
}

// IsConnected returns true if the bot is currently connected.
func (c *ChatBotClient) IsConnected() bool {
	return c.poller != nil && c.poller.IsRunning()
}

// subscribeToPoller registers handlers on the underlying poller.
func (c *ChatBotClient) subscribeToPoller() {
	var unsubs []func()

	// Message handler - dispatches to semantic handlers
	unsubs = append(unsubs, c.poller.OnMessage(func(msg *LiveChatMessage) {
		c.handleMessage(msg)
	}))

	// Delete handler
	unsubs = append(unsubs, c.poller.OnDelete(func(id string) {
		c.dispatchMessageDeleted(id)
	}))

	// Ban handler
	unsubs = append(unsubs, c.poller.OnBan(func(details *UserBannedDetails) {
		c.dispatchUserBanned(details)
	}))

	// Error handler
	unsubs = append(unsubs, c.poller.OnError(func(err error) {
		c.dispatchError(err)
	}))

	// Connect handler
	unsubs = append(unsubs, c.poller.OnConnect(func() {
		c.dispatchConnect()
	}))

	// Disconnect handler
	unsubs = append(unsubs, c.poller.OnDisconnect(func() {
		c.dispatchDisconnect()
	}))

	// Store combined unsubscribe function
	c.pollerUnsub = func() {
		for _, unsub := range unsubs {
			unsub()
		}
	}
}

// handleMessage routes a message to the appropriate semantic handler.
func (c *ChatBotClient) handleMessage(msg *LiveChatMessage) {
	if msg.Snippet == nil {
		return
	}

	switch msg.Snippet.Type {
	case MessageTypeText:
		c.dispatchChatMessage(msg)
	case MessageTypeSuperChat:
		c.dispatchSuperChat(msg)
	case MessageTypeSuperSticker:
		c.dispatchSuperSticker(msg)
	case MessageTypeMembership:
		c.dispatchMembership(msg)
	case MessageTypeMemberMilestone:
		c.dispatchMemberMilestone(msg)
	case MessageTypeMembershipGifting:
		c.dispatchGiftMembership(msg)
	case MessageTypeGiftMembershipReceived:
		c.dispatchGiftMembershipReceived(msg)
	}
}

// Say sends a message to the chat.
func (c *ChatBotClient) Say(ctx context.Context, message string) error {
	if !c.IsConnected() {
		return ErrNotRunning
	}
	_, err := c.poller.SendMessage(ctx, message)
	return err
}

// Delete deletes a message from the chat.
func (c *ChatBotClient) Delete(ctx context.Context, messageID string) error {
	if !c.IsConnected() {
		return ErrNotRunning
	}
	return c.poller.DeleteMessage(ctx, messageID)
}

// Ban permanently bans a user from the chat.
func (c *ChatBotClient) Ban(ctx context.Context, channelID string) error {
	if !c.IsConnected() {
		return ErrNotRunning
	}
	_, err := c.poller.BanUser(ctx, channelID)
	return err
}

// Timeout temporarily bans a user from the chat.
func (c *ChatBotClient) Timeout(ctx context.Context, channelID string, seconds int) error {
	if !c.IsConnected() {
		return ErrNotRunning
	}
	if seconds <= 0 {
		return fmt.Errorf("timeout duration must be positive")
	}
	_, err := c.poller.TimeoutUser(ctx, channelID, int64(seconds))
	return err
}

// Unban removes a ban from the chat.
func (c *ChatBotClient) Unban(ctx context.Context, banID string) error {
	if !c.IsConnected() {
		return ErrNotRunning
	}
	return c.poller.UnbanUser(ctx, banID)
}

// AddModerator adds a moderator to the chat.
func (c *ChatBotClient) AddModerator(ctx context.Context, channelID string) error {
	if !c.IsConnected() {
		return ErrNotRunning
	}
	_, err := c.poller.AddModerator(ctx, channelID)
	return err
}

// RemoveModerator removes a moderator from the chat.
func (c *ChatBotClient) RemoveModerator(ctx context.Context, moderatorID string) error {
	if !c.IsConnected() {
		return ErrNotRunning
	}
	return c.poller.RemoveModerator(ctx, moderatorID)
}

// OnMessage registers a handler for chat messages.
func (c *ChatBotClient) OnMessage(fn func(*ChatMessage)) func() {
	c.mu.Lock()
	defer c.mu.Unlock()

	h := &chatMessageHandler{fn: fn}
	c.messageHandlers = append(c.messageHandlers, h)

	var once sync.Once
	return func() {
		once.Do(func() {
			c.mu.Lock()
			defer c.mu.Unlock()
			for i, handler := range c.messageHandlers {
				if handler == h {
					c.messageHandlers = slices.Delete(c.messageHandlers, i, i+1)
					return
				}
			}
		})
	}
}

// OnSuperChat registers a handler for Super Chat events.
func (c *ChatBotClient) OnSuperChat(fn func(*SuperChatEvent)) func() {
	c.mu.Lock()
	defer c.mu.Unlock()

	h := &superChatHandler{fn: fn}
	c.superChatHandlers = append(c.superChatHandlers, h)

	var once sync.Once
	return func() {
		once.Do(func() {
			c.mu.Lock()
			defer c.mu.Unlock()
			for i, handler := range c.superChatHandlers {
				if handler == h {
					c.superChatHandlers = slices.Delete(c.superChatHandlers, i, i+1)
					return
				}
			}
		})
	}
}

// OnSuperSticker registers a handler for Super Sticker events.
func (c *ChatBotClient) OnSuperSticker(fn func(*SuperStickerEvent)) func() {
	c.mu.Lock()
	defer c.mu.Unlock()

	h := &superStickerHandler{fn: fn}
	c.superStickerHandlers = append(c.superStickerHandlers, h)

	var once sync.Once
	return func() {
		once.Do(func() {
			c.mu.Lock()
			defer c.mu.Unlock()
			for i, handler := range c.superStickerHandlers {
				if handler == h {
					c.superStickerHandlers = slices.Delete(c.superStickerHandlers, i, i+1)
					return
				}
			}
		})
	}
}

// OnMembership registers a handler for new membership events.
func (c *ChatBotClient) OnMembership(fn func(*MembershipEvent)) func() {
	c.mu.Lock()
	defer c.mu.Unlock()

	h := &membershipHandler{fn: fn}
	c.membershipHandlers = append(c.membershipHandlers, h)

	var once sync.Once
	return func() {
		once.Do(func() {
			c.mu.Lock()
			defer c.mu.Unlock()
			for i, handler := range c.membershipHandlers {
				if handler == h {
					c.membershipHandlers = slices.Delete(c.membershipHandlers, i, i+1)
					return
				}
			}
		})
	}
}

// OnMemberMilestone registers a handler for member milestone events.
func (c *ChatBotClient) OnMemberMilestone(fn func(*MemberMilestoneEvent)) func() {
	c.mu.Lock()
	defer c.mu.Unlock()

	h := &memberMilestoneHandler{fn: fn}
	c.memberMilestoneHandlers = append(c.memberMilestoneHandlers, h)

	var once sync.Once
	return func() {
		once.Do(func() {
			c.mu.Lock()
			defer c.mu.Unlock()
			for i, handler := range c.memberMilestoneHandlers {
				if handler == h {
					c.memberMilestoneHandlers = slices.Delete(c.memberMilestoneHandlers, i, i+1)
					return
				}
			}
		})
	}
}

// OnGiftMembership registers a handler for gift membership events.
func (c *ChatBotClient) OnGiftMembership(fn func(*GiftMembershipEvent)) func() {
	c.mu.Lock()
	defer c.mu.Unlock()

	h := &giftMembershipHandler{fn: fn}
	c.giftMembershipHandlers = append(c.giftMembershipHandlers, h)

	var once sync.Once
	return func() {
		once.Do(func() {
			c.mu.Lock()
			defer c.mu.Unlock()
			for i, handler := range c.giftMembershipHandlers {
				if handler == h {
					c.giftMembershipHandlers = slices.Delete(c.giftMembershipHandlers, i, i+1)
					return
				}
			}
		})
	}
}

// OnGiftMembershipReceived registers a handler for received gift membership events.
func (c *ChatBotClient) OnGiftMembershipReceived(fn func(*GiftMembershipReceivedEvent)) func() {
	c.mu.Lock()
	defer c.mu.Unlock()

	h := &giftMembershipReceivedHandler{fn: fn}
	c.giftMembershipReceivedHandlers = append(c.giftMembershipReceivedHandlers, h)

	var once sync.Once
	return func() {
		once.Do(func() {
			c.mu.Lock()
			defer c.mu.Unlock()
			for i, handler := range c.giftMembershipReceivedHandlers {
				if handler == h {
					c.giftMembershipReceivedHandlers = slices.Delete(c.giftMembershipReceivedHandlers, i, i+1)
					return
				}
			}
		})
	}
}

// OnMessageDeleted registers a handler for message deletion events.
func (c *ChatBotClient) OnMessageDeleted(fn func(string)) func() {
	c.mu.Lock()
	defer c.mu.Unlock()

	h := &messageDeletedHandler{fn: fn}
	c.messageDeletedHandlers = append(c.messageDeletedHandlers, h)

	var once sync.Once
	return func() {
		once.Do(func() {
			c.mu.Lock()
			defer c.mu.Unlock()
			for i, handler := range c.messageDeletedHandlers {
				if handler == h {
					c.messageDeletedHandlers = slices.Delete(c.messageDeletedHandlers, i, i+1)
					return
				}
			}
		})
	}
}

// OnUserBanned registers a handler for user ban events.
func (c *ChatBotClient) OnUserBanned(fn func(*BanEvent)) func() {
	c.mu.Lock()
	defer c.mu.Unlock()

	h := &userBannedHandler{fn: fn}
	c.userBannedHandlers = append(c.userBannedHandlers, h)

	var once sync.Once
	return func() {
		once.Do(func() {
			c.mu.Lock()
			defer c.mu.Unlock()
			for i, handler := range c.userBannedHandlers {
				if handler == h {
					c.userBannedHandlers = slices.Delete(c.userBannedHandlers, i, i+1)
					return
				}
			}
		})
	}
}

// OnConnect registers a handler called when the bot connects.
func (c *ChatBotClient) OnConnect(fn func()) func() {
	c.mu.Lock()
	defer c.mu.Unlock()

	h := &chatConnectHandler{fn: fn}
	c.connectHandlers = append(c.connectHandlers, h)

	var once sync.Once
	return func() {
		once.Do(func() {
			c.mu.Lock()
			defer c.mu.Unlock()
			for i, handler := range c.connectHandlers {
				if handler == h {
					c.connectHandlers = slices.Delete(c.connectHandlers, i, i+1)
					return
				}
			}
		})
	}
}

// OnDisconnect registers a handler called when the bot disconnects.
func (c *ChatBotClient) OnDisconnect(fn func()) func() {
	c.mu.Lock()
	defer c.mu.Unlock()

	h := &chatDisconnectHandler{fn: fn}
	c.disconnectHandlers = append(c.disconnectHandlers, h)

	var once sync.Once
	return func() {
		once.Do(func() {
			c.mu.Lock()
			defer c.mu.Unlock()
			for i, handler := range c.disconnectHandlers {
				if handler == h {
					c.disconnectHandlers = slices.Delete(c.disconnectHandlers, i, i+1)
					return
				}
			}
		})
	}
}

// OnError registers a handler for errors.
func (c *ChatBotClient) OnError(fn func(error)) func() {
	c.mu.Lock()
	defer c.mu.Unlock()

	h := &chatErrorHandler{fn: fn}
	c.errorHandlers = append(c.errorHandlers, h)

	var once sync.Once
	return func() {
		once.Do(func() {
			c.mu.Lock()
			defer c.mu.Unlock()
			for i, handler := range c.errorHandlers {
				if handler == h {
					c.errorHandlers = slices.Delete(c.errorHandlers, i, i+1)
					return
				}
			}
		})
	}
}

// Dispatch methods

func (c *ChatBotClient) dispatchChatMessage(msg *LiveChatMessage) {
	c.mu.RLock()
	handlers := make([]*chatMessageHandler, len(c.messageHandlers))
	copy(handlers, c.messageHandlers)
	c.mu.RUnlock()

	chatMsg := &ChatMessage{
		ID:          msg.ID,
		Message:     msg.Message(),
		Author:      parseAuthor(msg.AuthorDetails),
		PublishedAt: msg.Snippet.PublishedAt,
		Raw:         msg,
	}

	for _, h := range handlers {
		c.safeCall(func() { h.fn(chatMsg) })
	}
}

func (c *ChatBotClient) dispatchSuperChat(msg *LiveChatMessage) {
	if msg.Snippet.SuperChatDetails == nil {
		return
	}

	c.mu.RLock()
	handlers := make([]*superChatHandler, len(c.superChatHandlers))
	copy(handlers, c.superChatHandlers)
	c.mu.RUnlock()

	sc := msg.Snippet.SuperChatDetails
	event := &SuperChatEvent{
		ID:           msg.ID,
		Author:       parseAuthor(msg.AuthorDetails),
		Message:      sc.UserComment,
		Amount:       sc.AmountDisplayString,
		AmountMicros: sc.AmountMicros,
		Currency:     sc.Currency,
		Tier:         sc.Tier,
		Raw:          msg,
	}

	for _, h := range handlers {
		c.safeCall(func() { h.fn(event) })
	}
}

func (c *ChatBotClient) dispatchSuperSticker(msg *LiveChatMessage) {
	if msg.Snippet.SuperStickerDetails == nil {
		return
	}

	c.mu.RLock()
	handlers := make([]*superStickerHandler, len(c.superStickerHandlers))
	copy(handlers, c.superStickerHandlers)
	c.mu.RUnlock()

	ss := msg.Snippet.SuperStickerDetails
	event := &SuperStickerEvent{
		ID:           msg.ID,
		Author:       parseAuthor(msg.AuthorDetails),
		StickerID:    ss.SuperStickerID,
		Amount:       ss.AmountDisplayString,
		AmountMicros: ss.AmountMicros,
		Currency:     ss.Currency,
		Tier:         ss.Tier,
		Raw:          msg,
	}

	if ss.SuperStickerMetadata != nil {
		event.AltText = ss.SuperStickerMetadata.AltText
	}

	for _, h := range handlers {
		c.safeCall(func() { h.fn(event) })
	}
}

func (c *ChatBotClient) dispatchMembership(msg *LiveChatMessage) {
	if msg.Snippet.NewSponsorDetails == nil {
		return
	}

	c.mu.RLock()
	handlers := make([]*membershipHandler, len(c.membershipHandlers))
	copy(handlers, c.membershipHandlers)
	c.mu.RUnlock()

	ns := msg.Snippet.NewSponsorDetails
	event := &MembershipEvent{
		ID:        msg.ID,
		Author:    parseAuthor(msg.AuthorDetails),
		LevelName: ns.MemberLevelName,
		IsUpgrade: ns.IsUpgrade,
		Raw:       msg,
	}

	for _, h := range handlers {
		c.safeCall(func() { h.fn(event) })
	}
}

func (c *ChatBotClient) dispatchMemberMilestone(msg *LiveChatMessage) {
	if msg.Snippet.MemberMilestoneChatDetails == nil {
		return
	}

	c.mu.RLock()
	handlers := make([]*memberMilestoneHandler, len(c.memberMilestoneHandlers))
	copy(handlers, c.memberMilestoneHandlers)
	c.mu.RUnlock()

	ms := msg.Snippet.MemberMilestoneChatDetails
	event := &MemberMilestoneEvent{
		ID:        msg.ID,
		Author:    parseAuthor(msg.AuthorDetails),
		Message:   ms.UserComment,
		LevelName: ms.MemberLevelName,
		Months:    ms.MemberMonth,
		Raw:       msg,
	}

	for _, h := range handlers {
		c.safeCall(func() { h.fn(event) })
	}
}

func (c *ChatBotClient) dispatchGiftMembership(msg *LiveChatMessage) {
	if msg.Snippet.MembershipGiftingDetails == nil {
		return
	}

	c.mu.RLock()
	handlers := make([]*giftMembershipHandler, len(c.giftMembershipHandlers))
	copy(handlers, c.giftMembershipHandlers)
	c.mu.RUnlock()

	gm := msg.Snippet.MembershipGiftingDetails
	event := &GiftMembershipEvent{
		ID:        msg.ID,
		Author:    parseAuthor(msg.AuthorDetails),
		LevelName: gm.MemberLevelName,
		Count:     gm.GiftMembershipsCount,
		Raw:       msg,
	}

	for _, h := range handlers {
		c.safeCall(func() { h.fn(event) })
	}
}

func (c *ChatBotClient) dispatchGiftMembershipReceived(msg *LiveChatMessage) {
	if msg.Snippet.GiftMembershipReceivedDetails == nil {
		return
	}

	c.mu.RLock()
	handlers := make([]*giftMembershipReceivedHandler, len(c.giftMembershipReceivedHandlers))
	copy(handlers, c.giftMembershipReceivedHandlers)
	c.mu.RUnlock()

	gr := msg.Snippet.GiftMembershipReceivedDetails
	event := &GiftMembershipReceivedEvent{
		ID:                         msg.ID,
		Author:                     parseAuthor(msg.AuthorDetails),
		LevelName:                  gr.MemberLevelName,
		GifterChannelID:            gr.GifterChannelID,
		AssociatedGiftingMessageID: gr.AssociatedMembershipGiftingMessageID,
		Raw:                        msg,
	}

	for _, h := range handlers {
		c.safeCall(func() { h.fn(event) })
	}
}

func (c *ChatBotClient) dispatchMessageDeleted(id string) {
	c.mu.RLock()
	handlers := make([]*messageDeletedHandler, len(c.messageDeletedHandlers))
	copy(handlers, c.messageDeletedHandlers)
	c.mu.RUnlock()

	for _, h := range handlers {
		c.safeCall(func() { h.fn(id) })
	}
}

func (c *ChatBotClient) dispatchUserBanned(details *UserBannedDetails) {
	if details == nil {
		return
	}

	c.mu.RLock()
	handlers := make([]*userBannedHandler, len(c.userBannedHandlers))
	copy(handlers, c.userBannedHandlers)
	c.mu.RUnlock()

	event := &BanEvent{
		BanType:  details.BanType,
		Duration: time.Duration(details.BanDurationSeconds) * time.Second,
		Raw:      details,
	}

	if details.BannedUserDetails != nil {
		event.BannedUser = &Author{
			ChannelID:       details.BannedUserDetails.ChannelID,
			DisplayName:     details.BannedUserDetails.DisplayName,
			ProfileImageURL: details.BannedUserDetails.ProfileImageURL,
		}
	}

	for _, h := range handlers {
		c.safeCall(func() { h.fn(event) })
	}
}

func (c *ChatBotClient) dispatchConnect() {
	c.mu.RLock()
	handlers := make([]*chatConnectHandler, len(c.connectHandlers))
	copy(handlers, c.connectHandlers)
	c.mu.RUnlock()

	for _, h := range handlers {
		c.safeCall(func() { h.fn() })
	}
}

func (c *ChatBotClient) dispatchDisconnect() {
	c.mu.RLock()
	handlers := make([]*chatDisconnectHandler, len(c.disconnectHandlers))
	copy(handlers, c.disconnectHandlers)
	c.mu.RUnlock()

	for _, h := range handlers {
		c.safeCall(func() { h.fn() })
	}
}

func (c *ChatBotClient) dispatchError(err error) {
	c.mu.RLock()
	handlers := make([]*chatErrorHandler, len(c.errorHandlers))
	copy(handlers, c.errorHandlers)
	c.mu.RUnlock()

	for _, h := range handlers {
		func() {
			defer func() { _ = recover() }() // Silently ignore panic in error handler
			h.fn(err)
		}()
	}
}

// safeCall executes a handler function with panic recovery.
func (c *ChatBotClient) safeCall(fn func()) {
	defer func() {
		if r := recover(); r != nil {
			c.dispatchError(fmt.Errorf("handler panic: %v", r))
		}
	}()
	fn()
}

// parseAuthor converts AuthorDetails to Author.
func parseAuthor(ad *AuthorDetails) *Author {
	if ad == nil {
		return nil
	}
	return &Author{
		ChannelID:       ad.ChannelID,
		DisplayName:     ad.DisplayName,
		ProfileImageURL: ad.ProfileImageURL,
		IsModerator:     ad.IsChatModerator,
		IsOwner:         ad.IsChatOwner,
		IsMember:        ad.IsChatSponsor,
		IsVerified:      ad.IsVerified,
	}
}
