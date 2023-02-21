package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/nicklaw5/helix"
	"github.com/saxypandabear/twitchsongrequests/pkg/db"
	"github.com/saxypandabear/twitchsongrequests/pkg/queue"
	"github.com/zmb3/spotify/v2"
	"golang.org/x/oauth2"
)

const (
	SongRequestsTitle = "TwitchSongRequests"
	verificationType  = "webhook_callback_verification"
	messageTypeHeader = "Twitch-Eventsub-Message-Type"
)

var spotifyRefreshMutex sync.Mutex

type EventSubNotification struct {
	Subscription helix.EventSubSubscription `json:"subscription"`
	Challenge    string                     `json:"challenge"`
	Event        json.RawMessage            `json:"event"`
}

type RewardHandler struct {
	secret    string
	publisher queue.Publisher
	userStore db.UserStore
	auth      *oauth2.Config
	refresher func(*oauth2.Token) (*oauth2.Token, error)
}

func defaultOAuthTokenRefresh(t *oauth2.Token) (*oauth2.Token, error) {
	source := oauth2.ReuseTokenSource(t, nil)
	return source.Token()
}

func NewRewardHandler(twitchSecret string, publisher queue.Publisher, userStore db.UserStore, auth *oauth2.Config, refresh func(*oauth2.Token) (*oauth2.Token, error)) *RewardHandler {
	if refresh == nil {
		refresh = defaultOAuthTokenRefresh
	}

	return &RewardHandler{
		secret:    twitchSecret,
		publisher: publisher,
		userStore: userStore,
		auth:      auth,
		refresher: refresh,
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
			w.WriteHeader(http.StatusOK)
			return
		}

		if !IsValidReward(&redeemEvent) {
			log.Println("not a valid song request, so dropping")
			w.WriteHeader(http.StatusOK)
			return
		}

		// c, err := GetSpotifyClient(h.userStore, h.auth, redeemEvent.BroadcasterUserID)
		tok, err := GetOAuthToken(h.userStore, redeemEvent.BroadcasterUserID)
		if err != nil {
			log.Printf("failed to verify user %s for spotify access: %v\n", redeemEvent.BroadcasterUserID, err)
			w.WriteHeader(http.StatusOK)
			return
		}

		source := oauth2.ReuseTokenSource(tok, nil)
		refreshed, err := source.Token()
		if err != nil {
			log.Println("failed to get valid token", err)
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, "failed to get refreshed token")
			return
		}

		// store the refreshed token
		spotifyRefreshMutex.Lock()
		defer spotifyRefreshMutex.Unlock()
		u, err := h.userStore.GetUser(redeemEvent.BroadcasterUserID)
		if err == nil {
			u.SpotifyAccessToken = refreshed.AccessToken
			u.SpotifyRefreshToken = refreshed.RefreshToken
			u.SpotifyExpiry = &refreshed.Expiry

			if err = h.userStore.UpdateUser(u); err != nil {
				// if we got a valid token but failed to update the DB this is not necessarily fatal.
				log.Println("failed to update user's spotify token", err)
			}
		}

		c := spotify.New(h.auth.Client(r.Context(), refreshed))

		log.Printf("User '%s' submitted '%s'", redeemEvent.UserName, redeemEvent.UserInput)

		if err = h.publisher.Publish(c, redeemEvent.UserInput); err != nil {
			log.Println("failed to publish:", err)
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, "failed to publish")
		}
	}

	w.WriteHeader(http.StatusOK)
	if _, err = w.Write([]byte("ok")); err != nil {
		log.Println("failed to write response body", err)
	}
}

func GetOAuthToken(userStore db.UserStore, id string) (*oauth2.Token, error) {
	u, err := userStore.GetUser(id)
	if err != nil {
		return nil, err
	}

	tok := oauth2.Token{
		AccessToken:  u.SpotifyAccessToken,
		RefreshToken: u.SpotifyRefreshToken,
		Expiry:       *u.SpotifyExpiry,
	}

	return &tok, nil
}

func IsVerificationRequest(r *http.Request) bool {
	return verificationType == r.Header.Get(strings.ToLower(messageTypeHeader))
}

// IsValidReward ensures that the redemption event has a title which contains the
// named keyword. This is a naive approach for dropping unwanted events.
func IsValidReward(e *helix.EventSubChannelPointsCustomRewardRedemptionEvent) bool {
	return e != nil && strings.Contains(e.Reward.Title, SongRequestsTitle)
}
