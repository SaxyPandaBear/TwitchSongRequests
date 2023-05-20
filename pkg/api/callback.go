package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/nicklaw5/helix/v2"
	"github.com/saxypandabear/twitchsongrequests/internal/util"
	"github.com/saxypandabear/twitchsongrequests/pkg/db"
	"github.com/saxypandabear/twitchsongrequests/pkg/o11y/metrics"
	"github.com/saxypandabear/twitchsongrequests/pkg/preferences"
	"github.com/saxypandabear/twitchsongrequests/pkg/queue"
	"github.com/zmb3/spotify/v2"
)

const (
	SongRequestsTitle = "TwitchSongRequests"
	verificationType  = "webhook_callback_verification"
	revocationType    = "revocation"
	messageTypeHeader = "Twitch-Eventsub-Message-Type"
)

type EventSubNotification struct {
	Subscription helix.EventSubSubscription `json:"subscription"`
	Challenge    string                     `json:"challenge"`
	Event        json.RawMessage            `json:"event"`
}

type RewardHandler struct {
	config *RewardHandlerConfig

	// OnSuccess is a callback function that executes after successfully
	// publishing to the queue
	OnSuccess func(*util.AuthConfig, db.UserStore, *helix.EventSubChannelPointsCustomRewardRedemptionEvent, bool) error
	CheckUser func(context.Context, *spotify.Client, *helix.EventSubChannelPointsCustomRewardRedemptionEvent)
}

type RewardHandlerConfig struct {
	Secret    string
	Publisher queue.Publisher
	UserStore db.UserStore
	PrefStore db.PreferenceStore
	MsgCount  db.MessageCounter
	Twitch    *util.AuthConfig
	Spotify   *util.AuthConfig
}

func NewRewardHandler(config *RewardHandlerConfig) *RewardHandler {
	return &RewardHandler{
		config:    config,
		OnSuccess: DoNothingOnSuccess,
		CheckUser: CheckSpotifyUser,
	}
}

