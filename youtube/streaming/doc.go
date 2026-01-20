// Package streaming provides live chat bot functionality for YouTube Live.
//
// # Architecture
//
// The package provides two layers:
//
//   - LiveChatPoller: Low-level HTTP polling with raw message handling
//   - ChatBotClient: High-level bot with semantic event handlers
//
// # ChatBotClient (Recommended)
//
// The high-level client for building chat bots:
//
//	bot, err := streaming.NewChatBotClient(client, authClient, liveChatID)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Register handlers
//	bot.OnMessage(func(msg *streaming.ChatMessage) {
//		log.Printf("[%s] %s", msg.Author.DisplayName, msg.Message)
//	})
//
//	bot.OnSuperChat(func(event *streaming.SuperChatEvent) {
//		log.Printf("SuperChat: %s - %s", event.Amount, event.Message)
//	})
//
//	// Connect and start listening
//	if err := bot.Connect(ctx); err != nil {
//		log.Fatal(err)
//	}
//	defer bot.Close()
//
// # LiveChatPoller (Advanced)
//
// The low-level poller for custom implementations:
//
//	poller := streaming.NewLiveChatPoller(client, liveChatID)
//
//	poller.OnMessage(func(event *streaming.LiveChatMessageEvent) {
//		// Handle raw message event
//	})
//
//	poller.Start(ctx)
//	defer poller.Stop()
//
// # Moderation
//
// Both clients support moderation actions:
//
//	bot.Ban(ctx, channelID)
//	bot.Timeout(ctx, channelID, 300) // 5 minute timeout
//	bot.Delete(ctx, messageID)
//
// # Handler Pattern
//
// Handlers return an unsubscribe function for cleanup:
//
//	unsub := bot.OnMessage(handler)
//	// Later...
//	unsub() // Safe to call multiple times (idempotent)
//
// # Handler Safety
//
// All handlers are called with panic recovery. If a handler panics,
// the error is dispatched to error handlers and polling continues.
// Error handlers themselves are protected against panics to prevent
// infinite recursion.
//
// # Thread Safety
//
// Both LiveChatPoller and ChatBotClient are safe for concurrent use.
// Handler registration and unsubscription can be done from any goroutine.
//
// # Raw Message Access
//
// All event types include a Raw field containing the underlying
// LiveChatMessage for access to additional fields not exposed in the
// semantic types. Note that the Raw pointer is shared across handlers
// for efficiency; if you need to modify it, use Clone() first.
package streaming
