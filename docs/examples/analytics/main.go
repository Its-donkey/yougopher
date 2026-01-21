// Example: Analytics Dashboard
//
// This example demonstrates how to query YouTube Analytics data
// and display channel statistics in a simple dashboard format.
//
// Usage:
//
//	export YOUTUBE_CLIENT_ID=your-client-id
//	export YOUTUBE_CLIENT_SECRET=your-client-secret
//	go run main.go
//
// The dashboard will display:
//  1. Channel overview (total views, watch time, subscribers)
//  2. Top 10 videos by views
//  3. Daily view trend for the last 30 days
//  4. Geographic breakdown of viewers
//  5. Device type breakdown
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Its-donkey/yougopher/youtube/analytics"
	"github.com/Its-donkey/yougopher/youtube/auth"
)

func main() {
	// Load credentials
	clientID := os.Getenv("YOUTUBE_CLIENT_ID")
	clientSecret := os.Getenv("YOUTUBE_CLIENT_SECRET")

	if clientID == "" || clientSecret == "" {
		log.Fatal("Set YOUTUBE_CLIENT_ID and YOUTUBE_CLIENT_SECRET environment variables")
	}

	ctx := context.Background()

	// Create auth client with analytics scopes
	authClient := auth.NewAuthClient(auth.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  "http://localhost:8080/callback",
		Scopes: []string{
			auth.ScopeReadOnly,
			auth.ScopePartner, // Required for analytics
		},
	})

	// Auth flow
	authDone := make(chan struct{})

	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		state := fmt.Sprintf("%d", time.Now().UnixNano())
		url := authClient.AuthorizationURL(state, auth.WithPrompt("consent"))
		http.Redirect(w, r, url, http.StatusFound)
	})

	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		_, err := authClient.Exchange(r.Context(), code)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "Authentication successful! Check your terminal for the dashboard.")
		close(authDone)
	})

	server := &http.Server{Addr: ":8080"}
	go func() {
		log.Println("Visit http://localhost:8080/login to authenticate")
		_ = server.ListenAndServe()
	}()

	<-authDone
	time.Sleep(500 * time.Millisecond)
	_ = server.Shutdown(ctx)

	// Create analytics client with token provider
	client := analytics.NewClient(
		analytics.WithTokenProvider(authClient.AccessToken),
	)

	// Date range: last 30 days
	endDate := time.Now().Format("2006-01-02")
	startDate := time.Now().AddDate(0, 0, -30).Format("2006-01-02")

	fmt.Println()
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("         YOUTUBE ANALYTICS DASHBOARD")
	fmt.Printf("         %s to %s\n", startDate, endDate)
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println()

	// 1. Channel Overview
	printChannelOverview(ctx, client, startDate, endDate)

	// 2. Top Videos
	printTopVideos(ctx, client, startDate, endDate)

	// 3. Daily Trend
	printDailyTrend(ctx, client, startDate, endDate)

	// 4. Geographic Breakdown
	printCountryBreakdown(ctx, client, startDate, endDate)

	// 5. Device Breakdown
	printDeviceBreakdown(ctx, client, startDate, endDate)

	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("Dashboard complete!")
}

func printChannelOverview(ctx context.Context, client *analytics.Client, startDate, endDate string) {
	fmt.Println("CHANNEL OVERVIEW")
	fmt.Println(strings.Repeat("-", 40))

	report, err := client.QueryChannelViews(ctx, startDate, endDate)
	if err != nil {
		log.Printf("Error fetching channel views: %v", err)
		return
	}

	if len(report.Rows()) == 0 {
		fmt.Println("No data available")
		fmt.Println()
		return
	}

	row := report.Rows()[0]
	views := row.GetInt("views")
	minutes := row.GetFloat("estimatedMinutesWatched")
	avgDuration := row.GetFloat("averageViewDuration")
	subsGained := row.GetInt("subscribersGained")
	subsLost := row.GetInt("subscribersLost")

	fmt.Printf("  Total Views:        %s\n", formatNumber(views))
	fmt.Printf("  Watch Time:         %s\n", formatMinutes(minutes))
	fmt.Printf("  Avg View Duration:  %s\n", formatSeconds(avgDuration))
	fmt.Printf("  Subscribers:        +%d / -%d (net: %+d)\n",
		subsGained, subsLost, subsGained-subsLost)
	fmt.Println()
}

func printTopVideos(ctx context.Context, client *analytics.Client, startDate, endDate string) {
	fmt.Println("TOP 10 VIDEOS")
	fmt.Println(strings.Repeat("-", 40))

	report, err := client.QueryTopVideos(ctx, startDate, endDate, 10)
	if err != nil {
		log.Printf("Error fetching top videos: %v", err)
		return
	}

	rows := report.Rows()
	if len(rows) == 0 {
		fmt.Println("No data available")
		fmt.Println()
		return
	}

	for i, row := range rows {
		videoID := row.GetString("video")
		views := row.GetInt("views")
		likes := row.GetInt("likes")
		comments := row.GetInt("comments")

		fmt.Printf("  %2d. %s\n", i+1, videoID)
		fmt.Printf("      Views: %s | Likes: %d | Comments: %d\n",
			formatNumber(views), likes, comments)
	}
	fmt.Println()
}

