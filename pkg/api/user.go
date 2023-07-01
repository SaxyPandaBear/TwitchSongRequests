package api

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/nicklaw5/helix/v2"
	"github.com/saxypandabear/twitchsongrequests/internal/constants"
	"github.com/saxypandabear/twitchsongrequests/internal/util"
	"github.com/saxypandabear/twitchsongrequests/pkg/db"
	"go.uber.org/zap"
)

type UserHandler struct {
	users       db.UserStore
	prefs       db.PreferenceStore
	redirectURL string
	twitch      *util.AuthConfig
	spotify     *util.AuthConfig
}

func NewUserHandler(d db.UserStore, p db.PreferenceStore, redirectURL string, twitch, spotify *util.AuthConfig) *UserHandler {
	return &UserHandler{
		users:       d,
		prefs:       p,
		redirectURL: redirectURL,
		twitch:      twitch,
		spotify:     spotify,
	}
}

func (h *UserHandler) RevokeUserAccesses(w http.ResponseWriter, r *http.Request) {
	userID, err := util.GetUserIDFromRequest(r)

	if err != nil {
		zap.L().Error("failed to get Twitch ID from request", zap.Error(err))
		http.Redirect(w, r, h.redirectURL, http.StatusFound)
		return
	}

	c, err := util.GetNewTwitchClient(h.twitch)
	if err != nil {
		zap.L().Error("failed to get Twitch client", zap.Error(err))
		http.Redirect(w, r, h.redirectURL, http.StatusFound)
		return
	}

	u, err := h.users.GetUser(userID)
	if err != nil {
		zap.L().Error("failed to get user", zap.Error(err))
		http.Redirect(w, r, h.redirectURL, http.StatusFound)
		return
	}

	// revoke subscription first
	appToken, err := c.RequestAppAccessToken(strings.Split(h.twitch.Scope, " "))
	if err != nil {
		zap.L().Error("failed to get app access token", zap.Error(err))
		http.Redirect(w, r, h.redirectURL, http.StatusFound)
		return
	}
	c.SetAppAccessToken(appToken.Data.AccessToken)
	res, err := c.RemoveEventSubSubscription(u.SubscriptionID)
	if err != nil {
		zap.L().Error("failed to remove eventsub subscription", zap.String("id", userID), zap.Error(err))
		http.Redirect(w, r, h.redirectURL, http.StatusFound)
		return
	} else if len(res.ErrorMessage) > 0 {
		zap.L().Error("failed to remove eventsub subscription", zap.String("id", userID), zap.Int("status", res.ErrorStatus), zap.String("err", res.Error), zap.String("error_msg", res.ErrorMessage))
		http.Redirect(w, r, h.redirectURL, http.StatusFound)
		return
	}

	tok, err := db.FetchTwitchToken(h.users, userID)
	if err != nil {
		zap.L().Error("failed to get user token", zap.Error(err))
		http.Redirect(w, r, h.redirectURL, http.StatusFound)
		return
	}

	c.SetUserAccessToken(tok.AccessToken)
	tokenResp, err := c.RefreshUserAccessToken(tok.RefreshToken)
	if err != nil {
		zap.L().Error("failed to refresh twitch token", zap.Error(err))
		http.Redirect(w, r, h.redirectURL, http.StatusFound)
		return
	}
	c.SetUserAccessToken(tokenResp.Data.AccessToken)

	// attempt to remove the reward, if the reward ID is non-empty
	prefs, err := h.prefs.GetPreference(userID)
	if err != nil {
		// This is really a non-issue. Log and move on.
		zap.L().Warn("failed to get user preferences", zap.String("id", userID), zap.Error(err))
	} else if len(prefs.CustomRewardID) > 0 {
		dcrResponse, err := c.DeleteCustomRewards(&helix.DeleteCustomRewardsParams{
			BroadcasterID: userID,
			ID:            prefs.CustomRewardID,
		})
		// again, if this fails it's really a non-issue. log and move on.
		if err != nil {
			zap.L().Warn("failed to call api to delete custom reward", zap.String("id", userID), zap.String("reward_id", prefs.CustomRewardID), zap.Error(err))
		} else if dcrResponse.StatusCode >= 400 {
			zap.L().Warn("failed to delete custom reward", zap.String("id", userID), zap.String("reward_id", prefs.CustomRewardID), zap.Int("status", dcrResponse.StatusCode), zap.String("error", dcrResponse.Error), zap.String("error_msg", dcrResponse.ErrorMessage))
		}
	}

	// make sure that the token is fresh before revoking
	_, err = c.RevokeUserAccessToken(tokenResp.Data.AccessToken)
	if err != nil {
		zap.L().Error("failed to revoke access", zap.Error(err))
		http.Redirect(w, r, h.redirectURL, http.StatusFound)
		return
	}

	err = h.users.DeleteUser(userID)
	if err != nil {
		zap.L().Error("failed to delete user", zap.Error(err))
		http.Redirect(w, r, h.redirectURL, http.StatusFound)
		return
	}

	log.Println("successfully deleted user", userID)

	err = h.prefs.DeletePreference(userID)
	if err != nil {
		// I'm not sure if this is fatal or not.
		zap.L().Error("failed to delete user", zap.Error(err))
	}

	twitchCookie := http.Cookie{
		Name:     constants.TwitchIDCookieKey,
		Path:     "/",
		Value:    "",
		MaxAge:   -1, // delete the cookie
		Expires:  time.Now().Add(-1 * time.Hour),
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(w, &twitchCookie)
	http.Redirect(w, r, h.redirectURL, http.StatusFound)
}
