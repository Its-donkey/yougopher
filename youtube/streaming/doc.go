// Package streaming provides live chat bot functionality for YouTube Live.
//
// # Architecture
//
// The package provides three layers for live chat:
//
//   - LiveChatPoller: Low-level HTTP polling with raw message handling
//   - LiveChatStream: Server-Sent Events (SSE) streaming for lower latency
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
// # LiveChatStream (SSE)
//
// Server-Sent Events streaming for lower latency than polling:
//
//	stream := streaming.NewLiveChatStream(client, liveChatID,
//		streaming.WithStreamAccessToken(token),
//		streaming.WithStreamMaxResults(1000),
//	)
//
//	stream.OnMessage(func(msg *streaming.LiveChatMessage) {
//		log.Printf("[%s] %s", msg.AuthorDetails.DisplayName, msg.Snippet.DisplayMessage)
//	})
//
//	stream.OnResponse(func(resp *streaming.LiveChatMessageListResponse) {
//		// Access full response metadata
//	})
//
//	stream.Start(ctx)
//	defer stream.Stop()
//
// SSE streaming provides automatic reconnection with token-based resumption.
// Use PageToken/SetPageToken for manual resumption across sessions.
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
//
// # Broadcasts
//
// Retrieve live broadcast information:
//
//	// Get a specific broadcast
//	broadcast, err := streaming.GetBroadcast(ctx, client, "broadcast-id")
//	if broadcast.IsLive() {
//		fmt.Println("Currently streaming!")
//	}
//
//	// Get authenticated user's active broadcast
//	myBroadcast, err := streaming.GetMyActiveBroadcast(ctx, client)
//
//	// Get live chat ID from broadcast
//	liveChatID, err := streaming.GetBroadcastLiveChatID(ctx, client, "broadcast-id")
//
// # Broadcast Management
//
// Create, update, and manage broadcast lifecycle:
//
//	// Create a new broadcast
//	broadcast := &streaming.LiveBroadcast{
//		Snippet: &streaming.BroadcastSnippet{
//			Title:              "My Stream",
//			ScheduledStartTime: &startTime,
//		},
//		Status: &streaming.BroadcastStatus{PrivacyStatus: "unlisted"},
//	}
//	created, err := streaming.InsertBroadcast(ctx, client, broadcast, "snippet", "status")
//
//	// Bind a stream to a broadcast
//	bound, err := streaming.BindBroadcast(ctx, client, &streaming.BindBroadcastParams{
//		BroadcastID: broadcastID,
//		StreamID:    streamID,
//	})
//
//	// Transition broadcast state
//	broadcast, err = streaming.TransitionBroadcast(ctx, client, broadcastID, streaming.TransitionLive)
//
// # Stream Management
//
// Create and manage live streams (the video feed):
//
//	// Create a new stream
//	stream := &streaming.LiveStream{
//		Snippet: &streaming.StreamSnippet{Title: "Primary Stream"},
//		CDN:     &streaming.StreamCDN{Resolution: "1080p", FrameRate: "30fps", IngestionType: "rtmp"},
//	}
//	created, err := streaming.InsertStream(ctx, client, stream, "snippet", "cdn")
//
//	// Get stream key and RTMP URL
//	stream, err := streaming.GetStream(ctx, client, streamID, "cdn")
//	streamKey := stream.StreamKey()
//	rtmpURL := stream.RTMPUrl()
//
//	// Check stream health
//	if stream.IsHealthy() {
//		fmt.Println("Stream is healthy")
//	}
//
// # StreamController (High-Level)
//
// For common streaming workflows, use StreamController:
//
//	controller, err := streaming.NewStreamController(client, authClient)
//
//	// Create broadcast with stream in one call
//	result, err := controller.CreateBroadcastWithStream(ctx, &streaming.CreateBroadcastParams{
//		Title:         "My Stream",
//		PrivacyStatus: "unlisted",
//		Resolution:    "1080p",
//	})
//
//	// Configure OBS with stream key
//	fmt.Printf("RTMP URL: %s\n", result.Stream.RTMPUrl())
//	fmt.Printf("Stream Key: %s\n", result.Stream.StreamKey())
//
//	// Start streaming in OBS, then go through states
//	controller.StartTesting(ctx, result.Broadcast.ID)
//	controller.GoLive(ctx, result.Broadcast.ID)
//	controller.EndBroadcast(ctx, result.Broadcast.ID)
package streaming
