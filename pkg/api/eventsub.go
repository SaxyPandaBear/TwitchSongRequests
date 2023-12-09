package api

import (
	"net/http"

	"github.com/nicklaw5/helix/v2"
	"github.com/saxypandabear/twitchsongrequests/internal/util"
	"github.com/saxypandabear/twitchsongrequests/pkg/db"
	"go.uber.org/zap"
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
	prefStore   db.PreferenceStore
	callbackURL string
	secret      string
}

func NewEventSubHandler(u db.UserStore, p db.PreferenceStore, auth *util.AuthConfig, callbackURL, secret string) *EventSubHandler {
	return &EventSubHandler{
		userStore:   u,
		prefStore:   p,
		auth:        auth,
		callbackURL: callbackURL,
		secret:      secret,
	}
}

// SubscribeToTopic
func (e *EventSubHandler) SubscribeToTopic(w http.ResponseWriter, r *http.Request) {
	id, err := util.GetUserIDFromRequest(r)
	if err != nil {
		zap.L().Error("failed to get Twitch ID from request", zap.Error(err))
		http.Redirect(w, r, e.callbackURL, http.StatusFound)
		return
	}

	user, err := e.userStore.GetUser(id)
	if err != nil {
		zap.L().Error("failed to get user", zap.Error(err))
		http.Redirect(w, r, e.callbackURL, http.StatusFound)
		return
	}

	pref, err := e.prefStore.GetPreference(id)
	if err != nil {
		zap.L().Error("failed to get user preferences", zap.Error(err))
		http.Redirect(w, r, e.callbackURL, http.StatusFound)
		return
	}

	// get user access token
	c, err := util.GetNewTwitchClient(e.auth)
	if err != nil {
		zap.L().Error("failed to get Twitch client", zap.String("id", id), zap.Error(err))
		http.Redirect(w, r, e.callbackURL, http.StatusFound)
		return
	}
	c.SetUserAccessToken(user.TwitchAccessToken)

	createReward := helix.ChannelCustomRewardsParams{
		BroadcasterID:       id,
		Title:               "Spotify Song Request",
		IsUserInputRequired: true,
		IsEnabled:           false, // create the reward, but don't enable it by default
		Cost:                1000,
		Prompt:              "Request with a Spotify URL, or search for a song with keywords",
	}

	rewardRes, err := c.CreateCustomReward(&createReward)
	if err != nil {
		zap.L().Error("failed to create Channel Point reward", zap.String("id", id), zap.Error(err))
		http.Redirect(w, r, e.callbackURL, http.StatusFound)
		return
	} else if len(rewardRes.ErrorMessage) > 0 || len(rewardRes.Data.ChannelCustomRewards) < 1 {
		zap.L().Error("error occurred while creating Custom Reward",
			zap.String("id", id),
			zap.Int("status", rewardRes.ErrorStatus),
			zap.String("err", rewardRes.Error),
			zap.String("error_msg", rewardRes.ErrorMessage))
		http.Redirect(w, r, e.callbackURL, http.StatusFound)
		return
	}

	createSub := helix.EventSubSubscription{
		Type:    helix.EventSubTypeChannelPointsCustomRewardRedemptionAdd,
		Version: topicVersion,
		Condition: helix.EventSubCondition{
			BroadcasterUserID: id,
			RewardID:          rewardRes.Data.ChannelCustomRewards[0].ID,
		},
		Transport: helix.EventSubTransport{
			Method:   subMethod,
			Callback: e.callbackURL + "/callback",
			Secret:   e.secret,
		},
	}

	// need to get a whole new client after setting the user access token, for some reason
	c, err = util.GetNewTwitchClient(e.auth)
	if err != nil {
		zap.L().Error("failed to get Twitch client", zap.String("id", id), zap.Error(err))
		http.Redirect(w, r, e.callbackURL, http.StatusFound)
		return
	}
	// creating the event sub subscription requires an app access token
	token, err := c.RequestAppAccessToken([]string{e.auth.Scope})
	if err != nil {
		zap.L().Error("failed to get updated access token", zap.String("id", id), zap.Error(err))
		http.Redirect(w, r, e.callbackURL, http.StatusFound)
		return
	}
	c.SetAppAccessToken(token.Data.AccessToken)

	res, err := c.CreateEventSubSubscription(&createSub)
	if err != nil {
		zap.L().Error("failed to create EventSub subscription", zap.String("id", id), zap.Error(err))
		http.Redirect(w, r, e.callbackURL, http.StatusFound)
		return
	} else if len(res.ErrorMessage) > 0 {
		zap.L().Error("error occurred while creating EventSub subscription",
			zap.String("id", id),
			zap.Int("status", res.ErrorStatus),
			zap.String("err", res.Error),
			zap.String("error_msg", res.ErrorMessage))
		http.Redirect(w, r, e.callbackURL, http.StatusFound)
		return
	}

	if len(res.Data.EventSubSubscriptions) < 1 {
		zap.L().Error("failed to subscribe to webhook event", zap.String("id", id))
		http.Redirect(w, r, e.callbackURL, http.StatusFound)
		return
	}

	// successfully subscribed
	user.Subscribed = true
	user.SubscriptionID = res.Data.EventSubSubscriptions[0].ID
	err = e.userStore.UpdateUser(user)
	if err != nil {
		zap.L().Error("failed to update user", zap.String("id", id), zap.Error(err))
		http.Redirect(w, r, e.callbackURL, http.StatusFound)
		return
	}

	pref.CustomRewardID = rewardRes.Data.ChannelCustomRewards[0].ID
	err = e.prefStore.UpdatePreference(pref)
	if err != nil {
		zap.L().Error("failed to update user preferences", zap.Error(err))
		http.Redirect(w, r, e.callbackURL, http.StatusFound)
		return
	}

	zap.L().Info("successfully subscribed to Channel Point topic", zap.String("id", id), zap.String("subscription", user.SubscriptionID))

	http.Redirect(w, r, e.callbackURL, http.StatusFound)
}
