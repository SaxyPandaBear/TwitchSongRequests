package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/nicklaw5/helix/v2"
	"github.com/saxypandabear/twitchsongrequests/internal/util"
	"github.com/saxypandabear/twitchsongrequests/pkg/db"
	"github.com/saxypandabear/twitchsongrequests/pkg/o11y/metrics"
	"github.com/saxypandabear/twitchsongrequests/pkg/preferences"
	"github.com/saxypandabear/twitchsongrequests/pkg/queue"
	"go.uber.org/zap"

	"github.com/saxypandabear/twitchsongrequests/pkg/chatbot"
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

	processedEvents sync.Map
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
		OnSuccess: UpdateRedemptionStatus,
	}
}

func (h *RewardHandler) ChannelPointRedeem(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		zap.L().Error("failed to read request body", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	// verify that the notification came from twitch using the secret.
	if !helix.VerifyEventSubNotification(h.config.Secret, r.Header, string(body)) {
		zap.L().Error("no valid signature on subscription")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var vals EventSubNotification
	if err = json.NewDecoder(bytes.NewReader(body)).Decode(&vals); err != nil {
		zap.L().Error("failed to unmarshal request body", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// when initially verifying the subscription, there won't be an event in the request
	// body. need to handle this gracefully in the API by responding directly with the challenge.
	// https://dev.twitch.tv/docs/eventsub/handling-webhook-events/#responding-to-a-challenge-request
	if IsVerificationRequest(r) {
		w.WriteHeader(http.StatusOK)
		if _, err = w.Write([]byte(vals.Challenge)); err != nil {
			zap.L().Error("failed to write challenge response for verification", zap.Error(err))
		}
		return // short-circuit here because of the request type
	}

	// A request can come in to revoke the subscription. Drop the request
	// https://dev.twitch.tv/docs/eventsub/handling-webhook-events/#revoking-your-subscription
	if IsRevocationRequest(r) {
		zap.L().Error("Revoked access",
			zap.String("subscriptionID", vals.Subscription.ID),
			zap.String("status", vals.Subscription.Status),
			zap.String("userID", vals.Subscription.Condition.BroadcasterUserID),
			zap.String("reason", vals.Subscription.Status))
		w.WriteHeader(http.StatusOK)
		return
	}

	zap.L().Debug("Received event to consume", zap.String("event", string(vals.Event)))
	var redeemEvent helix.EventSubChannelPointsCustomRewardRedemptionEvent
	if err = json.NewDecoder(bytes.NewReader(vals.Event)).Decode(&redeemEvent); err != nil {
		zap.L().Error("failed to unmarshal payload", zap.Error(err))
		w.WriteHeader(http.StatusOK)
		return
	}

	userID := redeemEvent.BroadcasterUserID
	broadcaster := redeemEvent.BroadcasterUserLogin

	// Check if the event has already been processed
	if _, exists := h.processedEvents.LoadOrStore(redeemEvent.ID, time.Now()); exists {
		zap.L().Info("Duplicate event detected, skipping",
			zap.String("id", userID),
			zap.String("broadcaster", broadcaster),
			zap.String("redeemID", redeemEvent.ID))
		w.WriteHeader(http.StatusOK)
		return
	}

	// Remove event from memory after 5 minutes
	defer func() {
		go func() {
			time.Sleep(5 * time.Minute)
			h.processedEvents.Delete(redeemEvent.ID)
		}()
	}()

	preferences, err := h.config.PrefStore.GetPreference(userID)
	if err != nil {
		zap.L().Error("failed to get user preferences", zap.String("id", userID), zap.String("broadcaster", broadcaster), zap.Error(err))
	}

	if !IsValidReward(&redeemEvent, preferences) {
		zap.L().Debug("not a valid song request, so dropping", zap.String("id", userID), zap.String("broadcaster", broadcaster))
		w.WriteHeader(http.StatusOK)
		return
	}

	tok, err := db.FetchSpotifyToken(h.config.UserStore, redeemEvent.BroadcasterUserID)
	if err != nil {
		zap.L().Error("failed to verify user for spotify access", zap.String("id", userID), zap.String("broadcaster", broadcaster), zap.Error(err))
		w.WriteHeader(http.StatusOK)
		return
	}

	refreshed, err := util.RefreshSpotifyToken(r.Context(), h.config.Spotify, tok)
	if err != nil {
		zap.L().Error("failed to get valid token", zap.String("id", userID), zap.String("broadcaster", broadcaster), zap.Error(err))
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "failed to get refreshed token")
		return
	}

	// store the refreshed token
	u, err := h.config.UserStore.GetUser(userID)
	if err == nil {
		u.SpotifyAccessToken = refreshed.AccessToken
		u.SpotifyRefreshToken = refreshed.RefreshToken
		u.SpotifyExpiry = &refreshed.Expiry

		zap.L().Debug("saving updated Spotify credentials", zap.String("id", userID), zap.String("broadcaster", broadcaster))

		if err = h.config.UserStore.UpdateUser(u); err != nil {
			// if we got a valid token but failed to update the DB this is not necessarily fatal.
			zap.L().Error("failed to update user's spotify token", zap.String("id", userID), zap.String("broadcaster", broadcaster), zap.Error(err))
		}
	}

	c := util.GetNewSpotifyClient(r.Context(), h.config.Spotify, refreshed)

	sID, err := h.config.Publisher.Publish(c, redeemEvent.UserInput, preferences)
	msg := metrics.Message{
		CreatedAt:     &redeemEvent.RedeemedAt.Time,
		BroadcasterID: redeemEvent.BroadcasterUserID,
		SpotifyTrack:  sID.String(), // TODO: not sure if this works if it fails to parse..
	}
	if err != nil {
		zap.L().Error("failed to publish",
			zap.String("input", redeemEvent.UserInput),
			zap.String("id", userID),
			zap.String("broadcaster", broadcaster),
			zap.Error(err))
	} else {
		msg.Success = 1
		zap.L().Info("Submitted song request",
			zap.String("user", redeemEvent.UserName),
			zap.String("uri", redeemEvent.UserInput),
			zap.String("name", sID.String()),
			zap.String("id", userID),
			zap.String("broadcaster", broadcaster))
	}

	h.config.MsgCount.AddMessage(&msg)

	// after publishing successfully, attempt to update the status of the
	// redemption
	if err = h.OnSuccess(h.config.Twitch, h.config.UserStore, &redeemEvent, err == nil); err != nil {
		// don't need to fail fast here because this is housekeeping
		zap.L().Error("failed to update Twitch reward redemption status", zap.String("id", userID), zap.String("broadcaster", broadcaster), zap.Error(err))
	}

	w.WriteHeader(http.StatusOK)
	if _, err = w.Write([]byte("ok")); err != nil {
		zap.L().Error("failed to write response body", zap.String("id", userID), zap.String("broadcaster", broadcaster), zap.Error(err))
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

// UpdateRedemptionStatus attempts to update the status for the channel point redemption.
// This is helpful so failed submissions should get refunded.
func UpdateRedemptionStatus(auth *util.AuthConfig,
	userStore db.UserStore,
	event *helix.EventSubChannelPointsCustomRewardRedemptionEvent,
	success bool) error {
	userID := event.BroadcasterUserID
	broadcaster := event.BroadcasterUserLogin

	client, err := util.GetNewTwitchClient(auth)
	if err != nil {
		zap.L().Error("failed to create Twitch client", zap.String("id", userID), zap.String("broadcaster", broadcaster), zap.Error(err))
		return err
	}

	u, err := userStore.GetUser(event.BroadcasterUserID)
	if err != nil {
		zap.L().Error("failed to get user", zap.String("id", userID), zap.String("broadcaster", broadcaster), zap.Error(err))
		return err
	}

	client.SetUserAccessToken(u.TwitchAccessToken)
	token, err := client.RefreshUserAccessToken(u.TwitchRefreshToken)
	if err != nil {
		zap.L().Error("failed to refresh Twitch token", zap.String("id", userID), zap.String("broadcaster", broadcaster), zap.Error(err))
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

	zap.L().Debug("updated redemptions",
		zap.String("id", userID),
		zap.String("broadcaster", broadcaster),
		zap.Int("status", resp.StatusCode),
		zap.String("error", resp.ErrorMessage),
		zap.Int("num", len(resp.Data.Redemptions)))
	if resp.StatusCode >= 400 {
		return errors.New(resp.ErrorMessage)
	}

	// update user details for Twitch auth
	u.TwitchAccessToken = token.Data.AccessToken
	u.TwitchRefreshToken = token.Data.RefreshToken
	if err = userStore.UpdateUser(u); err != nil {
		zap.L().Error("failed to update Twitch credentials", zap.String("id", userID), zap.String("broadcaster", broadcaster), zap.Error(err))
		return err
	}

	// After updating the status, we send a message to the chat
	if success {
		if err := chatbot.SendChatMessage(broadcaster, userStore, event); err != nil {
			zap.L().Error("Error sending chat message",
				zap.String("id", userID),
				zap.String("broadcaster", broadcaster),
				zap.Error(err))
			return err
		}
	}

	return nil
}
