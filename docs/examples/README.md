# Yougopher Examples

Example applications demonstrating Yougopher's capabilities.

## Prerequisites

1. Create OAuth credentials in the [Google Cloud Console](https://console.cloud.google.com/apis/credentials)
2. Enable the YouTube Data API v3 and YouTube Analytics API
3. Set environment variables:
   ```bash
   export YOUTUBE_CLIENT_ID=your-client-id
   export YOUTUBE_CLIENT_SECRET=your-client-secret
   ```

## Running Examples

1. Navigate to the example directory
2. Run `go run main.go`
3. Open `http://localhost:8080/login` in your browser
4. Complete the OAuth flow
5. The example will connect to your active broadcast (or display analytics)

---

## Chat Bot

**Directory:** `chatbot/`

**Description:** A basic live chat bot that responds to commands.

**Run:**
```bash
cd chatbot
go run main.go
```

**Features:**
- OAuth authentication with local callback server
- Automatic broadcast detection
- Message event handling
- SuperChat and membership event logging

**Commands:**
| Command | Description |
|---------|-------------|
| `!hello` | Greet the user |
| `!time` | Show current time |
| `!help` | List available commands |

**Customize:** Add new commands in `handleCommand()`

---

## Moderation Bot

**Directory:** `modbot/`

**Description:** An auto-moderation bot with spam detection and bad word filtering.

**Run:**
```bash
cd modbot
go run main.go
```

**Features:**
- Bad word filtering with auto-delete and timeout
- Spam detection (messages per minute tracking)
- Role-based permissions (mods and owner bypass filters)
- Ban event logging

**Commands:**
| Command | Description |
|---------|-------------|
| `!ban @user` | Permanently ban a user |
| `!timeout @user` | Timeout a user (5 minutes) |
| `!unban @user` | Remove a user's ban |
| `!stats` | Show moderation statistics |

**Customize:** Edit `ModerationConfig` for bad words and thresholds

---

## Analytics Dashboard

**Directory:** `analytics/`

**Description:** A CLI dashboard displaying channel statistics.

**Run:**
```bash
cd analytics
go run main.go
```

**Features:**
- Channel overview (views, watch time, subscribers)
- Top 10 videos by views
- Daily view trend with ASCII bar chart
- Geographic breakdown by country
- Device type breakdown

**Scopes Required:**
- `youtube.readonly`
- `youtubepartner`

**Customize:** Change date ranges or add more report types in `main.go`

---

## Notes

| Example | Requirements |
|---------|--------------|
| chatbot | Active YouTube live stream |
| modbot | Active live stream + moderator permissions |
| analytics | No live stream required |

**Security:** Examples do not persist tokens. Store tokens securely in production.