func printDailyTrend(ctx context.Context, client *analytics.Client, startDate, endDate string) {
	fmt.Println("DAILY VIEW TREND (Last 7 Days)")
	fmt.Println(strings.Repeat("-", 40))

	// Get last 7 days for the trend display
	recentStart := time.Now().AddDate(0, 0, -7).Format("2006-01-02")
	report, err := client.QueryDailyViews(ctx, recentStart, endDate)
	if err != nil {
		log.Printf("Error fetching daily views: %v", err)
		return
	}

	rows := report.Rows()
	if len(rows) == 0 {
		fmt.Println("No data available")
		fmt.Println()
		return
	}

	// Find max views for scaling the bar chart
	var maxViews int64
	for _, row := range rows {
		views := row.GetInt("views")
		if views > maxViews {
			maxViews = views
		}
	}

	for _, row := range rows {
		date := row.GetString("day")
		views := row.GetInt("views")

		// Create a simple bar chart
		barLength := 0
		if maxViews > 0 {
			barLength = int(float64(views) / float64(maxViews) * 30)
		}
		bar := strings.Repeat("█", barLength)

		fmt.Printf("  %s | %s %s\n", date, bar, formatNumber(views))
	}
	fmt.Println()
}

func printCountryBreakdown(ctx context.Context, client *analytics.Client, startDate, endDate string) {
	fmt.Println("TOP COUNTRIES")
	fmt.Println(strings.Repeat("-", 40))

	report, err := client.QueryCountryBreakdown(ctx, startDate, endDate)
	if err != nil {
		log.Printf("Error fetching country breakdown: %v", err)
		return
	}

	rows := report.Rows()
	if len(rows) == 0 {
		fmt.Println("No data available")
		fmt.Println()
		return
	}

	// Show top 10 countries
	displayCount := 10
	if len(rows) < displayCount {
		displayCount = len(rows)
	}

	totalViews := report.TotalViews()
	for i := 0; i < displayCount; i++ {
		row := rows[i]
		country := row.GetString("country")
		views := row.GetInt("views")
		pct := float64(views) / float64(totalViews) * 100

		fmt.Printf("  %2d. %s: %s (%.1f%%)\n",
			i+1, countryName(country), formatNumber(views), pct)
	}
	fmt.Println()
}

func printDeviceBreakdown(ctx context.Context, client *analytics.Client, startDate, endDate string) {
	fmt.Println("DEVICE BREAKDOWN")
	fmt.Println(strings.Repeat("-", 40))

	report, err := client.QueryDeviceBreakdown(ctx, startDate, endDate)
	if err != nil {
		log.Printf("Error fetching device breakdown: %v", err)
		return
	}

	rows := report.Rows()
	if len(rows) == 0 {
		fmt.Println("No data available")
		fmt.Println()
		return
	}

	totalViews := report.TotalViews()
	for _, row := range rows {
		device := row.GetString("deviceType")
		views := row.GetInt("views")
		pct := float64(views) / float64(totalViews) * 100

		fmt.Printf("  %-15s: %s (%.1f%%)\n",
			deviceName(device), formatNumber(views), pct)
	}
	fmt.Println()
}

// Helper functions

func formatNumber(n int64) string {
	if n >= 1000000 {
		return fmt.Sprintf("%.1fM", float64(n)/1000000)
	}
	if n >= 1000 {
		return fmt.Sprintf("%.1fK", float64(n)/1000)
	}
	return fmt.Sprintf("%d", n)
}

func formatMinutes(minutes float64) string {
	hours := minutes / 60
	if hours >= 24 {
		days := hours / 24
		return fmt.Sprintf("%.1f days", days)
	}
	return fmt.Sprintf("%.1f hours", hours)
}

func formatSeconds(seconds float64) string {
	if seconds >= 60 {
		mins := int(seconds / 60)
		secs := int(seconds) % 60
		return fmt.Sprintf("%dm %ds", mins, secs)
	}
	return fmt.Sprintf("%.0fs", seconds)
}

func countryName(code string) string {
	countries := map[string]string{
		"US": "United States",
		"GB": "United Kingdom",
		"CA": "Canada",
		"AU": "Australia",
		"DE": "Germany",
		"FR": "France",
		"JP": "Japan",
		"BR": "Brazil",
		"IN": "India",
		"MX": "Mexico",
	}
	if name, ok := countries[code]; ok {
		return name
	}
	return code
}

func deviceName(device string) string {
	devices := map[string]string{
		"MOBILE":       "Mobile",
		"DESKTOP":      "Desktop",
		"TABLET":       "Tablet",
		"TV":           "TV",
		"GAME_CONSOLE": "Game Console",
		"UNKNOWN":      "Unknown",
	}
	if name, ok := devices[device]; ok {
		return name
	}
	return device
}
