---
layout: default
title: Example Chat Bot
description: Basic live chat bot example
---

A basic live chat bot that responds to commands.

**Directory:** `examples/chatbot/`

## Run

**Requirements:** Active YouTube live stream

1. Complete the [setup steps](./index.md#setup) (clone, credentials, env vars)
2. Navigate to the example:
   ```bash
   cd examples/chatbot
   ```
3. Run the bot:
   ```bash
   go run main.go
   ```
4. Open http://localhost:8080/login and complete OAuth
5. The bot will connect to your active broadcast and start responding to commands

## Features

- OAuth authentication with local callback server
- Automatic broadcast detection
- Message event handling
- Super Chat and membership event logging

## Commands

| Command | Description |
|---------|-------------|
| `!hello` | Greet the user |
| `!time` | Show current time |
| `!help` | List available commands |

## Customization

Add new commands in `handleCommand()`:

```go
func handleCommand(ctx context.Context, poller *streaming.LiveChatPoller, msg *streaming.LiveChatMessage) {
    text := msg.Message()
    author := msg.AuthorDetails.DisplayName

    switch {
    case strings.HasPrefix(text, "!hello"):
        poller.SendMessage(ctx, fmt.Sprintf("Hello, %s!", author))

    case strings.HasPrefix(text, "!time"):
        poller.SendMessage(ctx, fmt.Sprintf("Current time: %s", time.Now().Format(time.RFC1123)))

    case strings.HasPrefix(text, "!help"):
        poller.SendMessage(ctx, "Commands: !hello, !time, !help")

    // Add your custom commands here
    case strings.HasPrefix(text, "!ping"):
        poller.SendMessage(ctx, "Pong!")
    }
}
```

## Event Handling

```go
poller := streaming.NewLiveChatPoller(client, liveChatID)

// Handle regular messages
poller.OnMessage(func(msg *streaming.LiveChatMessage) {
    fmt.Printf("[%s] %s\n", msg.AuthorDetails.DisplayName, msg.Message())

    if strings.HasPrefix(msg.Message(), "!") {
        handleCommand(ctx, poller, msg)
    }
})

// Handle Super Chats
poller.OnSuperChat(func(msg *streaming.LiveChatMessage) {
    details := msg.Snippet.SuperChatDetails
    fmt.Printf("SUPER CHAT from %s: %s - %s\n",
        msg.AuthorDetails.DisplayName,
        details.AmountDisplayString,
        details.UserComment)
})

// Handle new members
poller.OnMembership(func(msg *streaming.LiveChatMessage) {
    fmt.Printf("NEW MEMBER: %s\n", msg.AuthorDetails.DisplayName)
})

poller.Start(ctx)
```
