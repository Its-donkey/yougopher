---
name: Bug Report
about: Report a bug to help us improve
title: '[BUG] '
labels: bug
assignees: ''
---

## Bug Description

A clear and concise description of what the bug is.

## To Reproduce

Steps to reproduce the behavior:

1. Call function '...'
2. With parameters '...'
3. See error

## Expected Behavior

A clear and concise description of what you expected to happen.

## Actual Behavior

What actually happened instead.

## Code Example

```go
// Minimal code example that reproduces the issue
client := core.NewClient()
authClient := auth.NewAuthClient(auth.Config{...})
bot := streaming.NewChatBotClient(client, authClient, liveChatID)
// Error or unexpected behavior occurs here
```

## Error Output

```
Paste any error messages or unexpected output here
```

## Environment

- **Go Version:** [e.g., 1.21.0]
- **OS:** [e.g., macOS 14.0, Ubuntu 22.04, Windows 11]
- **Package Version:** [e.g., v1.0.0 or commit hash]

## YouTube API Response (if applicable)

```json
{
  "error": {
    "code": 0,
    "message": "",
    "errors": []
  }
}
```

## Additional Context

Add any other context about the problem here, such as:
- Related issues or PRs
- Workarounds you've tried
- Screenshots if applicable
