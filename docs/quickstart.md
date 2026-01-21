---
layout: default
title: Quick Start
description: Get started with Yougopher in minutes
---

<div style="text-align: center; margin-bottom: 2rem;">
  <img src="{{ '/assets/images/logo.png' | relative_url }}" alt="Yougopher" style="width: 180px; height: 180px;">
  <h1 style="margin-top: 1rem;">Quick Start</h1>
  <p style="font-size: 1.125rem; color: #53535F;">Get up and running with Yougopher in minutes.</p>
</div>

## Prerequisites

Before you begin, you'll need:

1. **Go 1.21+** installed
2. **Google Cloud Project** with YouTube Data API v3 enabled
3. **OAuth 2.0 credentials** (client ID and secret)

## Step 1: Create Google Cloud Credentials

1. Go to the [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select an existing one
3. Enable the **YouTube Data API v3**
4. Go to **Credentials** → **Create Credentials** → **OAuth client ID**
5. Select **Desktop app** or **Web application**
6. Download the credentials JSON file

## Step 2: Install Yougopher

```bash
go get github.com/Its-donkey/yougopher
```

## Step 3: Set Up Authentication

### Option A: Device Flow (Recommended for CLI apps)

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/Its-donkey/yougopher/youtube/auth"
)

func main() {
    ctx := context.Background()

    // Create OAuth config
    config := auth.NewOAuth2Config(
        "YOUR_CLIENT_ID",
        "YOUR_CLIENT_SECRET",
        "", // No redirect URI needed for device flow
        auth.ScopeYouTube,
    )

    // Start device flow
    device, err := auth.StartDeviceFlow(ctx, config)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Go to: %s\n", device.VerificationURL)
    fmt.Printf("Enter code: %s\n", device.UserCode)

    // Poll for token
    token, err := auth.PollDeviceToken(ctx, config, device)
    if err != nil {
        log.Fatal(err)
    }

    // Save token for future use
    if err := auth.SaveToken("token.json", token); err != nil {
        log.Fatal(err)
    }

    fmt.Println("Authentication successful!")
}
```

### Option B: Web Flow

```go
config := auth.NewOAuth2Config(
    "YOUR_CLIENT_ID",
    "YOUR_CLIENT_SECRET",
    "http://localhost:8080/callback",
    auth.ScopeYouTube,
)

// Generate auth URL
url := config.AuthCodeURL("state")
fmt.Printf("Visit: %s\n", url)

// Exchange code for token (in your callback handler)
token, err := config.Exchange(ctx, authCode)
```

## Step 4: Make API Calls

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/Its-donkey/yougopher/youtube/auth"
    "github.com/Its-donkey/yougopher/youtube/core"
    "github.com/Its-donkey/yougopher/youtube/data"
)

func main() {
    ctx := context.Background()

    // Load saved token
    token, err := auth.LoadToken("token.json")
    if err != nil {
        log.Fatal(err)
    }

    // Create config and client
    config := auth.NewOAuth2Config(
        "YOUR_CLIENT_ID",
        "YOUR_CLIENT_SECRET",
        "",
        auth.ScopeYouTube,
    )

    client := core.NewClient(ctx, config.Client(ctx, token))

    // Get your channel
    resp, err := data.GetChannels(ctx, client, &data.GetChannelsParams{
        Mine: true,
        Part: []string{"snippet", "statistics"},
    })
    if err != nil {
        log.Fatal(err)
    }

    if len(resp.Items) > 0 {
        ch := resp.Items[0]
        fmt.Printf("Channel: %s\n", ch.Snippet.Title)
        fmt.Printf("Subscribers: %d\n", ch.Statistics.SubscriberCount)
        fmt.Printf("Videos: %d\n", ch.Statistics.VideoCount)
    }
}
```

## Step 5: Work with Live Streams

```go
import "github.com/Its-donkey/yougopher/youtube/streaming"

// Get active broadcasts
broadcasts, err := streaming.GetBroadcasts(ctx, client, &streaming.GetBroadcastsParams{
    BroadcastStatus: "active",
    Part:            []string{"snippet", "status"},
})

for _, b := range broadcasts.Items {
    fmt.Printf("Live: %s\n", b.Snippet.Title)

    // Get live chat ID
    chatID := b.Snippet.LiveChatID

    // Poll chat messages
    poller := streaming.NewLiveChatPoller(client, chatID)
    poller.OnMessage(func(msg *streaming.LiveChatMessage) {
        fmt.Printf("%s: %s\n",
            msg.AuthorDetails.DisplayName,
            msg.Snippet.DisplayMessage)
    })
    poller.Start(ctx)
}
```

## Next Steps

- [API Reference](api-reference) - Complete documentation for all packages
- [Authentication Guide](auth) - Detailed auth options
- [Cookbook](cookbook) - Common patterns and recipes
- [Live Streaming API](live-streaming/) - Real-time chat and streaming

## Common Issues

### "Access Not Configured" Error

Enable the YouTube Data API v3 in your Google Cloud Console.

### "Invalid Credentials" Error

Check that your client ID and secret are correct, and that your OAuth consent screen is configured.

### Quota Exceeded

The YouTube API has daily quota limits. See [Quota Management](troubleshooting#quota-management) for tips.
