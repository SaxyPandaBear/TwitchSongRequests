package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/nicklaw5/helix"
	"github.com/saxypandabear/twitchsongrequests/internal/util"
	"github.com/saxypandabear/twitchsongrequests/pkg/db"
	"github.com/saxypandabear/twitchsongrequests/pkg/queue"
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
	secret    string
	publisher queue.Publisher
	userStore db.UserStore
	spotify   *util.AuthConfig
}

func NewRewardHandler(twitchSecret string, publisher queue.Publisher, userStore db.UserStore, auth *util.AuthConfig) *RewardHandler {
	return &RewardHandler{
		secret:    twitchSecret,
		publisher: publisher,
		userStore: userStore,
		spotify:   auth,
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
	if !helix.VerifyEventSubNotification(h.secret, r.Header, string(body)) {
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

	if !IsValidReward(&redeemEvent) {
		log.Println("not a valid song request, so dropping")
		w.WriteHeader(http.StatusOK)
		return
	}

	tok, err := db.FetchSpotifyToken(h.userStore, redeemEvent.BroadcasterUserID)
	if err != nil {
		log.Printf("failed to verify user %s for spotify access: %v\n", redeemEvent.BroadcasterUserID, err)
		w.WriteHeader(http.StatusOK)
		return
	}

	refreshed, err := util.RefreshSpotifyToken(r.Context(), h.spotify, tok)
	if err != nil {
		log.Println("failed to get valid token", err)
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "failed to get refreshed token")
		return
	}

	// store the refreshed token
	u, err := h.userStore.GetUser(redeemEvent.BroadcasterUserID)
	if err == nil {
		u.SpotifyAccessToken = refreshed.AccessToken
		u.SpotifyRefreshToken = refreshed.RefreshToken
		u.SpotifyExpiry = &refreshed.Expiry

		log.Println("saving updated Spotify credentials for", u.TwitchID)

		if err = h.userStore.UpdateUser(u); err != nil {
			// if we got a valid token but failed to update the DB this is not necessarily fatal.
			log.Println("failed to update user's spotify token", err)
		}
	}

	c := util.GetNewSpotifyClient(r.Context(), h.spotify, refreshed)

	log.Printf("User '%s' submitted '%s'", redeemEvent.UserName, redeemEvent.UserInput)

	if err = h.publisher.Publish(c, redeemEvent.UserInput); err != nil {
		log.Println("failed to publish:", err)
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "failed to publish")
		return
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

// IsValidReward ensures that the redemption event has a title which contains the
// named keyword. This is a naive approach for dropping unwanted events.
func IsValidReward(e *helix.EventSubChannelPointsCustomRewardRedemptionEvent) bool {
	return e != nil && strings.Contains(e.Reward.Title, SongRequestsTitle)
}
