package api

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/nicklaw5/helix"
	"github.com/saxypandabear/twitchsongrequests/pkg/db"
	"github.com/saxypandabear/twitchsongrequests/pkg/queue"
)

const (
	SongRequestsTitle = "TwitchSongRequests"
	verificationType  = "webhook_callback_verification"
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
}

func NewRewardHandler(twitchSecret string, publisher queue.Publisher, userStore db.UserStore) *RewardHandler {
	return &RewardHandler{
		secret:    twitchSecret,
		publisher: publisher,
		userStore: userStore,
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

	if vals.Event != nil {
		log.Printf("Found event to consume: %s", string(vals.Event))
		var redeemEvent helix.EventSubChannelPointsCustomRewardRedemptionEvent
		if err = json.NewDecoder(bytes.NewReader(vals.Event)).Decode(&redeemEvent); err != nil {
			log.Println("failed to unmarshal payload", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if !IsValidSongRequest(&redeemEvent) {
			log.Println("not a valid song request, so dropping")
			w.WriteHeader(http.StatusOK)
			return
		}

		log.Printf("User '%s' submitted '%s'", redeemEvent.UserName, redeemEvent.UserInput)
		// TODO: get access token for Spotify and create Spotify client
		if err = h.publisher.Publish(nil, redeemEvent.UserInput); err != nil {
			log.Println("failed to publish")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	if _, err = w.Write([]byte("ok")); err != nil {
		log.Println("failed to write response body", err)
	}
}

func IsVerificationRequest(r *http.Request) bool {
	return verificationType == r.Header.Get(strings.ToLower(messageTypeHeader))
}

func IsValidSongRequest(e *helix.EventSubChannelPointsCustomRewardRedemptionEvent) bool {
	return e != nil && strings.Contains(e.Reward.Title, SongRequestsTitle)
}
