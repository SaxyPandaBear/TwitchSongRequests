package api

import (
	"encoding/base64"
	"log"
	"net/http"

	"github.com/nicklaw5/helix/v2"
	"github.com/saxypandabear/twitchsongrequests/internal/constants"
	"github.com/saxypandabear/twitchsongrequests/internal/util"
	"github.com/saxypandabear/twitchsongrequests/pkg/db"
	"github.com/saxypandabear/twitchsongrequests/pkg/preferences"
	"github.com/saxypandabear/twitchsongrequests/pkg/users"
	"go.uber.org/zap"
)

type TwitchAuthZHandler struct {
	redirectURL string
	auth        *util.AuthConfig
	userStore   db.UserStore
	prefStore   db.PreferenceStore
}

func NewTwitchAuthZHandler(url string, auth *util.AuthConfig, userStore db.UserStore, prefStore db.PreferenceStore) *TwitchAuthZHandler {
	return &TwitchAuthZHandler{
		redirectURL: url,
		auth:        auth,
		userStore:   userStore,
		prefStore:   prefStore,
	}
}

// Authorize handles the callback from the OAuth authorization
func (h *TwitchAuthZHandler) Authorize(w http.ResponseWriter, r *http.Request) {
	// https://dev.twitch.tv/docs/authentication/getting-tokens-oauth/
	if r.URL.Query().Has("error") {
		zap.L().Error("failed to authorize", zap.String("error", r.URL.Query().Get("error_description")))
		http.Redirect(w, r, h.redirectURL, http.StatusFound)
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		zap.L().Error("could not extract access code from redirect")
		http.Redirect(w, r, h.redirectURL, http.StatusFound)
		return
	}

	// validate state key matches
	state := r.URL.Query().Get("state")
	if state != h.auth.State {
		zap.L().Error("could not verify request state")
		http.Redirect(w, r, h.redirectURL, http.StatusFound)
		return
	}

	// https://dev.twitch.tv/docs/authentication/getting-tokens-oauth/#use-the-authorization-code-to-get-a-token
	client, err := util.GetNewTwitchClient(h.auth)
	if err != nil {
		zap.L().Error("failed to get Twitch client", zap.Error(err))
		http.Redirect(w, r, h.redirectURL, http.StatusFound)
		return
	}

	token, err := client.RequestUserAccessToken(code)
	if err != nil {
		zap.L().Error("failed to retrieve user access token", zap.Error(err))
		http.Redirect(w, r, h.redirectURL, http.StatusFound)
		return
	}
	zap.L().Debug("token response", zap.Int("status", token.StatusCode), zap.String("error", token.ErrorMessage))

	// authorize for this call
	client.SetUserAccessToken(token.Data.AccessToken)

	ok, data, err := client.ValidateToken(token.Data.AccessToken)
	if err != nil {
		zap.L().Error("error occurred while validating Twitch OAuth token", zap.Error(err))
		http.Redirect(w, r, h.redirectURL, http.StatusFound)
		return
	} else if !ok {
		zap.L().Error("failed to validate client token", zap.Int("status", data.ErrorStatus), zap.String("error", data.ErrorMessage))
		http.Redirect(w, r, h.redirectURL, http.StatusFound)
		return
	}

	log.Printf("validated token for %v:%s\n", data.Data.UserID, data.Data.Login)

	// Check that the user is affiliated or partnered before letting them continue
	res, err := client.GetUsers(&helix.UsersParams{
		IDs: []string{data.Data.UserID},
	})
	if err != nil {
		zap.L().Error("failed to query user details", zap.String("id", data.Data.UserID), zap.Error(err))
	} else if len(res.Data.Users) == 1 {
		fetched := res.Data.Users[0]
		if fetched.BroadcasterType == "" {
			zap.L().Error("user is not affiliated or partnered", zap.String("id", data.Data.UserID))
			http.Redirect(w, r, h.redirectURL, http.StatusFound)
			return
		}
	}

	user := users.User{
		TwitchID:           data.Data.UserID,
		TwitchAccessToken:  token.Data.AccessToken,
		TwitchRefreshToken: token.Data.RefreshToken,
	}

	err = h.userStore.AddUser(&user)
	if err != nil {
		zap.L().Error("failed to store user auth details", zap.String("id", user.TwitchID), zap.Error(err))
		http.Redirect(w, r, h.redirectURL, http.StatusFound)
		return // don't set a cookie on the client
	}

	// store preference for the new user
	pref := preferences.Preference{
		TwitchID: data.Data.UserID,
	}
	err = h.prefStore.AddPreference(&pref)
	if err != nil {
		zap.L().Error("failed to store preference details for", zap.String("id", pref.TwitchID), zap.Error(err))
		http.Redirect(w, r, h.redirectURL, http.StatusFound)
		return // don't set a cookie on the client
	}

	// add a cookie in the response
	twitchCookie := http.Cookie{
		Name:     constants.TwitchIDCookieKey,
		Path:     "/",
		Value:    base64.StdEncoding.EncodeToString([]byte(user.TwitchID)),
		Secure:   true,
		SameSite: http.SameSiteLaxMode, // Lax so it can be passed around other domains
	}
	http.SetCookie(w, &twitchCookie)

	http.Redirect(w, r, h.redirectURL, http.StatusFound)
}
