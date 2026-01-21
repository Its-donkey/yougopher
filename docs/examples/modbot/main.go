// Example: Moderation Bot
//
// This example demonstrates how to create a YouTube live chat moderation bot
// with auto-moderation features like spam detection, bad word filtering,
// and user management.
//
// Usage:
//
//	export YOUTUBE_CLIENT_ID=your-client-id
//	export YOUTUBE_CLIENT_SECRET=your-client-secret
//	go run main.go
//
// The bot will:
//  1. Authenticate via OAuth
//  2. Connect to your active broadcast's live chat
//  3. Monitor messages for spam and bad words
//  4. Allow moderators to use !ban, !timeout, and !unban commands
//  5. Track message frequency for spam detection
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/Its-donkey/yougopher/youtube/auth"
	"github.com/Its-donkey/yougopher/youtube/core"
	"github.com/Its-donkey/yougopher/youtube/streaming"
)

// ModerationConfig holds configuration for auto-moderation.
type ModerationConfig struct {
	// BadWords to filter (case-insensitive)
	BadWords []string

	// MaxMessagesPerMinute before warning
	MaxMessagesPerMinute int

	// TimeoutDuration for auto-moderation actions
	TimeoutDuration int // seconds

	// EnableBadWordFilter toggles bad word filtering
	EnableBadWordFilter bool

	// EnableSpamDetection toggles spam detection
	EnableSpamDetection bool
}

// MessageTracker tracks user message frequency.
type MessageTracker struct {
	mu       sync.Mutex
	messages map[string][]time.Time // channelID -> timestamps
}

func NewMessageTracker() *MessageTracker {
	return &MessageTracker{
		messages: make(map[string][]time.Time),
	}
}

func (t *MessageTracker) AddMessage(channelID string) int {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-1 * time.Minute)

	// Filter old messages and add new one
	var recent []time.Time
	for _, ts := range t.messages[channelID] {
		if ts.After(cutoff) {
			recent = append(recent, ts)
		}
	}
	recent = append(recent, now)
	t.messages[channelID] = recent

	return len(recent)
}

func main() {
	// Load credentials
	clientID := os.Getenv("YOUTUBE_CLIENT_ID")
	clientSecret := os.Getenv("YOUTUBE_CLIENT_SECRET")

	if clientID == "" || clientSecret == "" {
		log.Fatal("Set YOUTUBE_CLIENT_ID and YOUTUBE_CLIENT_SECRET environment variables")
	}

	// Moderation configuration
	config := ModerationConfig{
		BadWords: []string{
			"spam", "scam", "free money",
			// Add your bad words here
		},
		MaxMessagesPerMinute: 10,
		TimeoutDuration:      300, // 5 minutes
		EnableBadWordFilter:  true,
		EnableSpamDetection:  true,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create clients
	authClient := auth.NewAuthClient(auth.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  "http://localhost:8080/callback",
		Scopes: []string{
			auth.ScopeLiveChat,
			auth.ScopeLiveChatModerate,
		},
	})

	client := core.NewClient()
	tracker := NewMessageTracker()

	// Auth flow (simplified for example)
	authDone := make(chan struct{})

	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		state := fmt.Sprintf("%d", time.Now().UnixNano())
		url := authClient.AuthorizationURL(state, auth.WithPrompt("consent"))
		http.Redirect(w, r, url, http.StatusFound)
	})

	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		token, err := authClient.Exchange(r.Context(), code)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		client.SetAccessToken(token.AccessToken)
		_ = authClient.StartAutoRefresh(ctx)

		fmt.Fprintf(w, "Authentication successful!")
		close(authDone)
	})

	server := &http.Server{Addr: ":8080"}
	go func() {
		log.Println("Visit http://localhost:8080/login to authenticate")
		_ = server.ListenAndServe()
	}()

	<-authDone
	time.Sleep(500 * time.Millisecond)
	_ = server.Shutdown(ctx)

	// Find broadcast
	log.Println("Looking for active broadcast...")
	broadcast, err := streaming.GetMyActiveBroadcast(ctx, client)
	if err != nil {
		log.Fatalf("Failed to get broadcast: %v", err)
	}
	if broadcast == nil {
		log.Fatal("No active broadcast. Start a live stream first!")
	}

	liveChatID := broadcast.Snippet.LiveChatID
	log.Printf("Connected to: %s", broadcast.Snippet.Title)

	// Create bot
	bot, err := streaming.NewChatBotClient(client, authClient, liveChatID)
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}

	// Message handler with moderation
	bot.OnMessage(func(msg *streaming.ChatMessage) {
		// Log message
		logMessage(msg)

		// Skip moderation for moderators and owner
		if msg.Author.IsModerator || msg.Author.IsOwner {
			// Handle mod commands
			handleModCommand(ctx, bot, msg)
			return
		}

		// Auto-moderation checks
		if config.EnableBadWordFilter && containsBadWord(msg.Message, config.BadWords) {
			log.Printf("BAD WORD DETECTED from %s: %s", msg.Author.DisplayName, msg.Message)

			// Delete message and timeout user
			if err := bot.Delete(ctx, msg.ID); err != nil {
				log.Printf("Failed to delete message: %v", err)
			}
			if err := bot.Timeout(ctx, msg.Author.ChannelID, config.TimeoutDuration); err != nil {
				log.Printf("Failed to timeout user: %v", err)
			}
			return
		}

		if config.EnableSpamDetection {
			count := tracker.AddMessage(msg.Author.ChannelID)
			if count > config.MaxMessagesPerMinute {
				log.Printf("SPAM DETECTED from %s: %d messages/min",
					msg.Author.DisplayName, count)

				if err := bot.Timeout(ctx, msg.Author.ChannelID, config.TimeoutDuration); err != nil {
					log.Printf("Failed to timeout spammer: %v", err)
				}
				return
			}
		}
	})

	bot.OnUserBanned(func(event *streaming.BanEvent) {
		banType := "BANNED"
		if event.BanType == streaming.BanTypeTemporary {
			banType = fmt.Sprintf("TIMED OUT (%v)", event.Duration)
		}
		userName := "unknown"
		if event.BannedUser != nil {
			userName = event.BannedUser.DisplayName
		}
		log.Printf("USER %s: %s", banType, userName)
	})

	bot.OnConnect(func() {
		log.Println("Moderation bot connected!")
	})

	bot.OnError(func(err error) {
		log.Printf("Error: %v", err)
	})

	// Connect
	if err := bot.Connect(ctx); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer bot.Close()

	log.Println("Moderation bot running! Commands: !ban <user>, !timeout <user> [seconds], !unban <banID>")

	// Wait for shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	log.Println("Shutting down...")
}

