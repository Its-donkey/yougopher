---
layout: default
title: Troubleshooting
description: Common issues and solutions for Yougopher
---

## Authentication Errors

### "Access Not Configured"

**Error:** `Access Not Configured. YouTube Data API has not been used in project X before or it is disabled.`

**Solution:**
1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Select your project
3. Navigate to APIs & Services → Library
4. Search for "YouTube Data API v3"
5. Click Enable

### "Invalid Client"

**Error:** `invalid_client: The OAuth client was not found.`

**Solution:**
- Verify your client ID is correct
- Check that you're using the right credentials for your environment (web vs desktop)
- Ensure the OAuth consent screen is configured

### "Invalid Grant"

**Error:** `invalid_grant: Token has been expired or revoked.`

**Solution:**
- Delete your saved token file and re-authenticate
- Check if the user revoked access in their Google account settings
- For refresh tokens, ensure offline access was requested

### "Redirect URI Mismatch"

**Error:** `redirect_uri_mismatch`

**Solution:**
- The redirect URI in your code must exactly match one configured in Google Cloud Console
- Check for trailing slashes, http vs https, and port numbers
- For device flow, leave redirect URI empty

---

## API Errors

### "Quota Exceeded"

**Error:** `quotaExceeded: The request cannot be completed because you have exceeded your quota.`

**Solution:**
- Wait until quota resets (midnight Pacific Time)
- Request a quota increase in Google Cloud Console
- Optimize your code to reduce API calls:
  - Use caching
  - Batch requests where possible
  - Only request needed `part` values
  - Reduce polling frequency

### "Rate Limit Exceeded"

**Error:** `rateLimitExceeded`

**Solution:**
Implement exponential backoff:

```go
func withBackoff(fn func() error) error {
    backoff := time.Second
    for i := 0; i < 5; i++ {
        err := fn()
        if err == nil {
            return nil
        }
        var apiErr *core.APIError
        if errors.As(err, &apiErr) {
            if apiErr.Code == 429 || apiErr.Reason == "rateLimitExceeded" {
                time.Sleep(backoff)
                backoff *= 2
                continue
            }
        }
        return err
    }
    return errors.New("max retries exceeded")
}
```

### "Forbidden"

**Error:** `forbidden: The request is not properly authorized.`

**Solution:**
- Check that your OAuth scopes include the required permissions
- For write operations, ensure you have `ScopeYouTube` not just `ScopeYouTubeReadOnly`
- Verify the authenticated user has permission for the requested resource

### "Not Found"

**Error:** `notFound: The resource could not be found.`

**Solution:**
- Verify the resource ID is correct
- Check if the resource was deleted
- Ensure the resource is accessible (not private)

---

## Live Streaming Errors

### "Live Chat Not Found"

**Error:** `liveChatNotFound`

**Solution:**
- Verify the broadcast is currently live
- Get a fresh chat ID from the broadcast object
- Check that chat is enabled for the broadcast

### "Live Chat Ended"

**Error:** `liveChatEnded: The live chat has ended.`

**Solution:**
- The broadcast has ended; stop polling/streaming
- Handle this gracefully in your application

### "Message Text Required"

**Error:** `messageTextRequired`

**Solution:**
- Ensure the message content is not empty
- Check for whitespace-only messages

### "Message Too Long"

**Error:** `messageTooLong`

**Solution:**
- YouTube limits chat messages to 200 characters
- Truncate or split long messages

### "Banned from Chat"

**Error:** `forbidden: The user is banned from this chat.`

**Solution:**
- The authenticated user has been banned from the chat
- Contact the broadcast owner to be unbanned

---

## Quota Management

### Understanding Quota Costs

| Operation | Cost |
|-----------|------|
| `list` operations | 1-5 units |
| `insert` operations | 50 units |
| `update` operations | 50 units |
| `delete` operations | 50 units |
| `liveChatMessages.list` | 5 units |
| Video upload | 1,600 units |

### Reducing Quota Usage

1. **Only request needed parts:**
   ```go
   // Bad - requests everything
   Part: []string{"snippet", "status", "contentDetails", "statistics"}

   // Good - only what you need
   Part: []string{"snippet"}
   ```

2. **Use caching:**
   ```go
   client := core.NewClient(ctx, httpClient,
       core.WithCache(cache, 5*time.Minute),
   )
   ```

3. **Paginate efficiently:**
   ```go
   // Request maximum items per page
   MaxResults: 50
   ```

4. **Use SSE instead of polling for live chat:**
   SSE uses a persistent connection, reducing the number of API calls.

---

## Network Issues

### "Context Deadline Exceeded"

**Error:** `context deadline exceeded`

**Solution:**
- Increase timeout duration
- Check network connectivity
- The API may be experiencing issues

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
```

### "Connection Refused"

**Error:** `connection refused`

**Solution:**
- Check internet connectivity
- Verify firewall settings
- YouTube API endpoints should be accessible

### "TLS Handshake Timeout"

**Error:** `net/http: TLS handshake timeout`

**Solution:**
- Network latency issues
- Try again later
- Check proxy/firewall settings

---

## Common Mistakes

### Not Handling Pagination

**Problem:** Only getting first page of results.

**Solution:**
```go
pageToken := ""
for {
    resp, err := data.Search(ctx, client, &data.SearchParams{
        Query:      "query",
        MaxResults: 50,
        PageToken:  pageToken,
    })
    if err != nil {
        break
    }

    // Process items...

    if resp.NextPageToken == "" {
        break
    }
    pageToken = resp.NextPageToken
}
```

### Not Refreshing Tokens

**Problem:** Tokens expire after 1 hour.

**Solution:** Use the OAuth2 client which auto-refreshes:
```go
httpClient := config.Client(ctx, token)  // Auto-refreshes
client := core.NewClient(ctx, httpClient)
```

### Ignoring Error Types

**Problem:** Generic error handling misses actionable errors.

**Solution:**
```go
if err != nil {
    var apiErr *core.APIError
    if errors.As(err, &apiErr) {
        switch apiErr.Reason {
        case "quotaExceeded":
            // Handle quota
        case "rateLimitExceeded":
            // Handle rate limit
        default:
            // Handle other API errors
        }
    }
}
```

---

## Getting Help

If you're still having issues:

1. Check the [FAQ](faq) for common questions
2. Search [existing issues](https://github.com/Its-donkey/yougopher/issues)
3. Open a new issue with:
   - Error message
   - Code snippet
   - Go version
   - Yougopher version
