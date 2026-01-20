---
layout: default
title: Auth API
description: OAuth 2.0 authentication for YouTube APIs with automatic token refresh.
---

## Overview

Handle OAuth 2.0 authentication for YouTube APIs:

**OAuth Flow:** Complete Google OAuth 2.0 implementation
- Generate authorization URLs
- Exchange codes for tokens
- Automatic token refresh

**Token Management:** Secure token handling
- Thread-safe token storage
- Automatic refresh before expiry
- Callbacks for token events

## Scopes

```go
const (
    // ScopeLiveChat grants read access to live chat messages.
    ScopeLiveChat = "https://www.googleapis.com/auth/youtube"

    // ScopeLiveChatModerate grants moderator access to live chat.
    ScopeLiveChatModerate = "https://www.googleapis.com/auth/youtube.force-ssl"

    // ScopeReadOnly grants read-only access to YouTube account.
    ScopeReadOnly = "https://www.googleapis.com/auth/youtube.readonly"

    // ScopeUpload grants access to upload videos and manage playlists.
    ScopeUpload = "https://www.googleapis.com/auth/youtube.upload"

    // ScopePartner grants access to YouTube Analytics.
    ScopePartner = "https://www.googleapis.com/auth/youtubepartner"
)
```

## NewAuthClient

Create a new OAuth client with configuration.

```go
authClient := auth.NewAuthClient(auth.Config{
    ClientID:     "your-client-id",
    ClientSecret: "your-client-secret",
    RedirectURL:  "http://localhost:8080/callback",
    Scopes: []string{
        auth.ScopeLiveChat,
        auth.ScopeLiveChatModerate,
    },
})
```

### Options

```go
authClient := auth.NewAuthClient(config,
    auth.WithHTTPClient(customHTTPClient),
    auth.WithToken(existingToken),
    auth.WithRefreshEarly(5*time.Minute),
    auth.WithOnTokenRefresh(func(token *auth.Token) {
        // Save token to storage
    }),
    auth.WithOnRefreshError(func(err error) {
        log.Printf("Token refresh failed: %v", err)
    }),
)
```

## OAuth Flow

### AuthorizationURL

Generate the URL to redirect users for authorization.

```go
state := generateRandomState() // Use a cryptographically secure random string
url := authClient.AuthorizationURL(state)

// Redirect user to url
http.Redirect(w, r, url, http.StatusFound)
```

### Options

```go
url := authClient.AuthorizationURL(state,
    auth.WithPrompt("consent"),      // Force consent screen
    auth.WithLoginHint("user@example.com"), // Hint which account
)
```

### Exchange

Exchange an authorization code for a token.

```go
// In your callback handler:
code := r.URL.Query().Get("code")

token, err := authClient.Exchange(ctx, code)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Access token: %s\n", token.AccessToken)
fmt.Printf("Expires: %s\n", token.Expiry)
```

## Token Management

### Token

Get the current token (thread-safe).

```go
token := authClient.Token()
if token != nil {
    fmt.Printf("Access token: %s\n", token.AccessToken)
    fmt.Printf("Valid: %v\n", token.Valid())
}
```

### SetToken

Set a token (e.g., loaded from storage).

```go
authClient.SetToken(&auth.Token{
    AccessToken:  savedAccessToken,
    RefreshToken: savedRefreshToken,
    Expiry:       savedExpiry,
    Scopes:       savedScopes,
})
```

### AccessToken

Get a valid access token, automatically refreshing if expired.

```go
accessToken, err := authClient.AccessToken(ctx)
if err != nil {
    log.Fatal(err)
}
// Use accessToken for API calls
```

### Refresh

Manually refresh the access token.

```go
newToken, err := authClient.Refresh(ctx)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("New access token: %s\n", newToken.AccessToken)
```

## Auto-Refresh

### StartAutoRefresh

Start automatic token refresh in the background.

```go
err := authClient.StartAutoRefresh(ctx)
if err != nil {
    log.Fatal(err)
}

// Token will be refreshed automatically before expiry
```

### StopAutoRefresh

Stop the auto-refresh goroutine.

```go
authClient.StopAutoRefresh()
```

## Token Type

```go
type Token struct {
    AccessToken  string
    TokenType    string
    RefreshToken string
    Expiry       time.Time
    Scopes       []string
}
```

### Methods

```go
// Check if token is valid (not expired)
valid := token.Valid()

// Get access token (thread-safe)
accessToken := token.GetAccessToken()

// Clone token (for safe passing between goroutines)
clone := token.Clone()

// Serialize to JSON
data, err := token.MarshalJSON()

// Deserialize from JSON
err := token.UnmarshalJSON(data)
```

## Example: Complete OAuth Flow

```go
package main

import (
    "context"
    "fmt"
    "log"
    "net/http"

    "github.com/Its-donkey/yougopher/youtube/auth"
)

func main() {
    authClient := auth.NewAuthClient(auth.Config{
        ClientID:     "your-client-id",
        ClientSecret: "your-client-secret",
        RedirectURL:  "http://localhost:8080/callback",
        Scopes:       []string{auth.ScopeLiveChat, auth.ScopeLiveChatModerate},
    },
        auth.WithOnTokenRefresh(func(token *auth.Token) {
            log.Println("Token refreshed, save to storage...")
        }),
    )

    // Step 1: Redirect to authorization URL
    http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
        state := "random-state-string" // Use crypto/rand in production
        url := authClient.AuthorizationURL(state)
        http.Redirect(w, r, url, http.StatusFound)
    })

    // Step 2: Handle callback
    http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
        code := r.URL.Query().Get("code")

        token, err := authClient.Exchange(r.Context(), code)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }

        // Start auto-refresh
        authClient.StartAutoRefresh(context.Background())

        fmt.Fprintf(w, "Authenticated! Token expires: %s", token.Expiry)
    })

    log.Println("Visit http://localhost:8080/login to authenticate")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

## Thread Safety

`AuthClient` is safe for concurrent use. All token operations are protected by mutex locks.