func logMessage(msg *streaming.ChatMessage) {
	role := ""
	if msg.Author.IsOwner {
		role = "[OWNER]"
	} else if msg.Author.IsModerator {
		role = "[MOD]"
	} else if msg.Author.IsMember {
		role = "[MEMBER]"
	}

	log.Printf("%s %s%s: %s",
		msg.PublishedAt.Format("15:04:05"),
		role,
		msg.Author.DisplayName,
		msg.Message,
	)
}

func containsBadWord(text string, badWords []string) bool {
	lower := strings.ToLower(text)
	for _, word := range badWords {
		if strings.Contains(lower, strings.ToLower(word)) {
			return true
		}
	}
	return false
}

var cmdRegex = regexp.MustCompile(`^!(\w+)\s*(.*)$`)

func handleModCommand(ctx context.Context, bot *streaming.ChatBotClient, msg *streaming.ChatMessage) {
	matches := cmdRegex.FindStringSubmatch(msg.Message)
	if matches == nil {
		return
	}

	cmd := strings.ToLower(matches[1])
	args := strings.TrimSpace(matches[2])

	switch cmd {
	case "ban":
		if args == "" {
			return
		}
		// Note: In real use, you'd need to resolve the username to channel ID
		log.Printf("MOD ACTION: %s requested ban for %s", msg.Author.DisplayName, args)
		// bot.Ban(ctx, channelID) would be called with resolved channel ID

	case "timeout":
		if args == "" {
			return
		}
		parts := strings.Fields(args)
		duration := 300 // default 5 minutes
		if len(parts) > 1 {
			fmt.Sscanf(parts[1], "%d", &duration)
		}
		log.Printf("MOD ACTION: %s requested %ds timeout for %s",
			msg.Author.DisplayName, duration, parts[0])
		// bot.Timeout(ctx, channelID, duration)

	case "unban":
		if args == "" {
			return
		}
		// args should be the ban ID
		if err := bot.Unban(ctx, args); err != nil {
			log.Printf("Failed to unban: %v", err)
		} else {
			log.Printf("MOD ACTION: %s unbanned %s", msg.Author.DisplayName, args)
		}

	case "stats":
		if err := bot.Say(ctx, "Moderation bot is active and monitoring chat."); err != nil {
			log.Printf("Failed to send stats: %v", err)
		}
	}
}
