package api

import (
	"log"
	"net/http"

	"github.com/nicklaw5/helix/v2"
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
		callbackURL: callbackURL + "/callback",
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
		w.Write([]byte(err.Error()))
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
		log.Println("failed to get Twitch client for", id, err)
		w.Write([]byte(err.Error()))
		http.Redirect(w, r, e.callbackURL, http.StatusFound)
		return
	}

	token, err := c.RequestAppAccessToken([]string{e.auth.Scope})
	if err != nil {
		log.Println("failed to get updated access token for", id, err)
		w.Write([]byte(err.Error()))
		http.Redirect(w, r, e.callbackURL, http.StatusFound)
		return
	}
	c.SetAppAccessToken(token.Data.AccessToken)

	res, err := c.CreateEventSubSubscription(&createSub)
	if err != nil {
		log.Println("failed to create EventSub subscription ", err)
		w.Write([]byte(err.Error()))
		http.Redirect(w, r, e.callbackURL, http.StatusFound)
		return
	} else if len(res.ErrorMessage) > 0 {
		log.Printf("error occurred while creating EventSub subscription | HTTP %v | %s | %s\n", res.ErrorStatus, res.Error, res.ErrorMessage)
		w.Write([]byte(res.ErrorMessage))
		http.Redirect(w, r, e.callbackURL, http.StatusFound)
		return
	}

	log.Println("Subscriptions:", res.Data.EventSubSubscriptions)

	// successfully subscribed
	user.Subscribed = true
	err = e.userStore.UpdateUser(user)

	if err != nil {
		log.Println("failed to update twitch credentials", err)
		w.Write([]byte(err.Error()))
		http.Redirect(w, r, e.callbackURL, http.StatusFound)
		return
	}

	log.Println("successfully subscribed to Channel Point topic for user", id)

	http.Redirect(w, r, e.callbackURL, http.StatusFound)
}
