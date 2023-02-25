package api

import (
	"encoding/base64"
	"log"
	"net/http"

	"github.com/nicklaw5/helix"
	"github.com/saxypandabear/twitchsongrequests/internal/locking"
	"github.com/saxypandabear/twitchsongrequests/pkg/constants"
	"github.com/saxypandabear/twitchsongrequests/pkg/db"
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
	userStore   db.UserStore
	callbackURL string
	secret      string
}

func NewEventSubHandler(u db.UserStore, c *helix.Client, callbackURL, secret string) *EventSubHandler {
	return &EventSubHandler{
		userStore:   u,
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
		http.Redirect(w, r, e.callbackURL, http.StatusFound)
		return
	}

	if err = c.Valid(); err != nil {
		log.Println("cookie expired", err)
		http.Redirect(w, r, e.callbackURL, http.StatusFound)
		return
	}

	idBytes, err := base64.StdEncoding.DecodeString(c.Value)
	if err != nil {
		log.Println("failed to decode cookie", err)
		http.Redirect(w, r, e.callbackURL, http.StatusFound)
		return
	}

	id := string(idBytes)

	user, err := e.userStore.GetUser(id)
	if err != nil {
		log.Println("failed to get user", err)
		http.Redirect(w, r, e.callbackURL, http.StatusFound)
		return
	}

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

	// get user access token
	locking.TwitchClientLock.Lock()
	defer locking.TwitchClientLock.Unlock()
	e.client.SetUserAccessToken(user.TwitchAccessToken)
	token, err := e.client.RefreshUserAccessToken(user.TwitchRefreshToken)

	if err != nil {
		log.Println("failed to get user access token", err)
		http.Redirect(w, r, e.callbackURL, http.StatusFound)
		return
	}

	e.client.SetUserAccessToken(token.Data.AccessToken)

	_, err = e.client.CreateEventSubSubscription(&createSub)
	if err != nil {
		log.Println("failed to create EventSub subscription ", err)
		http.Redirect(w, r, e.callbackURL, http.StatusFound)
		return
	}

	// refresh db with the updated token
	user.TwitchAccessToken = token.Data.AccessToken
	user.TwitchRefreshToken = token.Data.RefreshToken
	err = e.userStore.UpdateUser(user)

	if err != nil {
		log.Println("failed to update twitch credentials", err)
		http.Redirect(w, r, e.callbackURL, http.StatusFound)
		return
	}

	log.Println("successfully subscribed to Channel Point topic for user", id)

	http.Redirect(w, r, e.callbackURL, http.StatusFound)
}
