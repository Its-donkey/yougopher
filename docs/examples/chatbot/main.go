// Example: Basic Chat Bot
//
// This example demonstrates how to create a simple YouTube live chat bot
// that responds to messages and handles various chat events.
//
// Usage:
//
//	export YOUTUBE_CLIENT_ID=your-client-id
//	export YOUTUBE_CLIENT_SECRET=your-client-secret
//	go run main.go
//
// The bot will:
//  1. Start a local server for OAuth callback
//  2. Open a browser for authentication
//  3. Connect to the live chat of your active broadcast
//  4. Respond to !hello and !time commands
//  5. Log all chat events
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/Its-donkey/yougopher/youtube/auth"
	"github.com/Its-donkey/yougopher/youtube/core"
	"github.com/Its-donkey/yougopher/youtube/streaming"
)

func main() {
	// Load credentials from environment
	clientID := os.Getenv("YOUTUBE_CLIENT_ID")
	clientSecret := os.Getenv("YOUTUBE_CLIENT_SECRET")

	if clientID == "" || clientSecret == "" {
		log.Fatal("Set YOUTUBE_CLIENT_ID and YOUTUBE_CLIENT_SECRET environment variables")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create auth client
	authClient := auth.NewAuthClient(auth.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  "http://localhost:8080/callback",
		Scopes: []string{
			auth.ScopeLiveChat,
			auth.ScopeLiveChatModerate,
		},
	},
		auth.WithOnTokenRefresh(func(token *auth.Token) {
			log.Println("Token refreshed automatically")
		}),
	)

	// Create core client
	client := core.NewClient()

	// Channel to signal when auth is complete
	authDone := make(chan struct{})

	// Start OAuth server
	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		state := fmt.Sprintf("%d", time.Now().UnixNano())
		url := authClient.AuthorizationURL(state, auth.WithPrompt("consent"))
		http.Redirect(w, r, url, http.StatusFound)
	})

	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code == "" {
			http.Error(w, "No code provided", http.StatusBadRequest)
			return
		}

		token, err := authClient.Exchange(r.Context(), code)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Update core client with access token
		client.SetAccessToken(token.AccessToken)

		// Start auto-refresh
		if err := authClient.StartAutoRefresh(ctx); err != nil {
			log.Printf("Warning: Could not start auto-refresh: %v", err)
		}

		_, _ = fmt.Fprintf(w, "Authentication successful! You can close this window.")
		close(authDone)
	})

	// Start HTTP server
	server := &http.Server{Addr: ":8080"}
	go func() {
		log.Println("Starting OAuth server on http://localhost:8080")
		log.Println("Visit http://localhost:8080/login to authenticate")
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	// Wait for authentication
	log.Println("Waiting for authentication...")
	<-authDone

	// Give a moment for the response to be sent
	time.Sleep(500 * time.Millisecond)

	// Shutdown OAuth server
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Error shutting down server: %v", err)
	}

	// Find active broadcast
	log.Println("Looking for active broadcast...")
	broadcast, err := streaming.GetMyActiveBroadcast(ctx, client)
	if err != nil {
		log.Fatalf("Failed to get active broadcast: %v", err)
	}
	if broadcast == nil {
		log.Fatal("No active broadcast found. Start a live stream first!")
	}

	liveChatID := broadcast.Snippet.LiveChatID
	if liveChatID == "" {
		log.Fatal("Broadcast has no live chat ID")
	}

	log.Printf("Found broadcast: %s", broadcast.Snippet.Title)
	log.Printf("Live Chat ID: %s", liveChatID)

	// Create chat bot
	bot, err := streaming.NewChatBotClient(client, authClient, liveChatID)
	if err != nil {
		log.Fatalf("Failed to create chat bot: %v", err)
	}

	// Register event handlers
	bot.OnMessage(func(msg *streaming.ChatMessage) {
		log.Printf("[%s] %s: %s",
			msg.PublishedAt.Format("15:04:05"),
			msg.Author.DisplayName,
			msg.Message,
		)

		// Handle commands
		handleCommand(ctx, bot, msg)
	})

	bot.OnSuperChat(func(event *streaming.SuperChatEvent) {
		log.Printf("SUPER CHAT from %s: $%.2f %s - %s",
			event.Author.DisplayName,
			float64(event.AmountMicros)/1000000,
			event.Currency,
			event.Message,
		)
	})

	bot.OnMembership(func(event *streaming.MembershipEvent) {
		log.Printf("NEW MEMBER: %s joined at %s level!",
			event.Author.DisplayName,
			event.LevelName,
		)
	})

	bot.OnConnect(func() {
		log.Println("Connected to live chat!")
	})

	bot.OnDisconnect(func() {
		log.Println("Disconnected from live chat")
	})

	bot.OnError(func(err error) {
		log.Printf("Error: %v", err)
	})

	// Connect to live chat
	log.Println("Connecting to live chat...")
	if err := bot.Connect(ctx); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer func() { _ = bot.Close() }()

	log.Println("Bot is running! Press Ctrl+C to stop.")

	// Wait for shutdown signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	log.Println("Shutting down...")
	cancel()
}

func handleCommand(ctx context.Context, bot *streaming.ChatBotClient, msg *streaming.ChatMessage) {
	text := strings.TrimSpace(msg.Message)

	switch text {
	case "!hello":
		if err := bot.Say(ctx, fmt.Sprintf("Hello, %s!", msg.Author.DisplayName)); err != nil {
			log.Printf("Failed to send message: %v", err)
		}

	case "!time":
		now := time.Now().Format("3:04 PM MST")
		if err := bot.Say(ctx, fmt.Sprintf("The current time is %s", now)); err != nil {
			log.Printf("Failed to send message: %v", err)
		}

	case "!help":
		if err := bot.Say(ctx, "Available commands: !hello, !time, !help"); err != nil {
			log.Printf("Failed to send message: %v", err)
		}
	}
}
