package api

import (
	"log"
	"net/http"

	"github.com/nicklaw5/helix"
	"github.com/saxypandabear/twitchsongrequests/internal/util"
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
	auth        *util.AuthConfig
	userStore   db.UserStore
	callbackURL string
	secret      string
}

func NewEventSubHandler(u db.UserStore, auth *util.AuthConfig, callbackURL, secret string) *EventSubHandler {
	return &EventSubHandler{
		userStore:   u,
		auth:        auth,
		callbackURL: callbackURL,
		secret:      secret,
	}
}

// SubscribeToTopic
func (e *EventSubHandler) SubscribeToTopic(w http.ResponseWriter, r *http.Request) {
	id, err := util.GetUserIDFromRequest(r)
	if err != nil {
		log.Println("failed to get Twitch ID from request", err)
		http.Redirect(w, r, e.callbackURL, http.StatusFound)
		return
	}

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
	c, err := util.GetNewTwitchClient(e.auth)
	if err != nil {
		log.Println("failed to get Twitch client for", id)
		http.Redirect(w, r, e.callbackURL, http.StatusFound)
		return
	}
	c.SetUserAccessToken(user.TwitchAccessToken)
	token, err := c.RefreshUserAccessToken(user.TwitchRefreshToken)
	if err != nil {
		log.Println("failed to get updated access token for", id)
		http.Redirect(w, r, e.callbackURL, http.StatusFound)
		return
	}
	c.SetUserAccessToken(token.Data.AccessToken)

	_, err = c.CreateEventSubSubscription(&createSub)
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
