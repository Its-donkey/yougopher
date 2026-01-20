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
//	bot := streaming.NewChatBotClient(client, authClient, liveChatID)
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
package streaming
