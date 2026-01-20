// Package auth provides OAuth 2.0 authentication for YouTube APIs.
//
// # Authorization Code Flow
//
// The standard OAuth flow for server-side applications:
//
//	authClient := auth.NewAuthClient(auth.Config{
//		ClientID:     "your-client-id",
//		ClientSecret: "your-client-secret",
//		RedirectURL:  "http://localhost:8080/callback",
//		Scopes:       []string{auth.ScopeLiveChat, auth.ScopeLiveChatModerate},
//	})
//
//	// Generate authorization URL
//	url := authClient.AuthorizationURL("state-token")
//
//	// Exchange code for token
//	token, err := authClient.Exchange(ctx, code)
//
// # Token Management
//
// Tokens are automatically refreshed when expired:
//
//	// Token includes access token, refresh token, and expiry
//	token := authClient.Token()
//
//	// Check if valid
//	if token.Valid() {
//		// Use token
//	}
//
// # Scopes
//
// Common YouTube API scopes:
//
//   - ScopeLiveChat: Read live chat messages
//   - ScopeLiveChatModerate: Moderate live chat (ban, delete, etc.)
//   - ScopeReadOnly: Read-only access to YouTube data
//   - ScopeUpload: Upload videos
package auth
