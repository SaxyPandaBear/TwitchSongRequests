package chatbot

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gempir/go-twitch-irc/v2"
	"github.com/nicklaw5/helix/v2"
	"github.com/saxypandabear/twitchsongrequests/internal/constants"
	"github.com/saxypandabear/twitchsongrequests/pkg/db"
	"github.com/zmb3/spotify/v2"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

func SendChatMessage(broadcaster string, userStore db.UserStore, event *helix.EventSubChannelPointsCustomRewardRedemptionEvent) error {
	u, err := userStore.GetUser(event.BroadcasterUserID)
	if err != nil {
		return err
	}

	chatClient := twitch.NewClient(broadcaster, "oauth:"+u.TwitchAccessToken)
	chatClient.Join(broadcaster)

	chatClient.OnConnect(func() {
		spotifyToken := &oauth2.Token{
			AccessToken:  u.SpotifyAccessToken,
			RefreshToken: u.SpotifyRefreshToken,
			Expiry:       *u.SpotifyExpiry,
		}

		spotifyConfig := &oauth2.Config{
			ClientID:     os.Getenv(constants.SpotifyClientIDKey),
			ClientSecret: os.Getenv(constants.SpotifyClientSecretKey),
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://accounts.spotify.com/authorize",
				TokenURL: "https://accounts.spotify.com/api/token",
			},
		}

		httpClient := spotifyConfig.Client(context.Background(), spotifyToken)
		spotifyClient := spotify.New(httpClient)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		track, trackErr := getSpotifyTrack(ctx, spotifyClient, event.UserInput)
		if trackErr != nil {
			zap.L().Error("Error in Spotify search or retrieving the track", zap.Error(trackErr))
			chatClient.Say(broadcaster, "There was an error searching for the song.")
			return
		}

		if track != nil {
			sendTrackMessage(chatClient, broadcaster, track, event.UserName)
		} else {
			chatClient.Say(broadcaster, "No matching song was found.")
		}
	})

	err = chatClient.Connect()
	if err != nil {
		zap.L().Error("Error connecting to Twitch chat",
			zap.String("id", event.BroadcasterUserID),
			zap.String("broadcaster", broadcaster),
			zap.Error(err))
		return err
	}

	return nil
}

func getSpotifyTrack(ctx context.Context, client *spotify.Client, input string) (*spotify.FullTrack, error) {
	if isSpotifyLink(input) {
		trackID := extractTrackID(input)
		if trackID == "" {
			return nil, fmt.Errorf("ungültige Spotify-URL")
		}
		return client.GetTrack(ctx, spotify.ID(trackID))
	}

	results, err := client.Search(ctx, input, spotify.SearchTypeTrack)
	if err != nil {
		return nil, err
	}
	if len(results.Tracks.Tracks) > 0 {
		return &results.Tracks.Tracks[0], nil
	}
	return nil, nil
}

func isSpotifyLink(input string) bool {
	return strings.Contains(input, "spotify.com/track/") || strings.Contains(input, "open.spotify.com/")
}

func extractTrackID(input string) string {
	parts := strings.Split(input, "/")
	for i, part := range parts {
		if part == "track" && i+1 < len(parts) {
			return strings.Split(parts[i+1], "?")[0]
		}
	}
	return ""
}

func sendTrackMessage(client *twitch.Client, broadcaster string, track *spotify.FullTrack, userName string) {
	secondsValue := track.Duration / 1000
	minutesValue := secondsValue / 60
	secondsValue = secondsValue % 60
	message := fmt.Sprintf("@%s 🔹 Enqueued '%s' by %s (%02d:%02d) SeemsGood",
		userName, track.Name, track.Artists[0].Name, minutesValue, secondsValue)
	client.Say(broadcaster, message)
}
