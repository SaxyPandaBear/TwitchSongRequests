package api

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/http"

	"github.com/nicklaw5/helix"
	"github.com/saxypandabear/twitchsongrequests/pkg/constants"
)

const (
	topicVersion = "1"
	subMethod    = "webhook"
)

type SubscribeRequest struct {
	UserID string `json:"user_id"`
}

type EventSubHandler struct {
	client      *helix.Client
	callbackURL string
	secret      string
}

func NewEventSubHandler(c *helix.Client, callbackURL, secret string) *EventSubHandler {
	return &EventSubHandler{
		client:      c,
		callbackURL: callbackURL,
		secret:      secret,
	}
}

// SubscribeToTopic
func (e *EventSubHandler) SubscribeToTopic(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie(constants.TwitchIDCookieKey)
	if err != nil {
		log.Println("could not extract cookie", err)
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintln(w, "cookie not found")
		return
	}

	if err = c.Valid(); err != nil {
		log.Println("cookie expired", err)
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintln(w, "cookie expired")
		return
	}

	idBytes, err := base64.StdEncoding.DecodeString(c.Value)
	if err != nil {
		log.Println("failed to decode cookie", err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "failed to decode cookie")
	}

	id := string(idBytes)

	createSub := helix.EventSubSubscription{
		Type:    helix.EventSubTypeChannelPointsCustomRewardRedemptionAdd,
		Version: topicVersion,
		Condition: helix.EventSubCondition{
			BroadcasterUserID: id,
		},
		Transport: helix.EventSubTransport{
			Method:   subMethod,
			Callback: e.callbackURL,
			Secret:   e.secret,
		},
	}
	_, err = e.client.CreateEventSubSubscription(&createSub)
	if err != nil {
		log.Println("failed to create EventSub subscription ", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Println("successfully subscribed to Channel Point topic for user", id)

	http.Redirect(w, r, e.callbackURL, http.StatusFound)
}
