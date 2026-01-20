package auth

import (
	"encoding/json"
	"testing"
	"time"
)

func TestToken_Valid(t *testing.T) {
	tests := []struct {
		name  string
		token *Token
		want  bool
	}{
		{
			name:  "nil token",
			token: nil,
			want:  false,
		},
		{
			name:  "empty access token",
			token: &Token{},
			want:  false,
		},
		{
			name: "valid token no expiry",
			token: &Token{
				AccessToken: "test-token",
			},
			want: true,
		},
		{
			name: "valid token with future expiry",
			token: &Token{
				AccessToken: "test-token",
				Expiry:      time.Now().Add(1 * time.Hour),
			},
			want: true,
		},
		{
			name: "expired token",
			token: &Token{
				AccessToken: "test-token",
				Expiry:      time.Now().Add(-1 * time.Hour),
			},
			want: false,
		},
		{
			name: "token expiring within delta",
			token: &Token{
				AccessToken: "test-token",
				Expiry:      time.Now().Add(5 * time.Second), // Within expiryDelta
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.token.Valid()
			if got != tt.want {
				t.Errorf("Valid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToken_Expired(t *testing.T) {
	tests := []struct {
		name  string
		token *Token
		want  bool
	}{
		{
			name:  "nil token",
			token: nil,
			want:  true,
		},
		{
			name:  "no expiry",
			token: &Token{AccessToken: "test"},
			want:  false,
		},
		{
			name: "future expiry",
			token: &Token{
				AccessToken: "test",
				Expiry:      time.Now().Add(1 * time.Hour),
			},
			want: false,
		},
		{
			name: "past expiry",
			token: &Token{
				AccessToken: "test",
				Expiry:      time.Now().Add(-1 * time.Hour),
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.token.Expired()
			if got != tt.want {
				t.Errorf("Expired() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToken_SetAccessToken(t *testing.T) {
	token := &Token{}
	expiry := time.Now().Add(1 * time.Hour)

	token.SetAccessToken("new-token", expiry)

	if token.AccessToken != "new-token" {
		t.Errorf("AccessToken = %q, want %q", token.AccessToken, "new-token")
	}
	if !token.Expiry.Equal(expiry) {
		t.Errorf("Expiry = %v, want %v", token.Expiry, expiry)
	}
}

func TestToken_GetAccessToken(t *testing.T) {
	t.Run("nil token", func(t *testing.T) {
		var token *Token
		got := token.GetAccessToken()
		if got != "" {
			t.Errorf("GetAccessToken() = %q, want empty", got)
		}
	})

	t.Run("with token", func(t *testing.T) {
		token := &Token{AccessToken: "test-token"}
		got := token.GetAccessToken()
		if got != "test-token" {
			t.Errorf("GetAccessToken() = %q, want %q", got, "test-token")
		}
	})
}

func TestToken_GetRefreshToken(t *testing.T) {
	t.Run("nil token", func(t *testing.T) {
		var token *Token
		got := token.GetRefreshToken()
		if got != "" {
			t.Errorf("GetRefreshToken() = %q, want empty", got)
		}
	})

	t.Run("with token", func(t *testing.T) {
		token := &Token{RefreshToken: "refresh-token"}
		got := token.GetRefreshToken()
		if got != "refresh-token" {
			t.Errorf("GetRefreshToken() = %q, want %q", got, "refresh-token")
		}
	})
}

func TestToken_ExpiresIn(t *testing.T) {
	t.Run("nil token", func(t *testing.T) {
		var token *Token
		got := token.ExpiresIn()
		if got != 0 {
			t.Errorf("ExpiresIn() = %v, want 0", got)
		}
	})

	t.Run("no expiry", func(t *testing.T) {
		token := &Token{AccessToken: "test"}
		got := token.ExpiresIn()
		if got != 0 {
			t.Errorf("ExpiresIn() = %v, want 0", got)
		}
	})

	t.Run("future expiry", func(t *testing.T) {
		token := &Token{
			AccessToken: "test",
			Expiry:      time.Now().Add(1 * time.Hour),
		}
		got := token.ExpiresIn()
		// Should be approximately 1 hour (within a second tolerance)
		if got < 59*time.Minute || got > 61*time.Minute {
			t.Errorf("ExpiresIn() = %v, want approximately 1h", got)
		}
	})

	t.Run("past expiry", func(t *testing.T) {
		token := &Token{
			AccessToken: "test",
			Expiry:      time.Now().Add(-1 * time.Hour),
		}
		got := token.ExpiresIn()
		if got != 0 {
			t.Errorf("ExpiresIn() = %v, want 0", got)
		}
	})
}

func TestToken_Clone(t *testing.T) {
	t.Run("nil token", func(t *testing.T) {
		var token *Token
		got := token.Clone()
		if got != nil {
			t.Errorf("Clone() = %v, want nil", got)
		}
	})

	t.Run("clones all fields", func(t *testing.T) {
		expiry := time.Now().Add(1 * time.Hour)
		original := &Token{
			AccessToken:  "access",
			TokenType:    "Bearer",
			RefreshToken: "refresh",
			Expiry:       expiry,
			Scopes:       []string{"scope1", "scope2"},
		}

		clone := original.Clone()

		if clone.AccessToken != original.AccessToken {
			t.Errorf("AccessToken mismatch")
		}
		if clone.TokenType != original.TokenType {
			t.Errorf("TokenType mismatch")
		}
		if clone.RefreshToken != original.RefreshToken {
			t.Errorf("RefreshToken mismatch")
		}
		if !clone.Expiry.Equal(original.Expiry) {
			t.Errorf("Expiry mismatch")
		}
		if len(clone.Scopes) != len(original.Scopes) {
			t.Errorf("Scopes length mismatch")
		}

		// Verify it's a deep copy
		clone.Scopes[0] = "modified"
		if original.Scopes[0] == "modified" {
			t.Error("Clone did not create a deep copy of Scopes")
		}
	})
}

func TestToken_MarshalJSON(t *testing.T) {
	expiry := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	token := &Token{
		AccessToken:  "access-token",
		TokenType:    "Bearer",
		RefreshToken: "refresh-token",
		Expiry:       expiry,
		Scopes:       []string{"scope1"},
	}

	data, err := json.Marshal(token)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if result["access_token"] != "access-token" {
		t.Errorf("access_token = %v, want 'access-token'", result["access_token"])
	}
	if result["token_type"] != "Bearer" {
		t.Errorf("token_type = %v, want 'Bearer'", result["token_type"])
	}
	if result["expiry"] != "2024-01-01T12:00:00Z" {
		t.Errorf("expiry = %v, want '2024-01-01T12:00:00Z'", result["expiry"])
	}
}

func TestToken_UnmarshalJSON(t *testing.T) {
	t.Run("with expiry string", func(t *testing.T) {
		data := `{
			"access_token": "access",
			"token_type": "Bearer",
			"refresh_token": "refresh",
			"expiry": "2024-01-01T12:00:00Z",
			"scopes": ["scope1", "scope2"]
		}`

		var token Token
		if err := json.Unmarshal([]byte(data), &token); err != nil {
			t.Fatalf("Unmarshal error: %v", err)
		}

		if token.AccessToken != "access" {
			t.Errorf("AccessToken = %q, want 'access'", token.AccessToken)
		}
		if token.TokenType != "Bearer" {
			t.Errorf("TokenType = %q, want 'Bearer'", token.TokenType)
		}
		if token.RefreshToken != "refresh" {
			t.Errorf("RefreshToken = %q, want 'refresh'", token.RefreshToken)
		}
		if len(token.Scopes) != 2 {
			t.Errorf("Scopes length = %d, want 2", len(token.Scopes))
		}
	})

	t.Run("with expires_in", func(t *testing.T) {
		data := `{
			"access_token": "access",
			"expires_in": 3600
		}`

		before := time.Now()
		var token Token
		if err := json.Unmarshal([]byte(data), &token); err != nil {
			t.Fatalf("Unmarshal error: %v", err)
		}
		after := time.Now()

		if token.AccessToken != "access" {
			t.Errorf("AccessToken = %q, want 'access'", token.AccessToken)
		}

		// Expiry should be approximately 1 hour from now
		expectedLow := before.Add(3600 * time.Second)
		expectedHigh := after.Add(3600 * time.Second)
		if token.Expiry.Before(expectedLow) || token.Expiry.After(expectedHigh) {
			t.Errorf("Expiry = %v, want between %v and %v", token.Expiry, expectedLow, expectedHigh)
		}
	})
}

func TestTokenResponse_ToToken(t *testing.T) {
	tr := &tokenResponse{
		AccessToken:  "access",
		TokenType:    "Bearer",
		ExpiresIn:    3600,
		RefreshToken: "refresh",
	}

	before := time.Now()
	token := tr.toToken()
	after := time.Now()

	if token.AccessToken != "access" {
		t.Errorf("AccessToken = %q, want 'access'", token.AccessToken)
	}
	if token.TokenType != "Bearer" {
		t.Errorf("TokenType = %q, want 'Bearer'", token.TokenType)
	}
	if token.RefreshToken != "refresh" {
		t.Errorf("RefreshToken = %q, want 'refresh'", token.RefreshToken)
	}

	expectedLow := before.Add(3600 * time.Second)
	expectedHigh := after.Add(3600 * time.Second)
	if token.Expiry.Before(expectedLow) || token.Expiry.After(expectedHigh) {
		t.Errorf("Expiry = %v, want between %v and %v", token.Expiry, expectedLow, expectedHigh)
	}
}

func TestTokenResponse_ToToken_NoExpiry(t *testing.T) {
	tr := &tokenResponse{
		AccessToken: "access",
		TokenType:   "Bearer",
		ExpiresIn:   0,
	}

	token := tr.toToken()

	if !token.Expiry.IsZero() {
		t.Errorf("Expiry = %v, want zero", token.Expiry)
	}
}
