---
layout: default
title: Examples
description: Live streaming example applications
---

Example applications demonstrating live streaming capabilities.

## Examples

| Example | Description |
|---------|-------------|
| [Chat Bot](chatbot) | Basic live chat bot that responds to commands |
| [Moderation Bot](modbot) | Auto-moderation bot with spam detection and filtering |

## Setup

1. Clone the repository:
   ```bash
   git clone https://github.com/Its-donkey/yougopher.git
   cd yougopher
   ```

2. Create OAuth credentials in the [Google Cloud Console](https://console.cloud.google.com/apis/credentials):
   - Create a new project (or select existing)
   - Enable the YouTube Data API v3
   - Create OAuth 2.0 credentials (Desktop or Web application)
   - Add `http://localhost:8080/callback` as an authorized redirect URI

3. Set environment variables:
   ```bash
   export YOUTUBE_CLIENT_ID=your-client-id
   export YOUTUBE_CLIENT_SECRET=your-client-secret
   ```

## Running Examples

1. Navigate to the example directory:
   ```bash
   cd examples/chatbot  # or examples/modbot
   ```
2. Run the example:
   ```bash
   go run main.go
   ```
3. Open `http://localhost:8080/login` in your browser
4. Complete the OAuth flow
5. The example will connect to your active broadcast

## Requirements

| Example | Requirements |
|---------|--------------|
| chatbot | Active YouTube live stream |
| modbot | Active live stream + moderator permissions |
