package streaming

import (
	"encoding/json"
	"testing"
	"time"
)

func TestLiveChatMessageListResponse_PollingInterval(t *testing.T) {
	tests := []struct {
		name   string
		millis int
		want   time.Duration
	}{
		{"zero", 0, 0},
		{"one second", 1000, 1 * time.Second},
		{"five seconds", 5000, 5 * time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &LiveChatMessageListResponse{PollingIntervalMillis: tt.millis}
			got := resp.PollingInterval()
			if got != tt.want {
				t.Errorf("PollingInterval() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLiveChatMessageListResponse_IsChatEnded(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name      string
		offlineAt *time.Time
		want      bool
	}{
		{"not ended", nil, false},
		{"ended", &now, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &LiveChatMessageListResponse{OfflineAt: tt.offlineAt}
			got := resp.IsChatEnded()
			if got != tt.want {
				t.Errorf("IsChatEnded() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLiveChatMessage_Type(t *testing.T) {
	tests := []struct {
		name    string
		snippet *MessageSnippet
		want    string
	}{
		{"nil snippet", nil, ""},
		{"text message", &MessageSnippet{Type: MessageTypeText}, MessageTypeText},
		{"super chat", &MessageSnippet{Type: MessageTypeSuperChat}, MessageTypeSuperChat},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := &LiveChatMessage{Snippet: tt.snippet}
			got := msg.Type()
			if got != tt.want {
				t.Errorf("Type() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestLiveChatMessage_Message(t *testing.T) {
	tests := []struct {
		name    string
		snippet *MessageSnippet
		want    string
	}{
		{"nil snippet", nil, ""},
		{"with message", &MessageSnippet{DisplayMessage: "hello"}, "hello"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := &LiveChatMessage{Snippet: tt.snippet}
			got := msg.Message()
			if got != tt.want {
				t.Errorf("Message() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestLiveChatMessage_TypeCheckers(t *testing.T) {
	tests := []struct {
		name              string
		msgType           string
		isText            bool
		isSuperChat       bool
		isSuperSticker    bool
		isMembership      bool
		isMemberMilestone bool
		isGiftMembership  bool
		isPoll            bool
	}{
		{
			name:    "text message",
			msgType: MessageTypeText,
			isText:  true,
		},
		{
			name:        "super chat",
			msgType:     MessageTypeSuperChat,
			isSuperChat: true,
		},
		{
			name:           "super sticker",
			msgType:        MessageTypeSuperSticker,
			isSuperSticker: true,
		},
		{
			name:         "membership",
			msgType:      MessageTypeMembership,
			isMembership: true,
		},
		{
			name:              "member milestone",
			msgType:           MessageTypeMemberMilestone,
			isMemberMilestone: true,
		},
		{
			name:             "gift membership",
			msgType:          MessageTypeMembershipGifting,
			isGiftMembership: true,
		},
		{
			name:    "poll event",
			msgType: MessageTypePoll,
			isPoll:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := &LiveChatMessage{Snippet: &MessageSnippet{Type: tt.msgType}}

			if got := msg.IsTextMessage(); got != tt.isText {
				t.Errorf("IsTextMessage() = %v, want %v", got, tt.isText)
			}
			if got := msg.IsSuperChat(); got != tt.isSuperChat {
				t.Errorf("IsSuperChat() = %v, want %v", got, tt.isSuperChat)
			}
			if got := msg.IsSuperSticker(); got != tt.isSuperSticker {
				t.Errorf("IsSuperSticker() = %v, want %v", got, tt.isSuperSticker)
			}
			if got := msg.IsMembership(); got != tt.isMembership {
				t.Errorf("IsMembership() = %v, want %v", got, tt.isMembership)
			}
			if got := msg.IsMemberMilestone(); got != tt.isMemberMilestone {
				t.Errorf("IsMemberMilestone() = %v, want %v", got, tt.isMemberMilestone)
			}
			if got := msg.IsGiftMembership(); got != tt.isGiftMembership {
				t.Errorf("IsGiftMembership() = %v, want %v", got, tt.isGiftMembership)
			}
			if got := msg.IsPoll(); got != tt.isPoll {
				t.Errorf("IsPoll() = %v, want %v", got, tt.isPoll)
			}
		})
	}
}

func TestLiveChatMessage_Clone(t *testing.T) {
	t.Run("nil message", func(t *testing.T) {
		var msg *LiveChatMessage
		got := msg.Clone()
		if got != nil {
			t.Errorf("Clone() of nil = %v, want nil", got)
		}
	})

	t.Run("full message", func(t *testing.T) {
		original := &LiveChatMessage{
			Kind: "youtube#liveChatMessage",
			ID:   "msg123",
			Snippet: &MessageSnippet{
				Type:           MessageTypeText,
				LiveChatID:     "chat123",
				DisplayMessage: "hello world",
				TextMessageDetails: &TextMessageDetails{
					MessageText: "hello world",
				},
			},
			AuthorDetails: &AuthorDetails{
				ChannelID:       "channel123",
				DisplayName:     "Test User",
				IsChatModerator: true,
			},
		}

		clone := original.Clone()
		if clone == original {
			t.Error("Clone() returned same pointer")
		}
		if clone.ID != original.ID {
			t.Errorf("Clone().ID = %q, want %q", clone.ID, original.ID)
		}
		if clone.Snippet.DisplayMessage != original.Snippet.DisplayMessage {
			t.Errorf("Clone().Snippet.DisplayMessage = %q, want %q",
				clone.Snippet.DisplayMessage, original.Snippet.DisplayMessage)
		}
		if clone.AuthorDetails.DisplayName != original.AuthorDetails.DisplayName {
			t.Errorf("Clone().AuthorDetails.DisplayName = %q, want %q",
				clone.AuthorDetails.DisplayName, original.AuthorDetails.DisplayName)
		}

		// Modify clone and verify original unchanged
		clone.ID = "modified"
		if original.ID == "modified" {
			t.Error("Clone() did not create deep copy")
		}
	})
}

func TestLiveChatMessage_JSON(t *testing.T) {
	jsonData := `{
		"kind": "youtube#liveChatMessage",
		"etag": "abc123",
		"id": "msg123",
		"snippet": {
			"type": "textMessageEvent",
			"liveChatId": "chat123",
			"authorChannelId": "channel123",
			"publishedAt": "2024-01-15T10:30:00Z",
			"hasDisplayContent": true,
			"displayMessage": "Hello world!",
			"textMessageDetails": {
				"messageText": "Hello world!"
			}
		},
		"authorDetails": {
			"channelId": "channel123",
			"channelUrl": "https://www.youtube.com/channel/channel123",
			"displayName": "Test User",
			"profileImageUrl": "https://example.com/image.jpg",
			"isVerified": false,
			"isChatOwner": false,
			"isChatSponsor": true,
			"isChatModerator": false
		}
	}`

	var msg LiveChatMessage
	err := json.Unmarshal([]byte(jsonData), &msg)
	if err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if msg.ID != "msg123" {
		t.Errorf("ID = %q, want 'msg123'", msg.ID)
	}
	if msg.Snippet.Type != MessageTypeText {
		t.Errorf("Snippet.Type = %q, want %q", msg.Snippet.Type, MessageTypeText)
	}
	if msg.Snippet.DisplayMessage != "Hello world!" {
		t.Errorf("Snippet.DisplayMessage = %q, want 'Hello world!'", msg.Snippet.DisplayMessage)
	}
	if msg.AuthorDetails.DisplayName != "Test User" {
		t.Errorf("AuthorDetails.DisplayName = %q, want 'Test User'", msg.AuthorDetails.DisplayName)
	}
	if !msg.AuthorDetails.IsChatSponsor {
		t.Error("AuthorDetails.IsChatSponsor = false, want true")
	}
}

func TestSuperChatDetails_JSON(t *testing.T) {
	jsonData := `{
		"snippet": {
			"type": "superChatEvent",
			"superChatDetails": {
				"amountMicros": "5000000",
				"currency": "USD",
				"amountDisplayString": "$5.00",
				"userComment": "Great stream!",
				"tier": 3
			}
		}
	}`

	var msg LiveChatMessage
	err := json.Unmarshal([]byte(jsonData), &msg)
	if err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	sc := msg.Snippet.SuperChatDetails
	if sc == nil {
		t.Fatal("SuperChatDetails is nil")
	}
	if sc.AmountMicros != 5000000 {
		t.Errorf("AmountMicros = %d, want 5000000", sc.AmountMicros)
	}
	if sc.Currency != "USD" {
		t.Errorf("Currency = %q, want 'USD'", sc.Currency)
	}
	if sc.AmountDisplayString != "$5.00" {
		t.Errorf("AmountDisplayString = %q, want '$5.00'", sc.AmountDisplayString)
	}
	if sc.UserComment != "Great stream!" {
		t.Errorf("UserComment = %q, want 'Great stream!'", sc.UserComment)
	}
	if sc.Tier != 3 {
		t.Errorf("Tier = %d, want 3", sc.Tier)
	}
}

func TestUserBannedDetails_JSON(t *testing.T) {
	jsonData := `{
		"snippet": {
			"type": "userBannedEvent",
			"userBannedDetails": {
				"bannedUserDetails": {
					"channelId": "banned123",
					"displayName": "Bad User"
				},
				"banType": "temporary",
				"banDurationSeconds": "300"
			}
		}
	}`

	var msg LiveChatMessage
	err := json.Unmarshal([]byte(jsonData), &msg)
	if err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	ban := msg.Snippet.UserBannedDetails
	if ban == nil {
		t.Fatal("UserBannedDetails is nil")
	}
	if ban.BanType != BanTypeTemporary {
		t.Errorf("BanType = %q, want %q", ban.BanType, BanTypeTemporary)
	}
	if ban.BanDurationSeconds != 300 {
		t.Errorf("BanDurationSeconds = %d, want 300", ban.BanDurationSeconds)
	}
	if ban.BannedUserDetails.ChannelID != "banned123" {
		t.Errorf("BannedUserDetails.ChannelID = %q, want 'banned123'", ban.BannedUserDetails.ChannelID)
	}
}

func TestMemberMilestoneChatDetails_JSON(t *testing.T) {
	jsonData := `{
		"snippet": {
			"type": "memberMilestoneChatEvent",
			"memberMilestoneChatDetails": {
				"memberLevelName": "Sponsor",
				"memberMonth": 12,
				"userComment": "One year anniversary!"
			}
		}
	}`

	var msg LiveChatMessage
	err := json.Unmarshal([]byte(jsonData), &msg)
	if err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	milestone := msg.Snippet.MemberMilestoneChatDetails
	if milestone == nil {
		t.Fatal("MemberMilestoneChatDetails is nil")
	}
	if milestone.MemberLevelName != "Sponsor" {
		t.Errorf("MemberLevelName = %q, want 'Sponsor'", milestone.MemberLevelName)
	}
	if milestone.MemberMonth != 12 {
		t.Errorf("MemberMonth = %d, want 12", milestone.MemberMonth)
	}
	if milestone.UserComment != "One year anniversary!" {
		t.Errorf("UserComment = %q, want 'One year anniversary!'", milestone.UserComment)
	}
}

func TestNewSponsorDetails_JSON(t *testing.T) {
	jsonData := `{
		"snippet": {
			"type": "newSponsorEvent",
			"newSponsorDetails": {
				"memberLevelName": "Premium",
				"isUpgrade": true
			}
		}
	}`

	var msg LiveChatMessage
	err := json.Unmarshal([]byte(jsonData), &msg)
	if err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	sponsor := msg.Snippet.NewSponsorDetails
	if sponsor == nil {
		t.Fatal("NewSponsorDetails is nil")
	}
	if sponsor.MemberLevelName != "Premium" {
		t.Errorf("MemberLevelName = %q, want 'Premium'", sponsor.MemberLevelName)
	}
	if !sponsor.IsUpgrade {
		t.Error("IsUpgrade = false, want true")
	}
}

func TestMessageDeletedDetails_JSON(t *testing.T) {
	jsonData := `{
		"snippet": {
			"type": "messageDeletedEvent",
			"messageDeletedDetails": {
				"deletedMessageId": "deleted123"
			}
		}
	}`

	var msg LiveChatMessage
	err := json.Unmarshal([]byte(jsonData), &msg)
	if err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	deleted := msg.Snippet.MessageDeletedDetails
	if deleted == nil {
		t.Fatal("MessageDeletedDetails is nil")
	}
	if deleted.DeletedMessageID != "deleted123" {
		t.Errorf("DeletedMessageID = %q, want 'deleted123'", deleted.DeletedMessageID)
	}
}

func TestLiveChatMessageListResponse_JSON(t *testing.T) {
	jsonData := `{
		"kind": "youtube#liveChatMessageListResponse",
		"etag": "xyz789",
		"nextPageToken": "page2",
		"pollingIntervalMillis": 5000,
		"pageInfo": {
			"totalResults": 100,
			"resultsPerPage": 25
		},
		"items": [
			{
				"id": "msg1",
				"snippet": {
					"type": "textMessageEvent",
					"displayMessage": "First message"
				}
			},
			{
				"id": "msg2",
				"snippet": {
					"type": "textMessageEvent",
					"displayMessage": "Second message"
				}
			}
		]
	}`

	var resp LiveChatMessageListResponse
	err := json.Unmarshal([]byte(jsonData), &resp)
	if err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if resp.NextPageToken != "page2" {
		t.Errorf("NextPageToken = %q, want 'page2'", resp.NextPageToken)
	}
	if resp.PollingIntervalMillis != 5000 {
		t.Errorf("PollingIntervalMillis = %d, want 5000", resp.PollingIntervalMillis)
	}
	if resp.PollingInterval() != 5*time.Second {
		t.Errorf("PollingInterval() = %v, want 5s", resp.PollingInterval())
	}
	if len(resp.Items) != 2 {
		t.Errorf("len(Items) = %d, want 2", len(resp.Items))
	}
	if resp.Items[0].ID != "msg1" {
		t.Errorf("Items[0].ID = %q, want 'msg1'", resp.Items[0].ID)
	}
	if resp.PageInfo.TotalResults != 100 {
		t.Errorf("PageInfo.TotalResults = %d, want 100", resp.PageInfo.TotalResults)
	}
}

func TestInsertMessageRequest_JSON(t *testing.T) {
	req := &InsertMessageRequest{
		Snippet: &InsertMessageSnippet{
			LiveChatID: "chat123",
			Type:       MessageTypeText,
			TextMessageDetails: &TextMessageDetails{
				MessageText: "Hello!",
			},
		},
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	snippet, ok := parsed["snippet"].(map[string]any)
	if !ok {
		t.Fatal("snippet not found in JSON")
	}
	if snippet["liveChatId"] != "chat123" {
		t.Errorf("liveChatId = %q, want 'chat123'", snippet["liveChatId"])
	}
	if snippet["type"] != MessageTypeText {
		t.Errorf("type = %q, want %q", snippet["type"], MessageTypeText)
	}
}

func TestInsertBanRequest_JSON(t *testing.T) {
	req := &InsertBanRequest{
		Snippet: &InsertBanSnippet{
			LiveChatID:         "chat123",
			Type:               BanTypeTemporary,
			BanDurationSeconds: 300,
			BannedUserDetails: &BannedUserDetails{
				ChannelID: "user123",
			},
		},
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	snippet, ok := parsed["snippet"].(map[string]any)
	if !ok {
		t.Fatal("snippet not found in JSON")
	}
	if snippet["type"] != BanTypeTemporary {
		t.Errorf("type = %q, want %q", snippet["type"], BanTypeTemporary)
	}
	if snippet["banDurationSeconds"] != float64(300) {
		t.Errorf("banDurationSeconds = %v, want 300", snippet["banDurationSeconds"])
	}
}

func TestMessageTypeConstants(t *testing.T) {
	// Verify message type constants match expected YouTube API values
	expectedTypes := map[string]string{
		"text":                   "textMessageEvent",
		"superChat":              "superChatEvent",
		"superSticker":           "superStickerEvent",
		"membership":             "newSponsorEvent",
		"memberMilestone":        "memberMilestoneChatEvent",
		"giftMembershipReceived": "giftMembershipReceivedEvent",
		"membershipGifting":      "membershipGiftingEvent",
		"chatEnded":              "chatEndedEvent",
		"messageDeleted":         "messageDeletedEvent",
		"userBanned":             "userBannedEvent",
	}

	actualTypes := map[string]string{
		"text":                   MessageTypeText,
		"superChat":              MessageTypeSuperChat,
		"superSticker":           MessageTypeSuperSticker,
		"membership":             MessageTypeMembership,
		"memberMilestone":        MessageTypeMemberMilestone,
		"giftMembershipReceived": MessageTypeGiftMembershipReceived,
		"membershipGifting":      MessageTypeMembershipGifting,
		"chatEnded":              MessageTypeChatEnded,
		"messageDeleted":         MessageTypeMessageDeleted,
		"userBanned":             MessageTypeUserBanned,
	}

	for key, expected := range expectedTypes {
		actual, ok := actualTypes[key]
		if !ok {
			t.Errorf("Missing type constant for %q", key)
			continue
		}
		if actual != expected {
			t.Errorf("MessageType%s = %q, want %q", key, actual, expected)
		}
	}
}

func TestBanTypeConstants(t *testing.T) {
	if BanTypePermanent != "permanent" {
		t.Errorf("BanTypePermanent = %q, want 'permanent'", BanTypePermanent)
	}
	if BanTypeTemporary != "temporary" {
		t.Errorf("BanTypeTemporary = %q, want 'temporary'", BanTypeTemporary)
	}
}

func TestPollStatusConstants(t *testing.T) {
	if PollStatusOpen != "open" {
		t.Errorf("PollStatusOpen = %q, want 'open'", PollStatusOpen)
	}
	if PollStatusClosed != "closed" {
		t.Errorf("PollStatusClosed = %q, want 'closed'", PollStatusClosed)
	}
}

func TestLiveChatMessage_Clone_SuperSticker(t *testing.T) {
	original := &LiveChatMessage{
		ID: "sticker123",
		Snippet: &MessageSnippet{
			Type: MessageTypeSuperSticker,
			SuperStickerDetails: &SuperStickerDetails{
				SuperStickerID: "sticker-001",
				SuperStickerMetadata: &SuperStickerMetadata{
					StickerID: "meta-001",
					AltText:   "Happy sticker",
				},
				AmountMicros:        1000000,
				Currency:            "USD",
				AmountDisplayString: "$1.00",
				Tier:                1,
			},
		},
	}

	clone := original.Clone()
	if clone == original {
		t.Error("Clone() returned same pointer")
	}
	if clone.Snippet.SuperStickerDetails == original.Snippet.SuperStickerDetails {
		t.Error("Clone() did not deep copy SuperStickerDetails")
	}
	if clone.Snippet.SuperStickerDetails.SuperStickerMetadata == original.Snippet.SuperStickerDetails.SuperStickerMetadata {
		t.Error("Clone() did not deep copy SuperStickerMetadata")
	}
	if clone.Snippet.SuperStickerDetails.SuperStickerID != "sticker-001" {
		t.Errorf("SuperStickerID = %q, want 'sticker-001'", clone.Snippet.SuperStickerDetails.SuperStickerID)
	}
	if clone.Snippet.SuperStickerDetails.SuperStickerMetadata.AltText != "Happy sticker" {
		t.Errorf("AltText = %q, want 'Happy sticker'", clone.Snippet.SuperStickerDetails.SuperStickerMetadata.AltText)
	}

	// Modify clone and verify original unchanged
	clone.Snippet.SuperStickerDetails.SuperStickerMetadata.AltText = "modified"
	if original.Snippet.SuperStickerDetails.SuperStickerMetadata.AltText == "modified" {
		t.Error("Clone() did not create deep copy of SuperStickerMetadata")
	}
}

func TestLiveChatMessage_Clone_UserBanned(t *testing.T) {
	original := &LiveChatMessage{
		ID: "ban123",
		Snippet: &MessageSnippet{
			Type: MessageTypeUserBanned,
			UserBannedDetails: &UserBannedDetails{
				BannedUserDetails: &BannedUserDetails{
					ChannelID:       "banned-channel",
					ChannelURL:      "https://youtube.com/channel/banned",
					DisplayName:     "Banned User",
					ProfileImageURL: "https://example.com/image.png",
				},
				BanType:            BanTypeTemporary,
				BanDurationSeconds: 300,
			},
		},
	}

	clone := original.Clone()
	if clone == original {
		t.Error("Clone() returned same pointer")
	}
	if clone.Snippet.UserBannedDetails == original.Snippet.UserBannedDetails {
		t.Error("Clone() did not deep copy UserBannedDetails")
	}
	if clone.Snippet.UserBannedDetails.BannedUserDetails == original.Snippet.UserBannedDetails.BannedUserDetails {
		t.Error("Clone() did not deep copy BannedUserDetails")
	}
	if clone.Snippet.UserBannedDetails.BanType != BanTypeTemporary {
		t.Errorf("BanType = %q, want %q", clone.Snippet.UserBannedDetails.BanType, BanTypeTemporary)
	}
	if clone.Snippet.UserBannedDetails.BannedUserDetails.DisplayName != "Banned User" {
		t.Errorf("DisplayName = %q, want 'Banned User'", clone.Snippet.UserBannedDetails.BannedUserDetails.DisplayName)
	}

	// Modify clone and verify original unchanged
	clone.Snippet.UserBannedDetails.BannedUserDetails.DisplayName = "modified"
	if original.Snippet.UserBannedDetails.BannedUserDetails.DisplayName == "modified" {
		t.Error("Clone() did not create deep copy of BannedUserDetails")
	}
}

func TestLiveChatMessage_Clone_Poll(t *testing.T) {
	original := &LiveChatMessage{
		ID: "poll123",
		Snippet: &MessageSnippet{
			Type: MessageTypePoll,
			PollDetails: &PollDetails{
				Question: "Favorite color?",
				Status:   PollStatusOpen,
				Choices: []PollChoice{
					{ChoiceID: "a", Text: "Red", NumVotes: 10},
					{ChoiceID: "b", Text: "Blue", NumVotes: 15},
					{ChoiceID: "c", Text: "Green", NumVotes: 5},
				},
			},
		},
	}

	clone := original.Clone()
	if clone == original {
		t.Error("Clone() returned same pointer")
	}
	if clone.Snippet.PollDetails == original.Snippet.PollDetails {
		t.Error("Clone() did not deep copy PollDetails")
	}
	if len(clone.Snippet.PollDetails.Choices) != 3 {
		t.Errorf("len(Choices) = %d, want 3", len(clone.Snippet.PollDetails.Choices))
	}
	if clone.Snippet.PollDetails.Question != "Favorite color?" {
		t.Errorf("Question = %q, want 'Favorite color?'", clone.Snippet.PollDetails.Question)
	}
	if clone.Snippet.PollDetails.Choices[1].Text != "Blue" {
		t.Errorf("Choices[1].Text = %q, want 'Blue'", clone.Snippet.PollDetails.Choices[1].Text)
	}

	// Modify clone's choices and verify original unchanged
	clone.Snippet.PollDetails.Choices[0].Text = "modified"
	if original.Snippet.PollDetails.Choices[0].Text == "modified" {
		t.Error("Clone() did not create deep copy of Choices slice")
	}
}

func TestLiveChatMessage_Clone_EmptyPollChoices(t *testing.T) {
	// Test poll with empty choices slice
	original := &LiveChatMessage{
		ID: "poll-empty",
		Snippet: &MessageSnippet{
			Type: MessageTypePoll,
			PollDetails: &PollDetails{
				Question: "Empty poll",
				Status:   PollStatusOpen,
				Choices:  nil,
			},
		},
	}

	clone := original.Clone()
	if clone.Snippet.PollDetails.Choices != nil {
		t.Errorf("Choices = %v, want nil", clone.Snippet.PollDetails.Choices)
	}
}

func TestLiveChatMessage_IsPoll(t *testing.T) {
	tests := []struct {
		name    string
		msgType string
		want    bool
	}{
		{"poll event", MessageTypePoll, true},
		{"text message", MessageTypeText, false},
		{"super chat", MessageTypeSuperChat, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := &LiveChatMessage{Snippet: &MessageSnippet{Type: tt.msgType}}
			if got := msg.IsPoll(); got != tt.want {
				t.Errorf("IsPoll() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLiveChatMessage_HasActivePoll(t *testing.T) {
	tests := []struct {
		name string
		msg  *LiveChatMessage
		want bool
	}{
		{
			name: "nil snippet",
			msg:  &LiveChatMessage{},
			want: false,
		},
		{
			name: "no active poll",
			msg:  &LiveChatMessage{Snippet: &MessageSnippet{}},
			want: false,
		},
		{
			name: "has active poll",
			msg: &LiveChatMessage{
				Snippet: &MessageSnippet{
					ActivePollItem: &LiveChatMessage{
						ID: "poll123",
						Snippet: &MessageSnippet{
							Type: MessageTypePoll,
							PollDetails: &PollDetails{
								Question: "Test poll?",
								Status:   PollStatusOpen,
							},
						},
					},
				},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.msg.HasActivePoll(); got != tt.want {
				t.Errorf("HasActivePoll() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLiveChatMessage_ActivePoll(t *testing.T) {
	t.Run("nil snippet", func(t *testing.T) {
		msg := &LiveChatMessage{}
		if got := msg.ActivePoll(); got != nil {
			t.Errorf("ActivePoll() = %v, want nil", got)
		}
	})

	t.Run("no active poll", func(t *testing.T) {
		msg := &LiveChatMessage{Snippet: &MessageSnippet{}}
		if got := msg.ActivePoll(); got != nil {
			t.Errorf("ActivePoll() = %v, want nil", got)
		}
	})

	t.Run("has active poll", func(t *testing.T) {
		activePoll := &LiveChatMessage{
			ID: "poll123",
			Snippet: &MessageSnippet{
				Type: MessageTypePoll,
				PollDetails: &PollDetails{
					Question: "Favorite color?",
					Status:   PollStatusOpen,
				},
			},
		}
		msg := &LiveChatMessage{
			Snippet: &MessageSnippet{
				ActivePollItem: activePoll,
			},
		}

		got := msg.ActivePoll()
		if got != activePoll {
			t.Errorf("ActivePoll() = %v, want %v", got, activePoll)
		}
		if got.ID != "poll123" {
			t.Errorf("ActivePoll().ID = %q, want 'poll123'", got.ID)
		}
		if got.Snippet.PollDetails.Question != "Favorite color?" {
			t.Errorf("ActivePoll().Snippet.PollDetails.Question = %q, want 'Favorite color?'",
				got.Snippet.PollDetails.Question)
		}
	})
}

func TestLiveChatMessage_Clone_ActivePollItem(t *testing.T) {
	original := &LiveChatMessage{
		ID: "msg123",
		Snippet: &MessageSnippet{
			Type: MessageTypeText,
			ActivePollItem: &LiveChatMessage{
				ID: "poll123",
				Snippet: &MessageSnippet{
					Type: MessageTypePoll,
					PollDetails: &PollDetails{
						Question: "Favorite color?",
						Status:   PollStatusOpen,
						Choices: []PollChoice{
							{ChoiceID: "a", Text: "Red"},
							{ChoiceID: "b", Text: "Blue"},
						},
					},
				},
			},
		},
	}

	clone := original.Clone()
	if clone == original {
		t.Error("Clone() returned same pointer")
	}
	if clone.Snippet.ActivePollItem == original.Snippet.ActivePollItem {
		t.Error("Clone() did not deep copy ActivePollItem")
	}
	if clone.Snippet.ActivePollItem.ID != "poll123" {
		t.Errorf("ActivePollItem.ID = %q, want 'poll123'", clone.Snippet.ActivePollItem.ID)
	}
	if clone.Snippet.ActivePollItem.Snippet.PollDetails.Question != "Favorite color?" {
		t.Errorf("ActivePollItem.Snippet.PollDetails.Question = %q, want 'Favorite color?'",
			clone.Snippet.ActivePollItem.Snippet.PollDetails.Question)
	}

	// Modify clone and verify original unchanged
	clone.Snippet.ActivePollItem.ID = "modified"
	if original.Snippet.ActivePollItem.ID == "modified" {
		t.Error("Clone() did not create deep copy of ActivePollItem")
	}
}

func TestActivePollItem_JSON(t *testing.T) {
	jsonData := `{
		"id": "msg123",
		"snippet": {
			"type": "textMessageEvent",
			"activePollItem": {
				"id": "poll123",
				"snippet": {
					"type": "pollEvent",
					"pollDetails": {
						"pollId": "poll-001",
						"question": "What's your favorite color?",
						"status": "open",
						"choices": [
							{"choiceId": "a", "text": "Red"},
							{"choiceId": "b", "text": "Blue"},
							{"choiceId": "c", "text": "Green"}
						]
					}
				}
			}
		}
	}`

	var msg LiveChatMessage
	err := json.Unmarshal([]byte(jsonData), &msg)
	if err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if !msg.HasActivePoll() {
		t.Fatal("HasActivePoll() = false, want true")
	}

	poll := msg.ActivePoll()
	if poll == nil {
		t.Fatal("ActivePoll() returned nil")
	}
	if poll.ID != "poll123" {
		t.Errorf("ActivePoll().ID = %q, want 'poll123'", poll.ID)
	}
	if poll.Snippet.Type != MessageTypePoll {
		t.Errorf("ActivePoll().Snippet.Type = %q, want %q", poll.Snippet.Type, MessageTypePoll)
	}
	if poll.Snippet.PollDetails.Question != "What's your favorite color?" {
		t.Errorf("PollDetails.Question = %q, want 'What's your favorite color?'",
			poll.Snippet.PollDetails.Question)
	}
	if len(poll.Snippet.PollDetails.Choices) != 3 {
		t.Errorf("len(Choices) = %d, want 3", len(poll.Snippet.PollDetails.Choices))
	}
}

func TestCloneFunctions_NilInputs(t *testing.T) {
	t.Run("cloneSuperChatDetails nil", func(t *testing.T) {
		result := cloneSuperChatDetails(nil)
		if result != nil {
			t.Error("cloneSuperChatDetails(nil) should return nil")
		}
	})

	t.Run("cloneSuperStickerMetadata nil", func(t *testing.T) {
		result := cloneSuperStickerMetadata(nil)
		if result != nil {
			t.Error("cloneSuperStickerMetadata(nil) should return nil")
		}
	})

	t.Run("cloneMemberMilestoneChatDetails nil", func(t *testing.T) {
		result := cloneMemberMilestoneChatDetails(nil)
		if result != nil {
			t.Error("cloneMemberMilestoneChatDetails(nil) should return nil")
		}
	})

	t.Run("cloneNewSponsorDetails nil", func(t *testing.T) {
		result := cloneNewSponsorDetails(nil)
		if result != nil {
			t.Error("cloneNewSponsorDetails(nil) should return nil")
		}
	})

	t.Run("cloneMembershipGiftingDetails nil", func(t *testing.T) {
		result := cloneMembershipGiftingDetails(nil)
		if result != nil {
			t.Error("cloneMembershipGiftingDetails(nil) should return nil")
		}
	})

	t.Run("cloneGiftMembershipReceivedDetails nil", func(t *testing.T) {
		result := cloneGiftMembershipReceivedDetails(nil)
		if result != nil {
			t.Error("cloneGiftMembershipReceivedDetails(nil) should return nil")
		}
	})

	t.Run("cloneMessageDeletedDetails nil", func(t *testing.T) {
		result := cloneMessageDeletedDetails(nil)
		if result != nil {
			t.Error("cloneMessageDeletedDetails(nil) should return nil")
		}
	})

	t.Run("cloneBannedUserDetails nil", func(t *testing.T) {
		result := cloneBannedUserDetails(nil)
		if result != nil {
			t.Error("cloneBannedUserDetails(nil) should return nil")
		}
	})

	t.Run("cloneFanFundingEventDetails nil", func(t *testing.T) {
		result := cloneFanFundingEventDetails(nil)
		if result != nil {
			t.Error("cloneFanFundingEventDetails(nil) should return nil")
		}
	})
}

func TestCloneFunctions_WithData(t *testing.T) {
	t.Run("cloneSuperChatDetails with data", func(t *testing.T) {
		original := &SuperChatDetails{
			AmountMicros:        5000000,
			Currency:            "USD",
			AmountDisplayString: "$5.00",
			Tier:                2,
		}
		result := cloneSuperChatDetails(original)
		if result == original {
			t.Error("cloneSuperChatDetails should return a copy, not the same pointer")
		}
		if result.AmountMicros != original.AmountMicros {
			t.Errorf("AmountMicros = %d, want %d", result.AmountMicros, original.AmountMicros)
		}
	})

	t.Run("cloneSuperStickerMetadata with data", func(t *testing.T) {
		original := &SuperStickerMetadata{
			StickerID:       "sticker123",
			AltText:         "Happy sticker",
			AltTextLanguage: "en",
		}
		result := cloneSuperStickerMetadata(original)
		if result == original {
			t.Error("cloneSuperStickerMetadata should return a copy, not the same pointer")
		}
		if result.StickerID != original.StickerID {
			t.Errorf("StickerID = %q, want %q", result.StickerID, original.StickerID)
		}
	})

	t.Run("cloneMemberMilestoneChatDetails with data", func(t *testing.T) {
		original := &MemberMilestoneChatDetails{
			MemberMonth: 12,
			UserComment: "12 months!",
		}
		result := cloneMemberMilestoneChatDetails(original)
		if result == original {
			t.Error("cloneMemberMilestoneChatDetails should return a copy, not the same pointer")
		}
		if result.MemberMonth != original.MemberMonth {
			t.Errorf("MemberMonth = %d, want %d", result.MemberMonth, original.MemberMonth)
		}
	})

	t.Run("cloneNewSponsorDetails with data", func(t *testing.T) {
		original := &NewSponsorDetails{
			MemberLevelName: "Gold Member",
			IsUpgrade:       true,
		}
		result := cloneNewSponsorDetails(original)
		if result == original {
			t.Error("cloneNewSponsorDetails should return a copy, not the same pointer")
		}
		if result.MemberLevelName != original.MemberLevelName {
			t.Errorf("MemberLevelName = %q, want %q", result.MemberLevelName, original.MemberLevelName)
		}
	})

	t.Run("cloneMembershipGiftingDetails with data", func(t *testing.T) {
		original := &MembershipGiftingDetails{
			GiftMembershipsCount: 5,
			MemberLevelName:      "Gold",
		}
		result := cloneMembershipGiftingDetails(original)
		if result == original {
			t.Error("cloneMembershipGiftingDetails should return a copy, not the same pointer")
		}
		if result.GiftMembershipsCount != original.GiftMembershipsCount {
			t.Errorf("GiftMembershipsCount = %d, want %d", result.GiftMembershipsCount, original.GiftMembershipsCount)
		}
	})

	t.Run("cloneGiftMembershipReceivedDetails with data", func(t *testing.T) {
		original := &GiftMembershipReceivedDetails{
			MemberLevelName:                      "Gold",
			GifterChannelID:                      "channel123",
			AssociatedMembershipGiftingMessageID: "gift123",
		}
		result := cloneGiftMembershipReceivedDetails(original)
		if result == original {
			t.Error("cloneGiftMembershipReceivedDetails should return a copy, not the same pointer")
		}
		if result.GifterChannelID != original.GifterChannelID {
			t.Errorf("GifterChannelID = %q, want %q", result.GifterChannelID, original.GifterChannelID)
		}
	})

	t.Run("cloneMessageDeletedDetails with data", func(t *testing.T) {
		original := &MessageDeletedDetails{
			DeletedMessageID: "msg123",
		}
		result := cloneMessageDeletedDetails(original)
		if result == original {
			t.Error("cloneMessageDeletedDetails should return a copy, not the same pointer")
		}
		if result.DeletedMessageID != original.DeletedMessageID {
			t.Errorf("DeletedMessageID = %q, want %q", result.DeletedMessageID, original.DeletedMessageID)
		}
	})

	t.Run("cloneBannedUserDetails with data", func(t *testing.T) {
		original := &BannedUserDetails{
			ChannelID:   "channel123",
			ChannelURL:  "https://youtube.com/channel/channel123",
			DisplayName: "Banned User",
		}
		result := cloneBannedUserDetails(original)
		if result == original {
			t.Error("cloneBannedUserDetails should return a copy, not the same pointer")
		}
		if result.ChannelID != original.ChannelID {
			t.Errorf("ChannelID = %q, want %q", result.ChannelID, original.ChannelID)
		}
	})

	t.Run("cloneFanFundingEventDetails with data", func(t *testing.T) {
		original := &FanFundingEventDetails{
			AmountMicros:        10000000,
			Currency:            "EUR",
			AmountDisplayString: "€10.00",
			UserComment:         "Great stream!",
		}
		result := cloneFanFundingEventDetails(original)
		if result == original {
			t.Error("cloneFanFundingEventDetails should return a copy, not the same pointer")
		}
		if result.AmountMicros != original.AmountMicros {
			t.Errorf("AmountMicros = %d, want %d", result.AmountMicros, original.AmountMicros)
		}
	})
}
