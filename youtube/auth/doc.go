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
// # Device Code Flow
//
// For devices with limited input capabilities (TVs, consoles, CLI apps):
//
//	deviceClient := auth.NewDeviceClient(auth.DeviceConfig{
//		ClientID: "your-client-id",
//		Scopes:   []string{auth.ScopeLiveChat},
//	})
//
//	// Request device code
//	authResp, err := deviceClient.RequestDeviceCode(ctx)
//	fmt.Printf("Go to %s and enter code: %s\n",
//		authResp.VerificationURL, authResp.UserCode)
//
//	// Poll for authorization
//	token, err := deviceClient.PollForToken(ctx, authResp)
//
// # Service Account Authentication
//
// For server-to-server communication without user interaction:
//
//	// From JSON credentials file
//	jsonData, _ := os.ReadFile("service-account.json")
//	client, err := auth.NewServiceAccountClientFromJSON(jsonData,
//		[]string{auth.ScopeReadOnly})
//
//	// Or from config
//	client, err := auth.NewServiceAccountClient(auth.ServiceAccountConfig{
//		Email:      "service@project.iam.gserviceaccount.com",
//		PrivateKey: pemKey,
//		Scopes:     []string{auth.ScopeReadOnly},
//	})
//
//	// Get access token (auto-refreshes)
//	token, err := client.AccessToken(ctx)
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
//   - ScopePartner: YouTube Analytics access
package auth