func (h *RewardHandler) ChannelPointRedeem(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	// verify that the notification came from twitch using the secret.
	if !helix.VerifyEventSubNotification(h.config.Secret, r.Header, string(body)) {
		log.Println("no valid signature on subscription")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	log.Println("verified signature for subscription")

	var vals EventSubNotification
	if err = json.NewDecoder(bytes.NewReader(body)).Decode(&vals); err != nil {
		log.Println("failed to unmarshal request body", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// when initially verifying the subscription, there won't be an event in the request
	// body. need to handle this gracefully in the API by responding directly with the challenge.
	// https://dev.twitch.tv/docs/eventsub/handling-webhook-events/#responding-to-a-challenge-request
	if IsVerificationRequest(r) {
		w.WriteHeader(http.StatusOK)
		if _, err = w.Write([]byte(vals.Challenge)); err != nil {
			log.Println("failed to write challenge response for verification", err)
		}
		return // short-circuit here because of the request type
	}

	// A request can come in to revoke the subscription. Drop the request
	// https://dev.twitch.tv/docs/eventsub/handling-webhook-events/#revoking-your-subscription
	if IsRevocationRequest(r) {
		log.Printf("Revoked access to %s: %s\n", vals.Subscription.ID, vals.Subscription.Status)
		w.WriteHeader(http.StatusOK)
		return
	}

	log.Printf("Found event to consume: %s", string(vals.Event))
	var redeemEvent helix.EventSubChannelPointsCustomRewardRedemptionEvent
	if err = json.NewDecoder(bytes.NewReader(vals.Event)).Decode(&redeemEvent); err != nil {
		log.Println("failed to unmarshal payload", err)
		w.WriteHeader(http.StatusOK)
		return
	}

	preferences, err := h.config.PrefStore.GetPreference(redeemEvent.BroadcasterUserID)
	if err != nil {
		log.Println("failed to get user preferences", err)
	}

	if !IsValidReward(&redeemEvent, preferences) {
		log.Println("not a valid song request, so dropping")
		w.WriteHeader(http.StatusOK)
		return
	}

	tok, err := db.FetchSpotifyToken(h.config.UserStore, redeemEvent.BroadcasterUserID)
	if err != nil {
		log.Printf("failed to verify user %s for spotify access: %v\n", redeemEvent.BroadcasterUserID, err)
		w.WriteHeader(http.StatusOK)
		return
	}

	refreshed, err := util.RefreshSpotifyToken(r.Context(), h.config.Spotify, tok)
	if err != nil {
		log.Println("failed to get valid token", err)
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "failed to get refreshed token")
		return
	}

	// store the refreshed token
	u, err := h.config.UserStore.GetUser(redeemEvent.BroadcasterUserID)
	if err == nil {
		u.SpotifyAccessToken = refreshed.AccessToken
		u.SpotifyRefreshToken = refreshed.RefreshToken
		u.SpotifyExpiry = &refreshed.Expiry

		log.Println("saving updated Spotify credentials for", u.TwitchID)

		if err = h.config.UserStore.UpdateUser(u); err != nil {
			// if we got a valid token but failed to update the DB this is not necessarily fatal.
			log.Println("failed to update user's spotify token", err)
		}
	}

	c := util.GetNewSpotifyClient(r.Context(), h.config.Spotify, refreshed)
	h.CheckUser(r.Context(), c, &redeemEvent)

	log.Printf("User '%s' submitted '%s'", redeemEvent.UserName, redeemEvent.UserInput)

	err = h.config.Publisher.Publish(c, redeemEvent.UserInput, preferences)
	msg := metrics.Message{
		CreatedAt: &redeemEvent.RedeemedAt.Time,
	}
	if err != nil {
		log.Println("failed to publish:", err)
	} else {
		log.Println("successfully published")
		msg.Success = 1
	}

	h.config.MsgCount.AddMessage(&msg)

	// after publishing successfully, attempt to update the status of the
	// redemption
	if err = h.OnSuccess(h.config.Twitch, h.config.UserStore, &redeemEvent, err == nil); err != nil {
		// don't need to fail fast here because this is housekeeping
		log.Println("failed to update Twitch reward redemption status", err)
	}

	w.WriteHeader(http.StatusOK)
	if _, err = w.Write([]byte("ok")); err != nil {
		log.Println("failed to write response body", err)
	}
}

func IsVerificationRequest(r *http.Request) bool {
	return verificationType == r.Header.Get(strings.ToLower(messageTypeHeader))
}

func IsRevocationRequest(r *http.Request) bool {
	return revocationType == r.Header.Get(strings.ToLower(messageTypeHeader))
}

// IsValidReward ensures that the redemption event is a valid event that we want to consume.
// Original implementation relies on the existence of a key word in the reward title. New implementation
// verifies with the stored CustomRewardID for a user's preference.
func IsValidReward(e *helix.EventSubChannelPointsCustomRewardRedemptionEvent, p *preferences.Preference) bool {
	if p != nil && p.CustomRewardID != "" {
		return e.Reward.ID == p.CustomRewardID
	}
	return e != nil && strings.Contains(e.Reward.Title, SongRequestsTitle)
}

// DoNothingOnSuccess is a no-op to satisfy the function interface. See https://github.com/SaxyPandaBear/TwitchSongRequests/issues/133
func DoNothingOnSuccess(auth *util.AuthConfig,
	userStore db.UserStore,
	event *helix.EventSubChannelPointsCustomRewardRedemptionEvent,
	success bool) error {
	return nil
}

// TODO: this doesn't work unless the application manages the channel point reward
//
//	see https://github.com/SaxyPandaBear/TwitchSongRequests/issues/133
func UpdateRedemptionStatus(auth *util.AuthConfig,
	userStore db.UserStore,
	event *helix.EventSubChannelPointsCustomRewardRedemptionEvent,
	success bool) error {
	client, err := util.GetNewTwitchClient(auth)
	if err != nil {
		log.Println("failed to create Twitch client", err)
		return err
	}

	u, err := userStore.GetUser(event.BroadcasterUserID)
	if err != nil {
		log.Println("failed to get user", err)
		return err
	}

	client.SetUserAccessToken(u.TwitchAccessToken)
	token, err := client.RefreshUserAccessToken(u.TwitchRefreshToken)
	if err != nil {
		log.Println("failed to refresh Twitch token", err)
		return err
	}
	client.SetUserAccessToken(token.Data.AccessToken)

	req := helix.UpdateChannelCustomRewardsRedemptionStatusParams{
		ID:            event.ID,
		BroadcasterID: event.BroadcasterUserID,
		RewardID:      event.Reward.ID,
	}

	if success {
		req.Status = "FULFILLED"
	} else {
		req.Status = "CANCELED"
	}
	resp, err := client.UpdateChannelCustomRewardsRedemptionStatus(&req)
	if err != nil {
		return err
	}

	log.Printf("responded with HTTP: %d | '%s' for %d redemptions\n", resp.StatusCode, resp.ErrorMessage, len(resp.Data.Redemptions))
	if resp.StatusCode >= 400 {
		return errors.New(resp.ErrorMessage)
	}

	for _, redemption := range resp.Data.Redemptions {
		log.Printf("successfully updated redemption status for %s to '%s'\n", redemption.ID, req.Status)
	}

	// update user details for Twitch auth
	u.TwitchAccessToken = token.Data.AccessToken
	u.TwitchRefreshToken = token.Data.RefreshToken
	if err = userStore.UpdateUser(u); err != nil {
		log.Println("failed to update Twitch credentials", err)
		return err
	}

	return nil
}

func CheckSpotifyUser(ctx context.Context, client *spotify.Client, redeemEvent *helix.EventSubChannelPointsCustomRewardRedemptionEvent) {
	spotifyUser, err := client.CurrentUser(ctx)
	if err != nil {
		log.Println("failed to get current Spotify user", err)
	} else {
		log.Printf("Spotify user tied to (%s|%s) = %s\n", redeemEvent.BroadcasterUserID, redeemEvent.BroadcasterUserLogin, spotifyUser.DisplayName)
	}
}
